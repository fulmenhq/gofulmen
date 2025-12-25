package appidentity

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEmbeddedAppIdentitySchemaMatchesCrucible(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current file path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))

	crucibleSchemaPath := filepath.Join(repoRoot, "schemas", "crucible-go", "config", "repository", "app-identity", "v1.0.0", "app-identity.schema.json")
	appidentitySchemaPath := filepath.Join(repoRoot, "appidentity", "app-identity.schema.json")

	crucibleSchema, err := os.ReadFile(crucibleSchemaPath)
	if err != nil {
		t.Fatalf("failed to read Crucible schema: %v", err)
	}
	appidentitySchema, err := os.ReadFile(appidentitySchemaPath)
	if err != nil {
		t.Fatalf("failed to read appidentity schema: %v", err)
	}

	if !bytes.Equal(crucibleSchema, appidentitySchema) {
		t.Fatalf("appidentity schema drift detected: run `make sync` to resync appidentity/app-identity.schema.json")
	}
}
