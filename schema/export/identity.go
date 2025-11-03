package export

import (
	"context"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// defaultIdentityProvider reads application identity from .fulmen/app.yaml
type defaultIdentityProvider struct{}

// GetIdentity reads identity from .fulmen/app.yaml with graceful fallback
// Returns nil if the file doesn't exist or can't be parsed (not an error)
func (p *defaultIdentityProvider) GetIdentity(ctx context.Context) (*Identity, error) {
	// Look for .fulmen/app.yaml starting from current directory
	identityFile, err := findAppYAML()
	if err != nil {
		// File not found - return nil (not an error)
		return nil, nil
	}

	// Read the file
	data, err := os.ReadFile(identityFile)
	if err != nil {
		// Can't read file - return nil (graceful fallback)
		return nil, nil
	}

	// Parse the YAML
	var appConfig struct {
		Vendor string `yaml:"vendor"`
		Binary string `yaml:"binary"`
		Name   string `yaml:"name"` // Alternative field name
	}

	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		// Can't parse - return nil (graceful fallback)
		return nil, nil
	}

	// Build identity from parsed data
	identity := &Identity{
		Vendor: appConfig.Vendor,
		Binary: appConfig.Binary,
	}

	// Use 'name' as fallback for 'binary' if binary is empty
	if identity.Binary == "" && appConfig.Name != "" {
		identity.Binary = appConfig.Name
	}

	// Only return identity if we have at least vendor or binary
	if identity.Vendor == "" && identity.Binary == "" {
		return nil, nil
	}

	return identity, nil
}

// findAppYAML searches for .fulmen/app.yaml starting from current directory
// and walking up to the root
func findAppYAML() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree
	dir := cwd
	for {
		appYAML := filepath.Join(dir, ".fulmen", "app.yaml")
		if _, err := os.Stat(appYAML); err == nil {
			return appYAML, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// NewDefaultIdentityProvider creates a default identity provider
// that reads from .fulmen/app.yaml with graceful fallback
func NewDefaultIdentityProvider() IdentityProvider {
	return &defaultIdentityProvider{}
}
