package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSchemaValidateCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	data := `{"relativePath":"file.txt","sourcePath":"/tmp/file.txt","logicalPath":"file.txt","loaderType":"local","metadata":{}}`
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "path-result.json")
	if err := os.WriteFile(dataFile, []byte(data), 0o600); err != nil {
		t.Fatalf("write data file: %v", err)
	}

	cmd := exec.Command("go", "run", "./main.go", "schema", "validate", "--schema-id", "pathfinder/v1.0.0/path-result", dataFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("cli validate command failed: %v (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
}
