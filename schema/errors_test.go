package schema

import (
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     ValidationError
		wantMsg string
	}{
		{
			name: "basic validation error",
			err: ValidationError{
				Field:   "name",
				Message: "is required",
				Value:   nil,
			},
			wantMsg: "validation error at name: is required",
		},
		{
			name: "validation error with value",
			err: ValidationError{
				Field:   "age",
				Message: "must be positive",
				Value:   -5,
			},
			wantMsg: "validation error at age: must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name    string
		errors  ValidationErrors
		wantMsg string
	}{
		{
			name:    "no errors",
			errors:  ValidationErrors{},
			wantMsg: "no validation errors",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "email", Message: "invalid format", Value: "bad-email"},
			},
			wantMsg: "validation errors:\nvalidation error at email: invalid format",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "name", Message: "is required", Value: nil},
				{Field: "age", Message: "must be positive", Value: -1},
			},
			wantMsg: "validation errors:\nvalidation error at name: is required\nvalidation error at age: must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.errors.Error()
			if got != tt.wantMsg {
				t.Errorf("ValidationErrors.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		message   string
		value     interface{}
		wantField string
		wantMsg   string
	}{
		{
			name:      "create validation error with string value",
			field:     "username",
			message:   "already exists",
			value:     "testuser",
			wantField: "username",
			wantMsg:   "already exists",
		},
		{
			name:      "create validation error with nil value",
			field:     "password",
			message:   "is required",
			value:     nil,
			wantField: "password",
			wantMsg:   "is required",
		},
		{
			name:      "create validation error with numeric value",
			field:     "count",
			message:   "exceeds maximum",
			value:     1000,
			wantField: "count",
			wantMsg:   "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message, tt.value)

			if err.Field != tt.wantField {
				t.Errorf("NewValidationError() Field = %q, want %q", err.Field, tt.wantField)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("NewValidationError() Message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.Value != tt.value {
				t.Errorf("NewValidationError() Value = %v, want %v", err.Value, tt.value)
			}

			// Verify error message format
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.wantField) {
				t.Errorf("Error message %q does not contain field %q", errMsg, tt.wantField)
			}
			if !strings.Contains(errMsg, tt.wantMsg) {
				t.Errorf("Error message %q does not contain message %q", errMsg, tt.wantMsg)
			}
		})
	}
}

func TestValidationErrors_Multiple(t *testing.T) {
	// Test building up multiple errors
	var errors ValidationErrors

	errors = append(errors, NewValidationError("field1", "error1", "value1"))
	errors = append(errors, NewValidationError("field2", "error2", nil))
	errors = append(errors, NewValidationError("field3", "error3", 42))

	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errors))
	}

	errMsg := errors.Error()

	// Verify all errors are included in the message
	expectedStrings := []string{"field1", "error1", "field2", "error2", "field3", "error3"}
	for _, expected := range expectedStrings {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("Error message does not contain %q: %s", expected, errMsg)
		}
	}
}
