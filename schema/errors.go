package schema

import (
	"fmt"
	"strings"
)

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error at %s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("validation errors:\n%s", strings.Join(msgs, "\n"))
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}
