package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// Validator wraps a compiled JSON schema and tracks its descriptor when available.
type Validator struct {
	schema     *jsonschema.Schema
	descriptor SchemaDescriptor
	metaDir    string
}

// NewValidator compiles a schema from raw bytes. Intended for standalone schemas that
// do not rely on relative references.
func NewValidator(schemaData []byte) (*Validator, error) {
	metaDir := filepath.Join(defaultSchemaBaseDir, metaDirName)
	compiler, err := newCompiler(metaDir)
	if err != nil {
		return nil, err
	}

	const virtualURL = "memory://schema.json"
	if err := compiler.AddResource(virtualURL, strings.NewReader(string(schemaData))); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}
	compiled, err := compiler.Compile(virtualURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator{
		schema:  compiled,
		metaDir: metaDir,
	}, nil
}

func newValidatorFromDescriptor(desc SchemaDescriptor, metaDir string) (*Validator, error) {
	compiler, err := newCompiler(metaDir)
	if err != nil {
		return nil, err
	}

	schemaURL := fileURL(desc.Path)
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, err
	}

	return &Validator{
		schema:     compiled,
		descriptor: desc,
		metaDir:    metaDir,
	}, nil
}

// ValidateData validates an in-memory value against the schema and returns diagnostics.
func (v *Validator) ValidateData(data interface{}) ([]Diagnostic, error) {
	err := v.schema.Validate(data)
	if err == nil {
		return nil, nil
	}

	validationErr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return nil, err
	}
	return diagnosticsFromValidationError(validationErr, sourceGoFulmen), nil
}

// ValidateJSON validates JSON bytes.
func (v *Validator) ValidateJSON(jsonData []byte) ([]Diagnostic, error) {
	var payload interface{}
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return v.ValidateData(payload)
}

// ValidateFile validates a JSON or YAML file on disk.
func (v *Validator) ValidateFile(path string) ([]Diagnostic, error) {
	content, err := os.ReadFile(path) // #nosec G304 -- User-provided path is intentional for validation API
	if err != nil {
		return nil, err
	}

	if isJSON(content) {
		return v.ValidateJSON(content)
	}

	var payload interface{}
	if err := yaml.Unmarshal(content, &payload); err != nil {
		return nil, err
	}
	return v.ValidateData(payload)
}

func newCompiler(metaDir string) (*jsonschema.Compiler, error) {
	if metaDir == "" {
		return nil, fmt.Errorf("meta directory is required")
	}

	loader := &localLoader{metaDir: metaDir}
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = loader.Load
	return compiler, nil
}

// ValidateSchemaBytes validates a schema document against the embedded metaschema bundles.
func ValidateSchemaBytes(schemaBytes []byte) ([]Diagnostic, error) {
	metaDir := filepath.Join(resolveDefaultBaseDir(), metaDirName)
	compiler, err := newCompiler(metaDir)
	if err != nil {
		return nil, err
	}

	const schemaURL = "memory://schema.json"
	if err := compiler.AddResource(schemaURL, strings.NewReader(string(schemaBytes))); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	_, err = compiler.Compile(schemaURL)
	if err == nil {
		return nil, nil
	}

	if schemaErr, ok := err.(*jsonschema.SchemaError); ok {
		if validationErr, ok := schemaErr.Err.(*jsonschema.ValidationError); ok {
			return diagnosticsFromValidationError(validationErr, sourceGoFulmen), nil
		}
		return nil, schemaErr.Err
	}
	return nil, err
}

// ValidateSchemaByID validates the schema definition referenced by ID.
func (c *Catalog) ValidateSchemaByID(id string) ([]Diagnostic, error) {
	desc, err := c.GetSchema(id)
	if err != nil {
		return nil, err
	}

	compiler, err := newCompiler(c.metaDir)
	if err != nil {
		return nil, err
	}

	schemaURL := fileURL(desc.Path)
	if _, err := compiler.Compile(schemaURL); err != nil {
		if schemaErr, ok := err.(*jsonschema.SchemaError); ok {
			if validationErr, ok := schemaErr.Err.(*jsonschema.ValidationError); ok {
				return diagnosticsFromValidationError(validationErr, sourceGoFulmen), nil
			}
			return nil, schemaErr.Err
		}
		return nil, err
	}
	return nil, nil
}

// ValidateDataByID validates a JSON payload against the schema identified by ID.
func (c *Catalog) ValidateDataByID(id string, payload []byte) ([]Diagnostic, error) {
	validator, err := c.ValidatorByID(id)
	if err != nil {
		return nil, err
	}
	return validator.ValidateJSON(payload)
}

// ValidateFileByID validates a file (JSON or YAML) against the schema identified by ID.
func (c *Catalog) ValidateFileByID(id string, path string) ([]Diagnostic, error) {
	validator, err := c.ValidatorByID(id)
	if err != nil {
		return nil, err
	}
	return validator.ValidateFile(path)
}

type localLoader struct {
	metaDir string
}

func (l *localLoader) Load(rawURL string) (io.ReadCloser, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("empty schema url")
	}

	trimmed := stripFragment(rawURL)
	if strings.HasPrefix(trimmed, "file://") {
		return openFileURL(trimmed)
	}

	for _, prefix := range []string{
		"https://json-schema.org/draft/2020-12/",
		"http://json-schema.org/draft/2020-12/",
	} {
		if strings.HasPrefix(trimmed, prefix) {
			return l.openDraftResource(trimmed, prefix, "draft-2020-12")
		}
	}
	for _, prefix := range []string{
		"https://json-schema.org/draft-07/",
		"http://json-schema.org/draft-07/",
	} {
		if strings.HasPrefix(trimmed, prefix) {
			return l.openDraftResource(trimmed, prefix, "draft-07")
		}
	}
	if strings.HasPrefix(trimmed, "https://schemas.fulmenhq.dev/") {
		remainder := strings.TrimPrefix(trimmed, "https://schemas.fulmenhq.dev/")
		baseDir := filepath.Dir(l.metaDir) // schemas/crucible-go (parent of metaDir)

		// Map URL path to local directory structure
		// URL pattern:  https://schemas.fulmenhq.dev/crucible/<category>/<name>-v<version>.json
		// File pattern: schemas/crucible-go/<category>/v<version>/<name>.schema.json
		//
		// Examples:
		// - URL:  https://schemas.fulmenhq.dev/crucible/ascii/string-analysis-v1.0.0.json
		//   File: schemas/crucible-go/ascii/v1.0.0/string-analysis.schema.json
		// - URL:  https://schemas.fulmenhq.dev/crucible/observability/logging/middleware-config-v1.0.0.json
		//   File: schemas/crucible-go/observability/logging/v1.0.0/middleware-config.schema.json

		localPath := mapSchemaURLToPath(baseDir, remainder)
		return os.Open(localPath) // #nosec G304 -- Local schema path constructed from trusted prefix
	}

	// Allow loading local relative paths.
	if !strings.Contains(trimmed, "://") {
		// For relative paths, resolve them from repository root
		absPath := trimmed
		if !filepath.IsAbs(trimmed) {
			repoRoot := findRepoRoot()
			if repoRoot != "" {
				absPath = filepath.Join(repoRoot, trimmed)
			}
		}
		return os.Open(absPath) // #nosec G304 -- Schema reference path is validated by JSON Schema spec
	}

	return nil, fmt.Errorf("unsupported schema reference: %s", rawURL)
}

var (
	repoRootOnce sync.Once
	repoRootPath string
)

// findRepoRoot finds the repository root by looking for .git or go.mod
// This is a simplified version to avoid import cycle with pathfinder package
func findRepoRoot() string {
	repoRootOnce.Do(func() {
		// Start from current working directory
		cwd, err := os.Getwd()
		if err != nil {
			repoRootPath = "" // fallback
			return
		}

		// Walk upward looking for .git or go.mod
		currentDir := cwd
		for {
			// Check for markers
			for _, marker := range []string{".git", "go.mod"} {
				markerPath := filepath.Join(currentDir, marker)
				if _, err := os.Stat(markerPath); err == nil {
					repoRootPath = currentDir
					return
				}
			}

			// Move up one directory
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				// Reached filesystem root without finding marker
				repoRootPath = cwd // fallback to cwd
				return
			}
			currentDir = parentDir
		}
	})
	return repoRootPath
}

// mapSchemaURLToPath maps a schema.fulmenhq.dev URL path to local file structure
// URL pattern:  crucible/<category>/<name>-v<version>.json
// File pattern: <baseDir>/<category>/v<version>/<name>.schema.json
func mapSchemaURLToPath(baseDir, urlPath string) string {
	// Strip '/crucible/' prefix since baseDir already points to schemas/crucible-go
	urlPath = strings.TrimPrefix(urlPath, "crucible/")

	// Parse out the pattern: <category-path>/<name>-v<version>.json
	// Example: observability/logging/middleware-config-v1.0.0.json
	parts := strings.Split(urlPath, "/")
	if len(parts) == 0 {
		return filepath.Join(baseDir, urlPath)
	}

	// Last part contains: <name>-v<version>.json
	filename := parts[len(parts)-1]
	categoryPath := strings.Join(parts[:len(parts)-1], "/")

	// Extract version from filename: <name>-v<version>.json -> v<version>
	// Example: middleware-config-v1.0.0.json -> v1.0.0
	versionPattern := "-v"
	versionIdx := strings.LastIndex(filename, versionPattern)

	var baseName, version string
	if versionIdx == -1 {
		// No version in filename - this is a relative reference like "severity-filter.schema.json"
		// Try to infer version from the category path by scanning the directory
		baseName = strings.TrimSuffix(filename, ".json")
		baseName = strings.TrimSuffix(baseName, ".schema")

		// Look for the schema file in the directory structure
		// Try v1.0.0 first (most common), then scan if needed
		repoRoot := findRepoRoot()
		if repoRoot != "" {
			categoryFullPath := filepath.Join(repoRoot, baseDir, categoryPath)
			// Try to find version subdirectories
			entries, err := os.ReadDir(categoryFullPath)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() && strings.HasPrefix(entry.Name(), "v") {
						// Found a version directory, try this one
						version = entry.Name()
						break
					}
				}
			}
		}

		// Fallback to v1.0.0 if we couldn't find a version
		if version == "" {
			version = "v1.0.0"
		}
	} else {
		baseName = filename[:versionIdx]                           // middleware-config
		versionAndExt := filename[versionIdx+len(versionPattern):] // 1.0.0.json
		version = "v" + strings.TrimSuffix(versionAndExt, ".json") // v1.0.0
	}

	// Construct local path: <baseDir>/<category>/v<version>/<name>.schema.json
	// NOTE: Use repository root to make baseDir absolute, avoiding test working directory issues
	absBaseDir := baseDir
	if !filepath.IsAbs(baseDir) {
		// Find repository root first
		repoRoot := findRepoRoot()
		if repoRoot != "" {
			// Resolve baseDir relative to repo root
			absBaseDir = filepath.Join(repoRoot, baseDir)
		} else {
			// Fallback to current approach
			var err error
			absBaseDir, err = filepath.Abs(baseDir)
			if err != nil {
				absBaseDir = baseDir
			}
		}
	}

	localFilename := baseName + ".schema.json"

	// Check if categoryPath already contains a version directory (e.g., "pathfinder/v1.0.0")
	// This happens with relative references like "../../pathfinder/v1.0.0/error-response.schema.json"
	categoryParts := strings.Split(categoryPath, "/")
	hasVersionInPath := false
	for _, part := range categoryParts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			// Check if it looks like a version (v1.0.0, v2.1.3, etc.)
			rest := part[1:]
			if len(rest) > 0 && (rest[0] >= '0' && rest[0] <= '9') {
				hasVersionInPath = true
				break
			}
		}
	}

	var localPath string
	if hasVersionInPath {
		// Version already in category path, don't add it again
		localPath = filepath.Join(absBaseDir, categoryPath, localFilename)
	} else {
		// Add version directory
		localPath = filepath.Join(absBaseDir, categoryPath, version, localFilename)
	}

	return localPath
}

func (l *localLoader) openDraftResource(raw, prefix, draftDir string) (io.ReadCloser, error) {
	remainder := strings.TrimPrefix(raw, prefix)
	remainder = strings.TrimSuffix(remainder, ".json")
	remainder = strings.TrimPrefix(remainder, "/")

	var relPath string
	switch remainder {
	case "", "schema":
		relPath = filepath.Join(draftDir, "schema.json")
	default:
		relPath = filepath.Join(draftDir, "meta", remainder+".json")
	}

	full := filepath.Join(l.metaDir, relPath)
	return os.Open(full) // #nosec G304 -- Metaschema path is constructed from trusted embedded assets
}

func stripFragment(raw string) string {
	if idx := strings.IndexRune(raw, '#'); idx >= 0 {
		return raw[:idx]
	}
	return raw
}

func openFileURL(raw string) (io.ReadCloser, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	path := u.Path
	if path == "" {
		return nil, fmt.Errorf("empty file path in url: %s", raw)
	}
	return os.Open(path) // #nosec G304 -- File URL path is parsed from schema reference
}

func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(abs),
	}).String()
}
