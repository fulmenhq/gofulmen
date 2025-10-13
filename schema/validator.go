package schema

import (
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Validator wraps the JSON Schema validator
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator creates a new validator from a schema
func NewValidator(schemaData []byte) (*Validator, error) {
	schema, err := jsonschema.CompileString("schema.json", string(schemaData))
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator{schema: schema}, nil
}

// Validate validates data against the schema
func (v *Validator) Validate(data interface{}) error {
	return v.schema.Validate(data)
}

// ValidateJSON validates JSON bytes against the schema
func (v *Validator) ValidateJSON(jsonData []byte) error {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return v.Validate(data)
}

// ValidateFile validates a JSON file against the schema
func (v *Validator) ValidateFile(filename string) error {
	data, err := LoadJSONFile(filename)
	if err != nil {
		return err
	}
	return v.ValidateJSON(data)
}
