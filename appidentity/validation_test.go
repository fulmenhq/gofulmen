package appidentity

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// TestValidateValidMinimal verifies valid-minimal.yaml passes validation.
func TestValidateValidMinimal(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "valid-minimal.yaml")

	err := Validate(ctx, fixturePath)
	if err != nil {
		t.Fatalf("expected valid fixture to pass validation, got: %v", err)
	}
}

// TestValidateValidComplete verifies valid-complete.yaml passes validation.
func TestValidateValidComplete(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "valid-complete.yaml")

	err := Validate(ctx, fixturePath)
	if err != nil {
		t.Fatalf("expected valid fixture to pass validation, got: %v", err)
	}
}

// TestValidateValidGofulmen verifies valid-gofulmen.yaml passes validation.
func TestValidateValidGofulmen(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "valid-gofulmen.yaml")

	err := Validate(ctx, fixturePath)
	if err != nil {
		t.Fatalf("expected valid fixture to pass validation, got: %v", err)
	}
}

// TestValidateInvalidMissingField verifies invalid-missing-field.yaml fails validation.
func TestValidateInvalidMissingField(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "invalid-missing-field.yaml")

	err := Validate(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(valErr.Errors) == 0 {
		t.Error("ValidationError should contain field errors")
	}

	// Should mention "description" as the missing field
	found := false
	for _, fieldErr := range valErr.Errors {
		t.Logf("Field error: %s - %s", fieldErr.Field, fieldErr.Message)
		if fieldErr.Field == "/app" && (fieldErr.Message == "missing properties: 'description'" ||
			fieldErr.Message == "missing property 'description'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation error to mention missing 'description' field")
	}
}

// TestValidateInvalidEnvPrefix verifies invalid-env-prefix.yaml fails validation.
func TestValidateInvalidEnvPrefix(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "invalid-env-prefix.yaml")

	err := Validate(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected validation error for invalid env_prefix")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(valErr.Errors) == 0 {
		t.Error("ValidationError should contain field errors")
	}

	// Should mention env_prefix pattern violation
	found := false
	for _, fieldErr := range valErr.Errors {
		t.Logf("Field error: %s - %s", fieldErr.Field, fieldErr.Message)
		if fieldErr.Field == "/app/env_prefix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation error for env_prefix field")
	}
}

// TestValidateIdentityStruct verifies ValidateIdentity with valid struct.
func TestValidateIdentityStruct(t *testing.T) {
	ctx := context.Background()

	identity := &Identity{
		BinaryName:  "testapp",
		Vendor:      "testvendor",
		EnvPrefix:   "TESTAPP_",
		ConfigName:  "testapp",
		Description: "Test application for validation",
	}

	err := ValidateIdentity(ctx, identity)
	if err != nil {
		t.Fatalf("expected valid identity to pass validation, got: %v", err)
	}
}

// TestValidateIdentityInvalid verifies ValidateIdentity with invalid struct.
func TestValidateIdentityInvalid(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		identity *Identity
		field    string
	}{
		{
			name: "missing description",
			identity: &Identity{
				BinaryName: "testapp",
				Vendor:     "testvendor",
				EnvPrefix:  "TESTAPP_",
				ConfigName: "testapp",
				// Description missing
			},
			field: "/app",
		},
		{
			name: "invalid env_prefix",
			identity: &Identity{
				BinaryName:  "testapp",
				Vendor:      "testvendor",
				EnvPrefix:   "TESTAPP", // Missing trailing underscore
				ConfigName:  "testapp",
				Description: "Test application",
			},
			field: "/app/env_prefix",
		},
		{
			name: "invalid binary_name",
			identity: &Identity{
				BinaryName:  "TestApp", // Uppercase not allowed
				Vendor:      "testvendor",
				EnvPrefix:   "TESTAPP_",
				ConfigName:  "testapp",
				Description: "Test application",
			},
			field: "/app/binary_name",
		},
		{
			name: "description too short",
			identity: &Identity{
				BinaryName:  "testapp",
				Vendor:      "testvendor",
				EnvPrefix:   "TESTAPP_",
				ConfigName:  "testapp",
				Description: "Short", // Less than 10 chars
			},
			field: "/app/description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentity(ctx, tt.identity)
			if err == nil {
				t.Fatal("expected validation error")
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError, got %T", err)
			}

			if len(valErr.Errors) == 0 {
				t.Error("ValidationError should contain field errors")
			}

			// Check that expected field is mentioned
			found := false
			for _, fieldErr := range valErr.Errors {
				t.Logf("Field error: %s - %s", fieldErr.Field, fieldErr.Message)
				if fieldErr.Field == tt.field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected validation error for field %s", tt.field)
			}
		})
	}
}

// TestValidationErrorDetails verifies ValidationError provides detailed diagnostics.
func TestValidationErrorDetails(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "invalid-missing-field.yaml")

	err := Validate(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	// Verify error message includes path
	errMsg := valErr.Error()
	if errMsg == "" {
		t.Error("ValidationError.Error() should return non-empty message")
	}

	if valErr.Path != fixturePath {
		t.Errorf("ValidationError.Path = %q, want %q", valErr.Path, fixturePath)
	}

	// Verify we can unwrap to ErrInvalid
	if !errors.Is(err, ErrInvalid) {
		t.Error("ValidationError should unwrap to ErrInvalid")
	}
}

// TestValidateMalformedFile verifies malformed YAML returns MalformedError.
func TestValidateMalformedFile(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "invalid-format.yaml")

	err := Validate(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}

	var malformedErr *MalformedError
	if !errors.As(err, &malformedErr) {
		t.Fatalf("expected MalformedError, got %T", err)
	}

	if !errors.Is(err, ErrMalformed) {
		t.Error("MalformedError should unwrap to ErrMalformed")
	}
}

// TestValidateNotFound verifies missing file returns appropriate error.
func TestValidateNotFound(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "nonexistent.yaml")

	err := Validate(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	// Should be a regular file read error (not ValidationError)
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		t.Error("should not return ValidationError for missing file")
	}
}
