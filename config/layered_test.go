package config

import (
	"os"
	"path/filepath"
	"testing"

	schemaPkg "github.com/fulmenhq/gofulmen/schema"
)

func sampleOptions() LayeredConfigOptions {
	defaultsRoot := "testdata"
	catalogRoot := filepath.Join("..", "schemas", "testdata")
	return LayeredConfigOptions{
		Category:     "sample",
		Version:      "v1.0.0",
		DefaultsFile: "sample-defaults.yaml",
		SchemaID:     "sample/v1.0.0/schema",
		DefaultsRoot: defaultsRoot,
		Catalog:      schemaPkg.NewCatalog(catalogRoot),
	}
}

func TestLoadLayeredConfig_DefaultsOnly(t *testing.T) {
	opts := sampleOptions()

	cfg, diags, err := LoadLayeredConfig(opts)
	if err != nil {
		t.Fatalf("LoadLayeredConfig returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected zero diagnostics, got %v", diags)
	}

	settings, ok := cfg["settings"].(map[string]any)
	if !ok {
		t.Fatalf("expected settings map, got %T", cfg["settings"])
	}
	if retries, ok := settings["retries"].(int); !ok || retries != 3 {
		t.Fatalf("expected default retries=3, got %v", settings["retries"])
	}
}

func TestLoadLayeredConfig_UserOverrides(t *testing.T) {
	userContent := `settings:
  retries: 5
  endpoints:
    - "https://api.override.fulmen.dev"
`
	tmpDir := t.TempDir()
	userFile := filepath.Join(tmpDir, "sample-overrides.yaml")
	if err := os.WriteFile(userFile, []byte(userContent), 0o600); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	opts := sampleOptions()
	opts.UserPaths = []string{userFile}

	cfg, diags, err := LoadLayeredConfig(opts)
	if err != nil {
		t.Fatalf("LoadLayeredConfig returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected zero diagnostics, got %v", diags)
	}

	settings := cfg["settings"].(map[string]any)
	if retries, ok := settings["retries"].(int); !ok || retries != 5 {
		t.Fatalf("expected user override retries=5, got %v", settings["retries"])
	}
	endpoints := settings["endpoints"].([]any)
	if len(endpoints) != 1 || endpoints[0] != "https://api.override.fulmen.dev" {
		t.Fatalf("expected endpoints override, got %v", endpoints)
	}
}

func TestLoadLayeredConfig_InvalidOverride(t *testing.T) {
	userContent := `settings:
  retries: -1
`
	tmpDir := t.TempDir()
	userFile := filepath.Join(tmpDir, "sample-invalid.yaml")
	if err := os.WriteFile(userFile, []byte(userContent), 0o600); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	opts := sampleOptions()
	opts.UserPaths = []string{userFile}

	_, diags, err := LoadLayeredConfig(opts)
	if err != nil {
		t.Fatalf("LoadLayeredConfig returned error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for invalid override")
	}
}
