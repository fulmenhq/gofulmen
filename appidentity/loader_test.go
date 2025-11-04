package appidentity

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadFrom verifies loading identity from an explicit path.
func TestLoadFrom(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		fixture     string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid minimal",
			fixture:     "valid-minimal.yaml",
			expectError: false,
		},
		{
			name:        "valid complete",
			fixture:     "valid-complete.yaml",
			expectError: false,
		},
		{
			name:        "valid gofulmen",
			fixture:     "valid-gofulmen.yaml",
			expectError: false,
		},
		{
			name:        "not found",
			fixture:     "nonexistent.yaml",
			expectError: true,
			errorType:   ErrNotFound,
		},
		{
			name:        "malformed YAML",
			fixture:     "invalid-format.yaml",
			expectError: true,
			errorType:   ErrMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("testdata", tt.fixture)

			identity, err := LoadFrom(ctx, fixturePath)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("expected error to wrap %v, got %v", tt.errorType, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if identity == nil {
				t.Fatal("identity should not be nil")
			}

			// Verify basic fields are populated
			if identity.BinaryName == "" {
				t.Error("BinaryName should not be empty")
			}
			if identity.Vendor == "" {
				t.Error("Vendor should not be empty")
			}
			if identity.EnvPrefix == "" {
				t.Error("EnvPrefix should not be empty")
			}
		})
	}
}

// TestLoadFromValidMinimal verifies minimal fixture content.
func TestLoadFromValidMinimal(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "valid-minimal.yaml")

	identity, err := LoadFrom(ctx, fixturePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]string{
		"BinaryName": "testapp",
		"Vendor":     "testvendor",
		"EnvPrefix":  "TESTAPP_",
		"ConfigName": "testapp",
	}

	if identity.BinaryName != expected["BinaryName"] {
		t.Errorf("BinaryName = %q, want %q", identity.BinaryName, expected["BinaryName"])
	}
	if identity.Vendor != expected["Vendor"] {
		t.Errorf("Vendor = %q, want %q", identity.Vendor, expected["Vendor"])
	}
	if identity.EnvPrefix != expected["EnvPrefix"] {
		t.Errorf("EnvPrefix = %q, want %q", identity.EnvPrefix, expected["EnvPrefix"])
	}
	if identity.ConfigName != expected["ConfigName"] {
		t.Errorf("ConfigName = %q, want %q", identity.ConfigName, expected["ConfigName"])
	}
}

// TestLoadFromValidComplete verifies complete fixture with metadata.
func TestLoadFromValidComplete(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "valid-complete.yaml")

	identity, err := LoadFrom(ctx, fixturePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify core fields
	if identity.BinaryName != "myapp" {
		t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "myapp")
	}

	// Verify metadata fields
	if identity.Metadata.ProjectURL != "https://github.com/example/myapp" {
		t.Errorf("ProjectURL = %q, want %q",
			identity.Metadata.ProjectURL, "https://github.com/example/myapp")
	}

	if identity.Metadata.License != "MIT" {
		t.Errorf("License = %q, want %q", identity.Metadata.License, "MIT")
	}

	if identity.Metadata.RepositoryCategory != "cli" {
		t.Errorf("RepositoryCategory = %q, want %q",
			identity.Metadata.RepositoryCategory, "cli")
	}

	// Verify Python metadata
	if identity.Metadata.Python == nil {
		t.Fatal("Python metadata should not be nil")
	}

	if identity.Metadata.Python.DistributionName != "my-app" {
		t.Errorf("DistributionName = %q, want %q",
			identity.Metadata.Python.DistributionName, "my-app")
	}

	if len(identity.Metadata.Python.ConsoleScripts) != 2 {
		t.Errorf("ConsoleScripts length = %d, want %d",
			len(identity.Metadata.Python.ConsoleScripts), 2)
	}

	// Verify extras/custom fields (should capture unknown metadata fields via inline tag)
	if identity.Metadata.Extras == nil {
		t.Fatal("Extras should not be nil")
	}

	if identity.Metadata.Extras["custom_field"] != "custom_value" {
		t.Errorf("Extras[custom_field] = %v, want %q",
			identity.Metadata.Extras["custom_field"], "custom_value")
	}

	// Verify custom_number is captured as well
	customNumber, ok := identity.Metadata.Extras["custom_number"]
	if !ok {
		t.Error("Extras should contain custom_number")
	}
	// YAML unmarshals integers as int
	if customNumber != 42 {
		t.Errorf("Extras[custom_number] = %v, want 42", customNumber)
	}

	// Verify known fields are NOT in Extras (they should be in their proper fields)
	if _, exists := identity.Metadata.Extras["project_url"]; exists {
		t.Error("Extras should not contain known field 'project_url'")
	}
	if _, exists := identity.Metadata.Extras["license"]; exists {
		t.Error("Extras should not contain known field 'license'")
	}
}

// TestFindIdentityFile verifies ancestor directory search.
func TestFindIdentityFile(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create nested structure:
	// tmp/
	//   .fulmen/app.yaml (root level)
	//   project/
	//     subdir/
	//       deep/
	//         (search starts here)

	rootIdentity := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(rootIdentity, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	rootIdentityFile := filepath.Join(rootIdentity, "app.yaml")
	content := []byte("app:\n  binary_name: test\n  vendor: test\n  env_prefix: TEST_\n  config_name: test\n  description: Test application\n")
	if err := os.WriteFile(rootIdentityFile, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	deepDir := filepath.Join(tmpDir, "project", "subdir", "deep")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}

	// Test: Search from deep directory should find root identity
	foundPath, err := findIdentityFile(deepDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if foundPath != rootIdentityFile {
		t.Errorf("findIdentityFile() = %q, want %q", foundPath, rootIdentityFile)
	}
}

// TestFindIdentityFileNotFound verifies not-found error with searched paths.
func TestFindIdentityFileNotFound(t *testing.T) {
	// Create temporary directory with NO identity file
	tmpDir := t.TempDir()
	deepDir := filepath.Join(tmpDir, "project", "subdir")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	_, err := findIdentityFile(deepDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}

	if len(notFoundErr.SearchedPaths) == 0 {
		t.Error("NotFoundError should include searched paths")
	}

	if notFoundErr.StartDir == "" {
		t.Error("NotFoundError should include start directory")
	}
}

// TestEnvIdentityPath verifies environment variable override.
func TestEnvIdentityPath(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "custom-app.yaml")
	content := []byte("app:\n  binary_name: envtest\n  vendor: envtest\n  env_prefix: ENVTEST_\n  config_name: envtest\n  description: Environment test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Set environment variable
	oldEnv := os.Getenv(EnvIdentityPath)
	defer func() {
		if oldEnv != "" {
			if err := os.Setenv(EnvIdentityPath, oldEnv); err != nil {
				t.Errorf("failed to restore env: %v", err)
			}
		} else {
			if err := os.Unsetenv(EnvIdentityPath); err != nil {
				t.Errorf("failed to unset env: %v", err)
			}
		}
	}()

	if err := os.Setenv(EnvIdentityPath, identityPath); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	// Search from a directory that doesn't contain .fulmen/app.yaml
	otherDir := t.TempDir()
	foundPath, err := findIdentityFile(otherDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find the env-specified path, not search ancestors
	if foundPath != identityPath {
		t.Errorf("findIdentityFile() = %q, want %q (from env)", foundPath, identityPath)
	}
}

// TestEnvIdentityPathNotFound verifies env var with nonexistent file.
func TestEnvIdentityPathNotFound(t *testing.T) {
	oldEnv := os.Getenv(EnvIdentityPath)
	defer func() {
		if oldEnv != "" {
			if err := os.Setenv(EnvIdentityPath, oldEnv); err != nil {
				t.Errorf("failed to restore env: %v", err)
			}
		} else {
			if err := os.Unsetenv(EnvIdentityPath); err != nil {
				t.Errorf("failed to unset env: %v", err)
			}
		}
	}()

	// Set to nonexistent path
	if err := os.Setenv(EnvIdentityPath, "/nonexistent/path/app.yaml"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	tmpDir := t.TempDir()
	_, err := findIdentityFile(tmpDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}

	// Should show the env var path in error
	if len(notFoundErr.SearchedPaths) == 0 {
		t.Error("should include env var path in searched paths")
	}
}

// TestMalformedError verifies YAML parsing errors.
func TestMalformedError(t *testing.T) {
	ctx := context.Background()
	fixturePath := filepath.Join("testdata", "invalid-format.yaml")

	_, err := LoadFrom(ctx, fixturePath)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}

	var malformedErr *MalformedError
	if !errors.As(err, &malformedErr) {
		t.Fatalf("expected MalformedError, got %T", err)
	}

	if malformedErr.Path != fixturePath {
		t.Errorf("MalformedError.Path = %q, want %q", malformedErr.Path, fixturePath)
	}

	if malformedErr.Err == nil {
		t.Error("MalformedError.Err should contain underlying parse error")
	}

	// Verify error message includes path
	errMsg := malformedErr.Error()
	if errMsg == "" {
		t.Error("MalformedError.Error() should return non-empty message")
	}
}
