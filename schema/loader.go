package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadJSONFile loads and parses a JSON file
func LoadJSONFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return data, nil
}

// LoadYAMLFile loads and parses a YAML file, converting to JSON
func LoadYAMLFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Convert YAML to JSON
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filename, err)
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON for %s: %w", filename, err)
	}

	return jsonData, nil
}

// LoadSchemaFile loads a schema from a file (supports both JSON and YAML)
func LoadSchemaFile(filename string) ([]byte, error) {
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return LoadYAMLFile(filename)
	}
	return LoadJSONFile(filename)
}

// LoadSchemaFromDir loads a schema by name from a directory
func LoadSchemaFromDir(dir, name string) ([]byte, error) {
	// Try YAML first, then JSON
	yamlFile := filepath.Join(dir, name+".yaml")
	if _, err := os.Stat(yamlFile); err == nil {
		return LoadSchemaFile(yamlFile)
	}

	jsonFile := filepath.Join(dir, name+".json")
	return LoadSchemaFile(jsonFile)
}

// ParseJSON parses JSON bytes into an interface{}
func ParseJSON(data []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(data, &result)
	return result, err
}
