package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorEnvelopeFixtures(t *testing.T) {
	// Load the error response schema for validation using the catalog
	catalog := schema.DefaultCatalog()
	validator, err := catalog.ValidatorByID("error-handling/v1.0.0/error-response")
	require.NoError(t, err, "Failed to load error response schema")

	// Find all fixture files
	fixturesDir := "../test/fixtures/errors"
	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir, entry.Name())

			// Read fixture
			data, err := os.ReadFile(fixturePath)
			require.NoError(t, err)

			// First validate against the JSON schema
			diagnostics, err := validator.ValidateJSON(data)
			require.NoError(t, err, "Fixture should validate against error response schema")
			assert.Empty(t, diagnostics, "Schema validation should pass without diagnostics")

			// Parse as ErrorEnvelope for additional Go-specific validation
			var envelope ErrorEnvelope
			err = json.Unmarshal(data, &envelope)
			require.NoError(t, err)

			// Validate required fields
			assert.NotEmpty(t, envelope.Code, "Code should not be empty")
			assert.NotEmpty(t, envelope.Message, "Message should not be empty")
			assert.NotEmpty(t, envelope.Timestamp, "Timestamp should not be empty")

			// Validate timestamp format
			_, err = time.Parse(time.RFC3339, envelope.Timestamp)
			assert.NoError(t, err, "Timestamp should be valid RFC3339")

			// Validate severity mapping if severity is set
			if envelope.Severity != "" {
				level, exists := SeverityLevel[envelope.Severity]
				assert.True(t, exists, "Severity should be valid")
				assert.Equal(t, level, envelope.SeverityLevel, "Severity level should match")
			}

			// Validate that it can be marshaled back to JSON
			marshaled, err := json.Marshal(envelope)
			require.NoError(t, err)
			assert.NotEmpty(t, marshaled)

			// Re-validate the marshaled JSON to ensure round-trip consistency
			diagnostics, err = validator.ValidateJSON(marshaled)
			require.NoError(t, err, "Marshaled envelope should still validate against schema")
			assert.Empty(t, diagnostics, "Round-trip validation should pass without diagnostics")
		})
	}
}
