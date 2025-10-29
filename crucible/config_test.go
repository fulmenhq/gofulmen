package crucible

import (
	"testing"
)

// TestConfigEmbeddingFromCrucible verifies that all Foundry config files
// are accessible from Crucible's embedded config (v0.2.1+).
// This test will FAIL if Crucible doesn't properly embed config/.
func TestConfigEmbeddingFromCrucible(t *testing.T) {
	tests := []struct {
		name     string
		loadFunc func() ([]byte, error)
		minBytes int // Minimum expected size to verify non-empty
	}{
		{
			name:     "Patterns",
			loadFunc: func() ([]byte, error) { return ConfigRegistry.Library().Foundry().Patterns() },
			minBytes: 1000, // patterns.yaml is ~6KB
		},
		{
			name:     "CountryCodes",
			loadFunc: func() ([]byte, error) { return ConfigRegistry.Library().Foundry().CountryCodes() },
			minBytes: 500, // country-codes.yaml is ~700 bytes
		},
		{
			name:     "HTTPStatuses",
			loadFunc: func() ([]byte, error) { return ConfigRegistry.Library().Foundry().HTTPStatuses() },
			minBytes: 2000, // http-statuses.yaml is ~4KB
		},
		{
			name:     "MIMETypes",
			loadFunc: func() ([]byte, error) { return ConfigRegistry.Library().Foundry().MIMETypes() },
			minBytes: 500, // mime-types.yaml is ~1KB
		},
		{
			name:     "SimilarityFixtures",
			loadFunc: func() ([]byte, error) { return ConfigRegistry.Library().Foundry().SimilarityFixtures() },
			minBytes: 5000, // similarity-fixtures.yaml is ~12KB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.loadFunc()
			if err != nil {
				t.Fatalf("Failed to load %s: %v\nThis likely means Crucible v0.2.1+ doesn't embed config/", tt.name, err)
			}

			if len(data) < tt.minBytes {
				t.Errorf("Config %s is suspiciously small: got %d bytes, expected at least %d bytes", tt.name, len(data), tt.minBytes)
			}

			t.Logf("✅ Loaded %s: %d bytes", tt.name, len(data))
		})
	}
}

// TestConfigGetConfig verifies the generic GetConfig function works
func TestConfigGetConfig(t *testing.T) {
	data, err := GetConfig("library/foundry/patterns.yaml")
	if err != nil {
		t.Fatalf("Failed to load via GetConfig: %v", err)
	}

	if len(data) == 0 {
		t.Error("GetConfig returned empty data")
	}

	t.Logf("✅ GetConfig loaded patterns.yaml: %d bytes", len(data))
}

// TestConfigListConfigs verifies the ListConfigs function works
func TestConfigListConfigs(t *testing.T) {
	files, err := ListConfigs("library/foundry")
	if err != nil {
		t.Fatalf("Failed to list configs: %v", err)
	}

	expectedFiles := map[string]bool{
		"patterns.yaml":            false,
		"country-codes.yaml":       false,
		"http-statuses.yaml":       false,
		"mime-types.yaml":          false,
		"similarity-fixtures.yaml": false,
	}

	for _, file := range files {
		if _, ok := expectedFiles[file]; ok {
			expectedFiles[file] = true
		}
	}

	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected config file not found: %s", file)
		}
	}

	t.Logf("✅ ListConfigs found %d files in library/foundry", len(files))
}
