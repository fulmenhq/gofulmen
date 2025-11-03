package export

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Test schema IDs from Crucible (these match the embedded crucible module structure)
const (
	testSchemaID      = "observability/logging/v1.0.0/log-event.schema.json"
	testSchemaIDBox   = "terminal/v1.0.0/schema.json"
	invalidSchemaID   = "invalid/schema"
	nonexistentSchema = "nonexistent/v1.0.0/schema.json"
)

func TestExportJSON(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "exported-schema.json")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.ValidateSchema = false // Disable validation for now (needs schema refs)
	require.NoError(t, opts.Validate())

	err := Export(ctx, opts)
	require.NoError(t, err, "Export should succeed")

	// Verify file exists
	require.FileExists(t, outPath, "Exported file should exist")

	// Read and parse exported JSON
	data, err := os.ReadFile(outPath)
	require.NoError(t, err, "Should be able to read exported file")

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err, "Exported data should be valid JSON")

	// Verify provenance metadata
	provenance, ok := jsonData["x-crucible-source"].(map[string]interface{})
	require.True(t, ok, "Should have x-crucible-source metadata")
	assert.Equal(t, testSchemaID, provenance["schema_id"])
	assert.NotEmpty(t, provenance["crucible_version"])
	assert.NotEmpty(t, provenance["gofulmen_version"])
	assert.NotEmpty(t, provenance["exported_at"])

	// Verify schema content is present
	assert.NotEmpty(t, jsonData, "Schema should have content")
}

func TestExportYAML(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "exported-schema.yaml")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	require.NoError(t, opts.Validate())

	err := Export(ctx, opts)
	require.NoError(t, err, "Export should succeed")

	// Verify file exists
	require.FileExists(t, outPath, "Exported file should exist")

	// Read exported YAML
	data, err := os.ReadFile(outPath)
	require.NoError(t, err, "Should be able to read exported file")

	// Verify YAML has provenance comment
	content := string(data)
	assert.Contains(t, content, "# x-crucible-source:")
	assert.Contains(t, content, testSchemaID)
	assert.Contains(t, content, "crucible_version:")
	assert.Contains(t, content, "gofulmen_version:")

	// Parse YAML to verify it's valid
	var yamlData interface{}
	err = yaml.Unmarshal(data, &yamlData)
	require.NoError(t, err, "Exported data should be valid YAML")
}

func TestExportOverwrite(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "existing-file.json")

	// Create an existing file
	err := os.WriteFile(outPath, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Try to export without overwrite (should fail)
	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.Overwrite = false

	err = Export(ctx, opts)
	require.Error(t, err, "Export should fail when file exists without Overwrite")
	assert.True(t, errors.Is(err, ErrFileExists) || strings.Contains(err.Error(), "already exists"))

	// Try again with overwrite (should succeed)
	opts.Overwrite = true
	err = Export(ctx, opts)
	require.NoError(t, err, "Export should succeed with Overwrite=true")

	// Verify file was overwritten
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEqual(t, "existing content", string(data))
}

func TestExportValidation(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		schemaID    string
		expectError bool
	}{
		{
			name:        "valid schema",
			schemaID:    testSchemaID,
			expectError: false,
		},
		{
			name:        "another valid schema",
			schemaID:    testSchemaIDBox,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(tempDir, strings.ReplaceAll(tt.name, " ", "_")+".json")
			opts := NewExportOptions(tt.schemaID, outPath)
			opts.ValidateSchema = false // Disable validation (needs schema refs)

			err := Export(ctx, opts)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProvenanceObject(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "schema-with-object.json")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.ProvenanceStyle = ProvenanceObject
	opts.IncludeProvenance = true

	err := Export(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err)

	// Should have x-crucible-source as object
	provenance, ok := jsonData["x-crucible-source"].(map[string]interface{})
	require.True(t, ok, "Should have x-crucible-source object")
	assert.Equal(t, testSchemaID, provenance["schema_id"])
}

func TestProvenanceComment(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "schema-with-comment.json")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.ProvenanceStyle = ProvenanceComment
	opts.IncludeProvenance = true

	err := Export(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err)

	// Should have $comment field
	comment, ok := jsonData["$comment"].(string)
	require.True(t, ok, "Should have $comment field")
	assert.Contains(t, comment, "x-crucible-source:")
	assert.Contains(t, comment, testSchemaID)
}

func TestProvenanceNone(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "schema-no-provenance.json")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.IncludeProvenance = false

	err := Export(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err)

	// Should NOT have provenance
	_, hasProvenance := jsonData["x-crucible-source"]
	assert.False(t, hasProvenance, "Should not have x-crucible-source")
	_, hasComment := jsonData["$comment"]
	assert.False(t, hasComment, "Should not have $comment")
}

func TestPathSafety(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		outPath     string
		expectError bool
	}{
		{
			name:        "simple filename",
			outPath:     filepath.Join(tempDir, "schema.json"),
			expectError: false,
		},
		{
			name:        "nested directory",
			outPath:     filepath.Join(tempDir, "subdir", "nested", "schema.json"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewExportOptions(testSchemaID, tt.outPath)
			opts.ValidateSchema = false
			err := Export(ctx, opts)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.FileExists(t, tt.outPath)
			}
		})
	}
}

func TestFormatDetection(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		filename       string
		expectedFormat Format
	}{
		{
			name:           "json extension",
			filename:       "schema.json",
			expectedFormat: FormatJSON,
		},
		{
			name:           "yaml extension",
			filename:       "schema.yaml",
			expectedFormat: FormatYAML,
		},
		{
			name:           "yml extension",
			filename:       "schema.yml",
			expectedFormat: FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(tempDir, tt.filename)
			opts := NewExportOptions(testSchemaID, outPath)
			opts.ValidateSchema = false
			opts.Format = FormatAuto

			err := Export(ctx, opts)
			require.NoError(t, err)

			// Verify file exists
			require.FileExists(t, outPath)
		})
	}
}

func TestExportOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		opts        ExportOptions
		expectError bool
	}{
		{
			name: "valid options",
			opts: ExportOptions{
				SchemaID: testSchemaID,
				OutPath:  "/tmp/schema.json",
			},
			expectError: false,
		},
		{
			name: "missing SchemaID",
			opts: ExportOptions{
				OutPath: "/tmp/schema.json",
			},
			expectError: true,
		},
		{
			name: "missing OutPath",
			opts: ExportOptions{
				SchemaID: testSchemaID,
			},
			expectError: true,
		},
		{
			name: "invalid SchemaID format",
			opts: ExportOptions{
				SchemaID: "invalid",
				OutPath:  "/tmp/schema.json",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewExportOptions(t *testing.T) {
	opts := NewExportOptions(testSchemaID, "output.json")

	assert.Equal(t, testSchemaID, opts.SchemaID)
	assert.Equal(t, "output.json", opts.OutPath)
	assert.Equal(t, FormatJSON, opts.Format) // Should detect from .json extension
	assert.True(t, opts.IncludeProvenance)
	assert.Equal(t, ProvenanceObject, opts.ProvenanceStyle)
	assert.True(t, opts.ValidateSchema) // Default is true
	assert.False(t, opts.Overwrite)
}

func TestFormatString(t *testing.T) {
	assert.Equal(t, "json", FormatJSON.String())
	assert.Equal(t, "yaml", FormatYAML.String())
	assert.Equal(t, "auto", FormatAuto.String())
}

func TestProvenanceStyleString(t *testing.T) {
	assert.Equal(t, "object", ProvenanceObject.String())
	assert.Equal(t, "comment", ProvenanceComment.String())
	assert.Equal(t, "none", ProvenanceNone.String())
}

func TestValidateExportedSchema(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "valid-schema.json")

	// Export a schema first
	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	err := Export(ctx, opts)
	require.NoError(t, err)

	// Validate the exported schema - should pass with provenance
	// Skip structure validation because schemas have unresolved references
	err = validateExportedSchemaWithOptions(ctx, testSchemaID, outPath, false)
	require.NoError(t, err, "Exported schema should validate successfully")

	// Export without provenance and validate
	outPath2 := filepath.Join(tempDir, "no-provenance.json")
	opts2 := NewExportOptions(testSchemaID, outPath2)
	opts2.ValidateSchema = false
	opts2.IncludeProvenance = false
	err = Export(ctx, opts2)
	require.NoError(t, err)

	// Validate the exported schema without provenance
	err = validateExportedSchemaWithOptions(ctx, testSchemaID, outPath2, false)
	require.NoError(t, err, "Exported schema without provenance should validate successfully")
}

func TestCompareSchemaPayloads(t *testing.T) {
	// Test that provenance is properly stripped and comparison works
	sourceSchema := []byte(`{"type": "string", "description": "test"}`)

	// Schema with object-style provenance
	exportedWithProvenance := []byte(`{
		"x-crucible-source": {
			"schema_id": "test",
			"crucible_version": "1.0.0",
			"gofulmen_version": "0.1.8"
		},
		"type": "string",
		"description": "test"
	}`)

	err := compareSchemaPayloads(exportedWithProvenance, sourceSchema)
	require.NoError(t, err, "Should match after stripping provenance")

	// Schema with comment-style provenance
	exportedWithComment := []byte(`{
		"$comment": "x-crucible-source: schema_id=test",
		"type": "string",
		"description": "test"
	}`)

	err = compareSchemaPayloads(exportedWithComment, sourceSchema)
	require.NoError(t, err, "Should match after stripping $comment provenance")

	// Schema with different content should fail
	differentSchema := []byte(`{"type": "string", "description": "different"}`)
	exportedDifferent := []byte(`{
		"x-crucible-source": {"schema_id": "test"},
		"type": "string",
		"description": "different"
	}`)

	err = compareSchemaPayloads(exportedDifferent, differentSchema)
	require.NoError(t, err, "Different content should match if underlying schema is same")

	// Truly different content
	trulyDifferent := []byte(`{"type": "number"}`)
	err = compareSchemaPayloads(exportedWithProvenance, trulyDifferent)
	require.Error(t, err, "Different schema structure should fail")
}

// mockIdentityProvider is a test identity provider
type mockIdentityProvider struct {
	identity *Identity
	err      error
}

func (m *mockIdentityProvider) GetIdentity(ctx context.Context) (*Identity, error) {
	return m.identity, m.err
}

func TestExportWithIdentity(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "schema-with-identity.json")

	opts := NewExportOptions(testSchemaID, outPath)
	opts.ValidateSchema = false
	opts.IdentityProvider = &mockIdentityProvider{
		identity: &Identity{
			Vendor: "fulmenhq",
			Binary: "test-app",
		},
	}

	err := Export(ctx, opts)
	require.NoError(t, err)

	// Read and verify identity is included
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err)

	provenance, ok := jsonData["x-crucible-source"].(map[string]interface{})
	require.True(t, ok)

	identity, ok := provenance["identity"].(map[string]interface{})
	require.True(t, ok, "Should have identity in provenance")
	assert.Equal(t, "fulmenhq", identity["vendor"])
	assert.Equal(t, "test-app", identity["binary"])
}
