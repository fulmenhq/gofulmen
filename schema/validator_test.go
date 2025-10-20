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

	if diags, err := validator.ValidateData(validData); err != nil || len(diags) > 0 {
		t.Fatalf("Valid data should pass validation, err=%v diagnostics=%v", err, diags)
	}

	invalidData := map[string]interface{}{
		"age": 30,
	}

	if diags, err := validator.ValidateData(invalidData); err != nil {
		t.Fatalf("unexpected error validating data: %v", err)
	} else if len(diags) == 0 {
		t.Error("expected diagnostics for invalid data")
	}
}

func TestValidateJSON(t *testing.T) {
	validator, err := NewValidator([]byte(testSchema))
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	validJSON := `{"name": "Jane", "age": 25}`
	if diags, err := validator.ValidateJSON([]byte(validJSON)); err != nil || len(diags) > 0 {
		t.Fatalf("Valid JSON should pass: err=%v diagnostics=%v", err, diags)
	}

	invalidJSON := `{"age": 25}`
	if diags, err := validator.ValidateJSON([]byte(invalidJSON)); err != nil {
		t.Fatalf("unexpected error validating json: %v", err)
	} else if len(diags) == 0 {
		t.Error("Invalid JSON should produce diagnostics")
	}
}
