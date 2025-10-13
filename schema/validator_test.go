package schema

import (
	"testing"
)

const testSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "age": {
      "type": "integer",
      "minimum": 0
    }
  },
  "required": ["name"]
}`

func TestNewValidator(t *testing.T) {
	validator, err := NewValidator([]byte(testSchema))
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	if validator == nil {
		t.Fatal("Validator is nil")
	}
}

func TestValidate(t *testing.T) {
	validator, err := NewValidator([]byte(testSchema))
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	validData := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	if err := validator.Validate(validData); err != nil {
		t.Errorf("Valid data should pass validation: %v", err)
	}

	invalidData := map[string]interface{}{
		"age": 30,
	}

	if err := validator.Validate(invalidData); err == nil {
		t.Error("Invalid data should fail validation")
	}
}

func TestValidateJSON(t *testing.T) {
	validator, err := NewValidator([]byte(testSchema))
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	validJSON := `{"name": "Jane", "age": 25}`
	if err := validator.ValidateJSON([]byte(validJSON)); err != nil {
		t.Errorf("Valid JSON should pass: %v", err)
	}

	invalidJSON := `{"age": 25}`
	if err := validator.ValidateJSON([]byte(invalidJSON)); err == nil {
		t.Error("Invalid JSON should fail")
	}
}
