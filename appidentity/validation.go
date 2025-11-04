package appidentity

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

//go:embed app-identity.schema.json
var embeddedSchema []byte

var (
	// validator is the compiled schema validator (lazily initialized)
	validator     *schema.Validator
	validatorErr  error
	validatorOnce sync.Once
)

// getValidator returns the compiled schema validator, initializing it on first use.
func getValidator() (*schema.Validator, error) {
	validatorOnce.Do(func() {
		validator, validatorErr = schema.NewValidator(embeddedSchema)
	})
	return validator, validatorErr
}

// Validate validates an identity file at the given path against the Crucible schema.
//
// This function loads the file, parses it, and validates it against the embedded
// app-identity schema. It returns detailed field-level diagnostics for any
// validation failures.
//
// Example:
//
//	if err := appidentity.Validate(ctx, "/path/to/.fulmen/app.yaml"); err != nil {
//	    var valErr *appidentity.ValidationError
//	    if errors.As(err, &valErr) {
//	        for _, fieldErr := range valErr.Errors {
//	            fmt.Printf("  %s: %s\n", fieldErr.Field, fieldErr.Message)
//	        }
//	    }
//	}
func Validate(ctx context.Context, path string) error {
	// Load file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read identity file: %w", err)
	}

	// Parse YAML to generic structure
	var payload interface{}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return &MalformedError{
			Path: path,
			Err:  err,
		}
	}

	// Get validator
	v, err := getValidator()
	if err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	// Validate
	diagnostics, err := v.ValidateData(payload)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(diagnostics) == 0 {
		return nil
	}

	// Convert diagnostics to ValidationError
	return diagnosticsToValidationError(path, diagnostics)
}

// ValidateIdentity validates an Identity struct directly.
//
// This method converts the identity to JSON and validates it against the schema.
// It's useful for validating programmatically-constructed identities.
//
// Example:
//
//	identity := &appidentity.Identity{
//	    BinaryName: "myapp",
//	    Vendor:     "myvendor",
//	    // ...
//	}
//	if err := appidentity.ValidateIdentity(ctx, identity); err != nil {
//	    return fmt.Errorf("invalid identity: %w", err)
//	}
func ValidateIdentity(ctx context.Context, identity *Identity) error {
	// Wrap in identityFile structure to match schema (app + metadata as siblings)
	// The Identity struct has json:"-" on Metadata field to prevent nesting
	file := identityFile{
		App:      *identity,
		Metadata: identity.Metadata,
	}

	// Convert to JSON for validation (schema validator expects JSON-compatible types)
	data, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}

	// Get validator
	v, err := getValidator()
	if err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	// Validate
	diagnostics, err := v.ValidateJSON(data)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(diagnostics) == 0 {
		return nil
	}

	// Convert diagnostics to ValidationError
	return diagnosticsToValidationError("<identity>", diagnostics)
}

// diagnosticsToValidationError converts schema diagnostics to ValidationError.
func diagnosticsToValidationError(path string, diagnostics []schema.Diagnostic) error {
	fieldErrors := make([]FieldError, 0, len(diagnostics))

	for _, diag := range diagnostics {
		if diag.Severity != schema.SeverityError {
			continue
		}

		field := diag.Pointer
		if field == "" {
			field = diag.Keyword
		}

		fieldErrors = append(fieldErrors, FieldError{
			Field:   field,
			Message: diag.Message,
		})
	}

	if len(fieldErrors) == 0 {
		return nil
	}

	return &ValidationError{
		Path:   path,
		Errors: fieldErrors,
	}
}
