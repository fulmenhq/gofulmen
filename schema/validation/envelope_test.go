package validation

import (
	"testing"

	"github.com/fulmenhq/gofulmen/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorEnvelope(t *testing.T) {
	// Create a simple test schema
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`

	validator, err := schema.NewValidator([]byte(schemaJSON))
	require.NoError(t, err)

	envelope := NewErrorEnvelope(validator)

	tests := []struct {
		name          string
		data          interface{}
		correlationID string
		expectError   bool
		expectNil     bool
	}{
		{
			name:          "valid data",
			data:          map[string]interface{}{"name": "test", "age": 25},
			correlationID: "test-123",
			expectError:   false,
			expectNil:     true,
		},
		{
			name:          "missing required field",
			data:          map[string]interface{}{"age": 25},
			correlationID: "test-456",
			expectError:   true,
			expectNil:     false,
		},
		{
			name:          "invalid field type",
			data:          map[string]interface{}{"name": 123, "age": "invalid"},
			correlationID: "test-789",
			expectError:   true,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := envelope.ValidateDataWithEnvelope(tt.data, tt.correlationID)

			if tt.expectNil {
				assert.Nil(t, env)
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, env)
				assert.Error(t, err)
				// Basic checks on the envelope structure
				assert.NotEmpty(t, env.Code)
				assert.NotEmpty(t, env.Message)
				assert.NotEmpty(t, env.Timestamp)
				assert.Equal(t, tt.correlationID, env.CorrelationID)
				assert.NotNil(t, env.Context)
			}
		})
	}
}

func TestValidateJSONWithEnvelope(t *testing.T) {
	schemaJSON := `{"type": "string"}`
	validator, err := schema.NewValidator([]byte(schemaJSON))
	require.NoError(t, err)

	envelope := NewErrorEnvelope(validator)

	tests := []struct {
		name          string
		jsonData      string
		correlationID string
		expectError   bool
		expectNil     bool
	}{
		{
			name:          "valid JSON",
			jsonData:      `"test string"`,
			correlationID: "json-123",
			expectError:   false,
			expectNil:     true,
		},
		{
			name:          "invalid JSON syntax",
			jsonData:      `{"invalid": json}`,
			correlationID: "json-456",
			expectError:   true,
			expectNil:     false,
		},
		{
			name:          "valid JSON but invalid schema",
			jsonData:      `123`,
			correlationID: "json-789",
			expectError:   true,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := envelope.ValidateJSONWithEnvelope([]byte(tt.jsonData), tt.correlationID)

			if tt.expectNil {
				assert.Nil(t, env)
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, env)
				assert.Error(t, err)
				// Basic checks on the envelope structure
				assert.NotEmpty(t, env.Code)
				assert.NotEmpty(t, env.Message)
				assert.NotEmpty(t, env.Timestamp)
				assert.Equal(t, tt.correlationID, env.CorrelationID)
			}
		})
	}
}

func TestLoadSchemaFileWithEnvelope(t *testing.T) {
	// Test with a non-existent file
	data, env, err := LoadSchemaFileWithEnvelope("non-existent-file.json", "load-123")

	assert.Nil(t, data)
	assert.NotNil(t, env)
	assert.Error(t, err)

	// Basic checks on the envelope structure - it should be an ErrorEnvelope
	assert.NotNil(t, env)
	assert.NotEmpty(t, env.Code)
	assert.NotEmpty(t, env.Message)
	assert.NotEmpty(t, env.Timestamp)
	assert.Equal(t, "load-123", env.CorrelationID)
	assert.NotNil(t, env.Original)
}

func TestNewSchemaCompilationError(t *testing.T) {
	originalErr := assert.AnError
	env := NewSchemaCompilationError(originalErr, "test-schema", "compile-123")

	assert.Equal(t, "SCHEMA_COMPILATION_ERROR", env.Code)
	assert.NotEmpty(t, env.Message)
	assert.NotEmpty(t, env.Timestamp)
	assert.Equal(t, "compile-123", env.CorrelationID)
	assert.Equal(t, "schema", env.Context["component"])
	assert.Equal(t, "compile_schema", env.Context["operation"])
	assert.Equal(t, "schema_compilation_error", env.Context["error_type"])
	assert.Equal(t, "test-schema", env.Context["schema_id"])
	assert.Equal(t, originalErr.Error(), env.Original)
}

func TestNewSchemaRegistryError(t *testing.T) {
	originalErr := assert.AnError
	env := NewSchemaRegistryError(originalErr, "get_schema", "registry-123")

	assert.Equal(t, "SCHEMA_REGISTRY_ERROR", env.Code)
	assert.NotEmpty(t, env.Message)
	assert.NotEmpty(t, env.Timestamp)
	assert.Equal(t, "registry-123", env.CorrelationID)
	assert.Equal(t, "schema", env.Context["component"])
	assert.Equal(t, "get_schema", env.Context["operation"])
	assert.Equal(t, "schema_registry_error", env.Context["error_type"])
	assert.NotEmpty(t, env.Context["timestamp"])
	assert.Equal(t, originalErr.Error(), env.Original)
}
