// Package validation provides structured error envelope wrappers for schema validation operations.
// This package exists to avoid import cycles between the schema and errors packages.
package validation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fulmenhq/gofulmen/errors"
	"github.com/fulmenhq/gofulmen/schema"
)

// ErrorEnvelope wraps schema validation operations with structured error envelopes
type ErrorEnvelope struct {
	validator *schema.Validator
}

// NewErrorEnvelope creates a new validation error envelope wrapper
func NewErrorEnvelope(validator *schema.Validator) *ErrorEnvelope {
	return &ErrorEnvelope{
		validator: validator,
	}
}

// ValidateDataWithEnvelope validates data and returns a structured error envelope on failure
func (v *ErrorEnvelope) ValidateDataWithEnvelope(data interface{}, correlationID string) (*errors.ErrorEnvelope, error) {
	diagnostics, err := v.validator.ValidateData(data)
	if err != nil {
		// Create error envelope for validation failure
		envelope := errors.NewErrorEnvelope("SCHEMA_VALIDATION_ERROR", "Schema validation failed due to internal error")
		envelope, severityErr := envelope.WithSeverity(errors.SeverityHigh)
		if severityErr != nil {
			// Log severity error but continue with envelope creation
			fmt.Printf("Warning: failed to set severity: %v\n", severityErr)
		}
		envelope = envelope.WithCorrelationID(correlationID)

		// Convert diagnostics to schema-compliant format (slice of strings)
		diagnosticStrings := make([]string, 0, len(diagnostics))
		for _, diag := range diagnostics {
			diagnosticStrings = append(diagnosticStrings, diag.Message)
		}

		context := map[string]interface{}{
			"component":   "schema",
			"operation":   "validate",
			"error_type":  "internal_validation_error",
			"diagnostics": diagnosticStrings,
		}

		envelope, contextErr := envelope.WithContext(context)
		if contextErr != nil {
			// If context validation fails, add the error to the envelope message
			envelope.Message = fmt.Sprintf("%s (context error: %v)", envelope.Message, contextErr)
		}

		envelope = envelope.WithOriginal(err)
		return envelope, err
	}

	if len(diagnostics) > 0 {
		// Create error envelope for validation failures
		envelope := errors.NewErrorEnvelope("SCHEMA_VALIDATION_FAILED", "Schema validation failed with validation errors")
		envelope, severityErr := envelope.WithSeverity(errors.SeverityMedium)
		if severityErr != nil {
			fmt.Printf("Warning: failed to set severity: %v\n", severityErr)
		}
		envelope = envelope.WithCorrelationID(correlationID)

		// Convert diagnostics to schema-compliant format (slice of strings)
		diagnosticStrings := make([]string, 0, len(diagnostics))
		for _, diag := range diagnostics {
			diagnosticStrings = append(diagnosticStrings, diag.Message)
		}

		context := map[string]interface{}{
			"component":   "schema",
			"operation":   "validate",
			"error_type":  "validation_errors",
			"diagnostics": diagnosticStrings,
		}

		envelope, contextErr := envelope.WithContext(context)
		if contextErr != nil {
			envelope.Message = fmt.Sprintf("%s (context error: %v)", envelope.Message, contextErr)
		}

		return envelope, fmt.Errorf("validation failed with %d diagnostic(s)", len(diagnostics))
	}

	return nil, nil // No errors, validation successful
}

// ValidateJSONWithEnvelope validates JSON and returns a structured error envelope on failure
func (v *ErrorEnvelope) ValidateJSONWithEnvelope(jsonData []byte, correlationID string) (*errors.ErrorEnvelope, error) {
	var payload interface{}
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		// Create error envelope for JSON parsing failure
		envelope := errors.NewErrorEnvelope("JSON_PARSE_ERROR", "Failed to parse JSON data")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":  "schema",
			"operation":  "validate_json",
			"error_type": "json_parse_error",
		})
		envelope = envelope.WithOriginal(err)
		return envelope, fmt.Errorf("invalid JSON: %w", err)
	}

	return v.ValidateDataWithEnvelope(payload, correlationID)
}

// ValidateFileWithEnvelope validates a file and returns a structured error envelope on failure
func (v *ErrorEnvelope) ValidateFileWithEnvelope(path string, correlationID string) (*errors.ErrorEnvelope, error) {
	// Use the underlying validator's file validation
	diagnostics, err := v.validator.ValidateFile(path)
	if err != nil {
		// Create error envelope for file access failure
		envelope := errors.NewErrorEnvelope("FILE_ACCESS_ERROR", "Failed to access validation file")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":  "schema",
			"operation":  "validate_file",
			"error_type": "file_access_error",
			"file_path":  path,
		})
		envelope = envelope.WithOriginal(err)
		return envelope, fmt.Errorf("failed to read file: %w", err)
	}

	if len(diagnostics) > 0 {
		// Create error envelope for validation failures
		envelope := errors.NewErrorEnvelope("SCHEMA_VALIDATION_FAILED", "Schema validation failed for file")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
		envelope = envelope.WithCorrelationID(correlationID)

		// Convert diagnostics to schema-compliant format (slice of strings)
		diagnosticStrings := make([]string, 0, len(diagnostics))
		for _, diag := range diagnostics {
			diagnosticStrings = append(diagnosticStrings, diag.Message)
		}

		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":   "schema",
			"operation":   "validate_file",
			"error_type":  "validation_errors",
			"file_path":   path,
			"diagnostics": diagnosticStrings,
		})

		return envelope, fmt.Errorf("validation failed with %d diagnostic(s)", len(diagnostics))
	}

	return nil, nil // No errors, validation successful
}

// LoadSchemaFileWithEnvelope loads a schema file and returns a structured error envelope on failure
func LoadSchemaFileWithEnvelope(filename string, correlationID string) ([]byte, *errors.ErrorEnvelope, error) {
	data, err := schema.LoadSchemaFile(filename)
	if err != nil {
		// Create error envelope for schema loading failure
		envelope := errors.NewErrorEnvelope("SCHEMA_LOAD_ERROR", "Failed to load schema file")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":  "schema",
			"operation":  "load_schema_file",
			"error_type": "schema_load_error",
			"file_path":  filename,
		})
		envelope = envelope.WithOriginal(err)
		return nil, envelope, err
	}
	return data, nil, nil
}

// NewSchemaCompilationError creates a structured error for schema compilation failures
func NewSchemaCompilationError(err error, schemaID string, correlationID string) *errors.ErrorEnvelope {
	envelope := errors.NewErrorEnvelope("SCHEMA_COMPILATION_ERROR", "Failed to compile JSON schema")
	envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
	envelope = envelope.WithCorrelationID(correlationID)
	envelope = errors.SafeWithContext(envelope, map[string]interface{}{
		"component":  "schema",
		"operation":  "compile_schema",
		"error_type": "schema_compilation_error",
		"schema_id":  schemaID,
	})
	envelope = envelope.WithOriginal(err)
	return envelope
}

// NewSchemaRegistryError creates a structured error for schema registry operations
func NewSchemaRegistryError(err error, operation string, correlationID string) *errors.ErrorEnvelope {
	envelope := errors.NewErrorEnvelope("SCHEMA_REGISTRY_ERROR", "Schema registry operation failed")
	envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
	envelope = envelope.WithCorrelationID(correlationID)
	envelope = errors.SafeWithContext(envelope, map[string]interface{}{
		"component":  "schema",
		"operation":  operation,
		"error_type": "schema_registry_error",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
	envelope = envelope.WithOriginal(err)
	return envelope
}
