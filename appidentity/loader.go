package appidentity

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultIdentityFilename is the standard name for identity files.
	DefaultIdentityFilename = "app.yaml"

	// DefaultIdentityDir is the standard directory for identity files.
	DefaultIdentityDir = ".fulmen"

	// DefaultIdentityPath is the standard relative path to identity files.
	DefaultIdentityPath = ".fulmen/app.yaml"

	// EnvIdentityPath is the environment variable for explicit identity file path.
	EnvIdentityPath = "FULMEN_APP_IDENTITY_PATH"

	// MaxSearchDepth is the maximum number of parent directories to search.
	MaxSearchDepth = 20
)

var (
	osExecutable = os.Executable
)

// identityFile represents the YAML file structure with "app" and "metadata" keys.
type identityFile struct {
	App      Identity `yaml:"app" json:"app"`
	Metadata Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Options controls identity loading behavior.
//
// IMPORTANT: When using GetWithOptions for the first time in a process,
// the options (especially ExplicitPath) determine what identity gets cached
// for the process lifetime. Subsequent calls to Get() will return that
// same cached identity, even if they don't specify the same options.
//
// Example of side effect:
//
//	// First call with explicit path - caches identity from custom location
//	identity1, _ := GetWithOptions(ctx, Options{ExplicitPath: "/custom/app.yaml"})
//
//	// Later call without options - returns same identity from cache
//	identity2, _ := Get(ctx)  // identity2 == identity1 (from /custom/app.yaml)
//
// To avoid confusion, either:
//  1. Use Get() exclusively and rely on standard discovery, OR
//  2. Use GetWithOptions with NoCache: true to bypass caching, OR
//  3. Call Reset() between test cases to clear the cache
type Options struct {
	// ExplicitPath forces loading from a specific path (highest priority).
	// If set, no discovery is performed.
	// WARNING: The first GetWithOptions call with an ExplicitPath will cache
	// that identity for the process lifetime (unless NoCache is true).
	ExplicitPath string

	// RepoRoot sets the starting point for ancestor search.
	// Default: current working directory.
	RepoRoot string

	// NoCache bypasses the process-level cache (testing only).
	// When true, each call loads identity fresh from disk.
	NoCache bool
}

// LoadFrom loads identity from an explicit file path without caching or discovery.
//
// This function is useful for testing or when you need to load identity from a
// non-standard location. It does not perform validation - use Validate() separately
// if schema validation is needed.
//
// Example:
//
//	identity, err := appidentity.LoadFrom(ctx, "/path/to/custom/app.yaml")
//	if err != nil {
//	    return fmt.Errorf("failed to load identity: %w", err)
//	}
func LoadFrom(ctx context.Context, path string) (*Identity, error) {
	identity, err := loadIdentityFile(path)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

// loadIdentityFile reads and parses a YAML identity file.
func loadIdentityFile(path string) (*Identity, error) {
	safePath, err := normalizeIdentityPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(safePath) // #nosec G304,G703 -- Explicit identity paths are normalized before use.
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{
				SearchedPaths: []string{path},
			}
		}
		return nil, fmt.Errorf("failed to read identity file: %w", err)
	}

	var file identityFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, &MalformedError{
			Path: path,
			Err:  err,
		}
	}

	// Copy metadata from file-level to Identity struct
	file.App.Metadata = file.Metadata

	return &file.App, nil
}

// findIdentityFile searches for .fulmen/app.yaml starting from startDir and walking upward.
//
// Returns the absolute path to the identity file if found, or an error with searched paths.
func findIdentityFile(startDir string) (string, error) {
	// Check environment variable first
	if envPath := os.Getenv(EnvIdentityPath); envPath != "" {
		absPath, err := normalizeIdentityPath(envPath)
		if err != nil {
			return "", fmt.Errorf("invalid %s path: %w", EnvIdentityPath, err)
		}
		if _, err := os.Stat(absPath); err == nil { // #nosec G703 -- Environment override is normalized before probing.
			return absPath, nil
		}
		return "", &NotFoundError{
			SearchedPaths: []string{absPath + " (from " + EnvIdentityPath + ")"},
		}
	}

	// Ensure startDir is absolute
	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("invalid start directory: %w", err)
	}

	searchedPaths := make([]string, 0, MaxSearchDepth)
	currentDir := absStartDir

	// Walk up the directory tree
	for depth := 0; depth < MaxSearchDepth; depth++ {
		identityPath := filepath.Join(currentDir, DefaultIdentityPath)
		searchedPaths = append(searchedPaths, identityPath)

		if _, err := os.Stat(identityPath); err == nil {
			return identityPath, nil
		}

		// Check if we've reached the filesystem root
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root (e.g., "/" on Unix or "C:\" on Windows)
			break
		}
		currentDir = parentDir
	}

	return "", &NotFoundError{
		SearchedPaths: searchedPaths,
		StartDir:      absStartDir,
	}
}

func normalizeIdentityPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("identity path cannot be empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absPath), nil
}

// executableStartDir returns the directory that should be used as the start point
// for the executable-dir fallback search.
func executableStartDir() (string, error) {
	exePath, err := osExecutable()
	if err != nil {
		return "", fmt.Errorf("failed to determine executable path: %w", err)
	}
	if exePath == "" {
		return "", fmt.Errorf("failed to determine executable path: empty path")
	}

	absExePath, err := filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return filepath.Dir(absExePath), nil
}

// discoverIdentity discovers and loads identity using the standard search process.
//
// Discovery precedence:
//  1. Context injection (checked by caller)
//  2. ExplicitPath in Options
//  3. Environment variable (FULMEN_APP_IDENTITY_PATH)
//  4. Embedded identity (registered via RegisterEmbeddedIdentityYAML)
//  5. Nearest ancestor search from RepoRoot (default: cwd)
//  6. Fallback: Nearest ancestor search from executable directory
func discoverIdentity(ctx context.Context, opts Options) (*Identity, error) {
	var identityPath string
	var err error

	// Priority 1: Explicit path in options
	if opts.ExplicitPath != "" {
		identityPath = opts.ExplicitPath
		if _, err := os.Stat(identityPath); err != nil {
			if os.IsNotExist(err) {
				return nil, &NotFoundError{
					SearchedPaths: []string{identityPath + " (explicit path)"},
				}
			}
			return nil, fmt.Errorf("failed to access identity file: %w", err)
		}

		return loadIdentityFile(identityPath)
	}

	// Priority 2-6: Environment variable, embedded identity, ancestor search,
	// and fallback.
	startDir := opts.RepoRoot
	if startDir == "" {
		startDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	if os.Getenv(EnvIdentityPath) != "" {
		identityPath, err = findIdentityFile(startDir)
		if err == nil {
			return loadIdentityFile(identityPath)
		}
		return nil, err
	}

	if embedded, ok := getEmbeddedIdentity(); ok {
		return embedded, nil
	}

	identityPath, err = findIdentityFile(startDir)
	if err == nil {
		return loadIdentityFile(identityPath)
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		return nil, err
	}

	exeStartDir, exeErr := executableStartDir()
	if exeErr != nil {
		return nil, err
	}

	fallbackPath, fallbackErr := findIdentityFile(exeStartDir)
	if fallbackErr == nil {
		return loadIdentityFile(fallbackPath)
	}

	var fallbackNotFoundErr *NotFoundError
	if errors.As(fallbackErr, &fallbackNotFoundErr) {
		return nil, &NotFoundError{
			SearchedPaths:         notFoundErr.SearchedPaths,
			StartDir:              notFoundErr.StartDir,
			FallbackStartDir:      fallbackNotFoundErr.StartDir,
			FallbackSearchedPaths: fallbackNotFoundErr.SearchedPaths,
		}
	}

	return nil, fallbackErr
}
