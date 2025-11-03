package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLISuccess(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "exported.json")

	// Build the command
	cmd := exec.Command("go", "run", ".",
		"--schema-id=terminal/v1.0.0/schema.json",
		"--out="+outPath,
		"--no-validate")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "CLI should succeed: %s", string(output))

	// Verify output file exists
	require.FileExists(t, outPath)

	// Verify it's valid JSON with provenance
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err, "Exported file should be valid JSON")

	// Check provenance
	provenance, ok := jsonData["x-crucible-source"].(map[string]interface{})
	require.True(t, ok, "Should have provenance metadata")
	assert.Equal(t, "terminal/v1.0.0/schema.json", provenance["schema_id"])
}

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "gofulmen-export-schema")
	assert.Contains(t, outputStr, "--schema-id")
	assert.Contains(t, outputStr, "--out")
}

func TestCLIMissingRequiredArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing schema-id",
			args: []string{"--out=/tmp/test.json"},
		},
		{
			name: "missing out",
			args: []string{"--schema-id=terminal/v1.0.0/schema.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"run", "."}, tt.args...)
			cmd := exec.Command("go", args...)
			err := cmd.Run()
			require.Error(t, err, "CLI should fail with missing required args")
		})
	}
}

func TestCLIOverwriteProtection(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "existing.json")

	// Create existing file
	err := os.WriteFile(outPath, []byte("existing"), 0644)
	require.NoError(t, err)

	// Try to export without --force (should fail)
	cmd := exec.Command("go", "run", ".",
		"--schema-id=terminal/v1.0.0/schema.json",
		"--out="+outPath,
		"--no-validate")

	err = cmd.Run()
	require.Error(t, err, "CLI should fail when file exists without --force")

	// Try again with --force (should succeed)
	cmd = exec.Command("go", "run", ".",
		"--schema-id=terminal/v1.0.0/schema.json",
		"--out="+outPath,
		"--no-validate",
		"--force")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "CLI should succeed with --force: %s", string(output))
}

func TestCLIYAMLFormat(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "exported.yaml")

	cmd := exec.Command("go", "run", ".",
		"--schema-id=terminal/v1.0.0/schema.json",
		"--out="+outPath,
		"--format=yaml",
		"--no-validate")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "CLI should succeed: %s", string(output))

	// Verify output file exists
	require.FileExists(t, outPath)

	// Verify it has YAML front-matter provenance
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "# x-crucible-source:")
	assert.Contains(t, content, "terminal/v1.0.0/schema.json")
}

func TestCLIProvenanceStyles(t *testing.T) {
	tests := []struct {
		name  string
		style string
	}{
		{
			name:  "object style",
			style: "object",
		},
		{
			name:  "comment style",
			style: "comment",
		},
		{
			name:  "none style",
			style: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outPath := filepath.Join(tempDir, "exported.json")

			cmd := exec.Command("go", "run", ".",
				"--schema-id=terminal/v1.0.0/schema.json",
				"--out="+outPath,
				"--provenance-style="+tt.style,
				"--no-validate")

			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "CLI should succeed: %s", string(output))

			require.FileExists(t, outPath)
		})
	}
}
