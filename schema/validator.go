package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
		baseDir := filepath.Dir(l.metaDir) // schemas/crucible-go
		localPath := filepath.Join(baseDir, remainder)
		return os.Open(localPath) // #nosec G304 -- Local schema path constructed from trusted prefix
	}

	// Allow loading local relative paths.
	if !strings.Contains(trimmed, "://") {
		return os.Open(trimmed) // #nosec G304 -- Schema reference path is validated by JSON Schema spec
	}

	return nil, fmt.Errorf("unsupported schema reference: %s", rawURL)
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
