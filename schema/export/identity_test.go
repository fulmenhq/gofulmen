package export

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultIdentityProvider_NoFile(t *testing.T) {
	// Change to a temp directory with no .fulmen/app.yaml
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	provider := NewDefaultIdentityProvider()
	identity, err := provider.GetIdentity(context.Background())

	// Should return nil identity (no error) when file doesn't exist
	require.NoError(t, err)
	assert.Nil(t, identity)
}

func TestDefaultIdentityProvider_WithFile(t *testing.T) {
	// Create temp directory with .fulmen/app.yaml
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	// Create .fulmen/app.yaml
	fulmenDir := filepath.Join(tempDir, ".fulmen")
	err = os.MkdirAll(fulmenDir, 0755)
	require.NoError(t, err)

	appYAML := filepath.Join(fulmenDir, "app.yaml")
	appContent := `vendor: fulmenhq
binary: test-app
`
	err = os.WriteFile(appYAML, []byte(appContent), 0644)
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	provider := NewDefaultIdentityProvider()
	identity, err := provider.GetIdentity(context.Background())

	require.NoError(t, err)
	require.NotNil(t, identity)
	assert.Equal(t, "fulmenhq", identity.Vendor)
	assert.Equal(t, "test-app", identity.Binary)
}

func TestDefaultIdentityProvider_NameFallback(t *testing.T) {
	// Test using 'name' field as fallback for 'binary'
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	fulmenDir := filepath.Join(tempDir, ".fulmen")
	err = os.MkdirAll(fulmenDir, 0755)
	require.NoError(t, err)

	appYAML := filepath.Join(fulmenDir, "app.yaml")
	appContent := `vendor: mycompany
name: my-app
`
	err = os.WriteFile(appYAML, []byte(appContent), 0644)
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	provider := NewDefaultIdentityProvider()
	identity, err := provider.GetIdentity(context.Background())

	require.NoError(t, err)
	require.NotNil(t, identity)
	assert.Equal(t, "mycompany", identity.Vendor)
	assert.Equal(t, "my-app", identity.Binary) // name used as binary
}

func TestDefaultIdentityProvider_InvalidYAML(t *testing.T) {
	// Invalid YAML should return nil (graceful fallback)
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	fulmenDir := filepath.Join(tempDir, ".fulmen")
	err = os.MkdirAll(fulmenDir, 0755)
	require.NoError(t, err)

	appYAML := filepath.Join(fulmenDir, "app.yaml")
	err = os.WriteFile(appYAML, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	provider := NewDefaultIdentityProvider()
	identity, err := provider.GetIdentity(context.Background())

	// Should gracefully handle invalid YAML
	require.NoError(t, err)
	assert.Nil(t, identity)
}

func TestDefaultIdentityProvider_EmptyFields(t *testing.T) {
	// Empty vendor and binary should return nil
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	fulmenDir := filepath.Join(tempDir, ".fulmen")
	err = os.MkdirAll(fulmenDir, 0755)
	require.NoError(t, err)

	appYAML := filepath.Join(fulmenDir, "app.yaml")
	appContent := `version: "1.0.0"
description: Test app
`
	err = os.WriteFile(appYAML, []byte(appContent), 0644)
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	provider := NewDefaultIdentityProvider()
	identity, err := provider.GetIdentity(context.Background())

	require.NoError(t, err)
	assert.Nil(t, identity) // No vendor or binary
}

func TestFindAppYAML_WalksUp(t *testing.T) {
	// Create nested structure: parent/.fulmen/app.yaml, parent/child/
	tempDir := t.TempDir()
	fulmenDir := filepath.Join(tempDir, ".fulmen")
	childDir := filepath.Join(tempDir, "child", "nested")

	err := os.MkdirAll(fulmenDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(childDir, 0755)
	require.NoError(t, err)

	appYAML := filepath.Join(fulmenDir, "app.yaml")
	err = os.WriteFile(appYAML, []byte("vendor: test\n"), 0644)
	require.NoError(t, err)

	// Change to child directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(childDir)
	require.NoError(t, err)

	// Should find app.yaml in parent
	found, err := findAppYAML()
	require.NoError(t, err)

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	expectedPath, err := filepath.EvalSymlinks(appYAML)
	require.NoError(t, err)
	foundPath, err := filepath.EvalSymlinks(found)
	require.NoError(t, err)

	assert.Equal(t, expectedPath, foundPath)
}
