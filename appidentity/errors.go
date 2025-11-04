package appidentity

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for identity operations.
var (
	// ErrNotFound is returned when .fulmen/app.yaml cannot be found.
	ErrNotFound = errors.New("app identity not found")

	// ErrInvalid is returned when identity fails schema validation.
	ErrInvalid = errors.New("app identity validation failed")

	// ErrMalformed is returned when YAML cannot be parsed.
	ErrMalformed = errors.New("app identity file malformed")
)

// NotFoundError provides detailed information about identity file discovery failure.
//
// This error includes the list of paths searched during ancestor discovery,
// helping users diagnose why their identity file wasn't found.
type NotFoundError struct {
	// SearchedPaths contains all paths checked during discovery.
	SearchedPaths []string

	// StartDir is the directory where the search began.
	StartDir string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	var sb strings.Builder
	sb.WriteString("app identity not found")

	if e.StartDir != "" {
		sb.WriteString(fmt.Sprintf(" (started from: %s)", e.StartDir))
	}

	if len(e.SearchedPaths) > 0 {
		sb.WriteString("\nSearched paths:")
		for _, path := range e.SearchedPaths {
			sb.WriteString(fmt.Sprintf("\n  - %s", path))
		}
	}

	sb.WriteString("\n\nTo resolve:")
	sb.WriteString("\n  1. Create .fulmen/app.yaml in your project root")
	sb.WriteString("\n  2. Set FULMEN_APP_IDENTITY_PATH environment variable")
	sb.WriteString("\n  3. Use LoadFrom() with explicit path")
	sb.WriteString("\n\nSee docs/appidentity/README.md for guidance on generating and managing identity files.")

	return sb.String()
}

// Unwrap returns the underlying ErrNotFound sentinel.
func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// ValidationError provides field-level diagnostics for schema validation failures.
//
// This error includes specific field errors with constraint violations,
// making it easier to fix invalid identity files.
type ValidationError struct {
	// Path is the file path that failed validation.
	Path string

	// Errors contains field-level validation errors.
	Errors []FieldError
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("app identity validation failed")

	if e.Path != "" {
		sb.WriteString(fmt.Sprintf(" (%s)", e.Path))
	}

	if len(e.Errors) > 0 {
		sb.WriteString("\nValidation errors:")
		for _, fieldErr := range e.Errors {
			sb.WriteString(fmt.Sprintf("\n  - %s: %s", fieldErr.Field, fieldErr.Message))
		}
	}

	return sb.String()
}

// Unwrap returns the underlying ErrInvalid sentinel.
func (e *ValidationError) Unwrap() error {
	return ErrInvalid
}

// FieldError describes a specific validation failure for a field.
type FieldError struct {
	// Field is the JSON path to the field (e.g., "app.binary_name").
	Field string

	// Message describes the validation constraint that failed.
	Message string

	// Value is the actual value that failed validation (optional).
	Value any
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%s: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// MalformedError wraps YAML parsing errors with additional context.
type MalformedError struct {
	// Path is the file path that failed to parse.
	Path string

	// Line is the line number where the error occurred (if available).
	Line int

	// Err is the underlying parsing error.
	Err error
}

// Error implements the error interface.
func (e *MalformedError) Error() string {
	var sb strings.Builder
	sb.WriteString("app identity file malformed")

	if e.Path != "" {
		if e.Line > 0 {
			sb.WriteString(fmt.Sprintf(" (%s:%d)", e.Path, e.Line))
		} else {
			sb.WriteString(fmt.Sprintf(" (%s)", e.Path))
		}
	}

	if e.Err != nil {
		sb.WriteString(fmt.Sprintf(": %s", e.Err.Error()))
	}

	return sb.String()
}

// Unwrap returns the sentinel error for error chain matching.
func (e *MalformedError) Unwrap() error {
	return ErrMalformed
}
