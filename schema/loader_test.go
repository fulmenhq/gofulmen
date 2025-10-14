package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONFile(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "valid JSON file",
			file:    "testdata/valid.json",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    "testdata/nonexistent.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadJSONFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadJSONFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("LoadJSONFile() returned empty data for valid file")
			}
		})
	}
}

func TestLoadYAMLFile(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "valid YAML file",
			file:    "testdata/valid.yaml",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    "testdata/nonexistent.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadYAMLFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadYAMLFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(data) == 0 {
					t.Error("LoadYAMLFile() returned empty data for valid file")
				}
				// Verify it's valid JSON
				var result interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("LoadYAMLFile() did not convert to valid JSON: %v", err)
				}
			}
		})
	}
}

func TestLoadSchemaFile(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "JSON file with .json extension",
			file:    "testdata/valid.json",
			wantErr: false,
		},
		{
			name:    "YAML file with .yaml extension",
			file:    "testdata/valid.yaml",
			wantErr: false,
		},
		{
			name:    "YAML file with .yml extension",
			file:    "testdata/config-schema.yaml",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    "testdata/missing.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadSchemaFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSchemaFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("LoadSchemaFile() returned empty data for valid file")
			}
		})
	}
}

func TestLoadSchemaFromDir(t *testing.T) {
	tests := []struct {
		name       string
		dir        string
		schemaName string
		wantErr    bool
	}{
		{
			name:       "load person-schema JSON from testdata",
			dir:        "testdata",
			schemaName: "person-schema",
			wantErr:    false,
		},
		{
			name:       "load config-schema YAML from testdata",
			dir:        "testdata",
			schemaName: "config-schema",
			wantErr:    false,
		},
		{
			name:       "non-existent schema",
			dir:        "testdata",
			schemaName: "nonexistent-schema",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadSchemaFromDir(tt.dir, tt.schemaName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSchemaFromDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(data) == 0 {
					t.Error("LoadSchemaFromDir() returned empty data for valid schema")
				}
				// Verify it's valid JSON
				var result interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("LoadSchemaFromDir() did not return valid JSON: %v", err)
				}
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid JSON object",
			data:    []byte(`{"name": "test", "value": 123}`),
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			data:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			data:    []byte(`"simple string"`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{invalid json`),
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte(``),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseJSON() returned nil for valid JSON")
			}
		})
	}
}

func TestLoadSchemaFileExtensionHandling(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Test .yml extension (should use YAML loader)
	ymlPath := filepath.Join(tmpDir, "test.yml")
	yamlContent := []byte("type: object\nproperties:\n  name:\n    type: string")
	if err := WriteTestFile(ymlPath, yamlContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data, err := LoadSchemaFile(ymlPath)
	if err != nil {
		t.Errorf("LoadSchemaFile() failed for .yml file: %v", err)
	}
	if len(data) == 0 {
		t.Error("LoadSchemaFile() returned empty data for .yml file")
	}

	// Verify conversion to JSON
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("LoadSchemaFile() did not convert .yml to valid JSON: %v", err)
	}
}

// WriteTestFile is a helper for creating test files
func WriteTestFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
