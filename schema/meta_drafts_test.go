package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSchemaBytes_SupportsAllCrucibleMetaDrafts(t *testing.T) {
	paths := []string{
		filepath.Join("..", "schemas", "crucible-go", "meta", "fixtures", "draft-04-sample.json"),
		filepath.Join("..", "schemas", "crucible-go", "meta", "fixtures", "draft-06-sample.json"),
		filepath.Join("..", "schemas", "crucible-go", "meta", "fixtures", "draft-07-sample.json"),
		filepath.Join("..", "schemas", "crucible-go", "meta", "fixtures", "draft-2019-09-sample.json"),
		filepath.Join("..", "schemas", "crucible-go", "meta", "fixtures", "draft-2020-12-sample.json"),
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read fixture %s: %v", path, err)
			}
			diags, err := ValidateSchemaBytes(data)
			if err != nil {
				t.Fatalf("ValidateSchemaBytes error for %s: %v", path, err)
			}
			if len(diags) != 0 {
				t.Fatalf("expected no diagnostics for %s, got %v", path, diags)
			}

			if _, err := NewValidator(data); err != nil {
				t.Fatalf("NewValidator failed to compile %s: %v", path, err)
			}
		})
	}
}
