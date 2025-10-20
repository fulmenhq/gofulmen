package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	defaultSchemaBaseDir = "schemas/crucible-go"
	metaDirName          = "meta"
)

// SchemaDescriptor describes a schema available in the local catalog.
type SchemaDescriptor struct {
	ID          string
	Category    string
	Version     string
	Name        string
	Path        string
	Draft       string
	Title       string
	Description string
}

// SchemaDiff describes a difference between schemas.
type SchemaDiff struct {
	Path    string
	Message string
}

// Catalog indexes schemas stored on disk.
type Catalog struct {
	baseDir string
	metaDir string

	mu          sync.RWMutex
	descriptors map[string]SchemaDescriptor
	validators  map[string]*Validator
	loaded      bool
}

// NewCatalog creates a catalog rooted at baseDir.
func NewCatalog(baseDir string) *Catalog {
	baseDir = filepath.Clean(baseDir)
	metaDir := filepath.Join(baseDir, metaDirName)
	return &Catalog{
		baseDir:     baseDir,
		metaDir:     metaDir,
		descriptors: make(map[string]SchemaDescriptor),
		validators:  make(map[string]*Validator),
	}
}

// DefaultCatalog returns a catalog rooted at the synced Crucible schemas directory.
func DefaultCatalog() *Catalog {
	return globalCatalog()
}

// ListSchemas returns descriptors whose IDs match the prefix (if provided).
func (c *Catalog) ListSchemas(prefix string) ([]SchemaDescriptor, error) {
	if err := c.ensureLoaded(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []SchemaDescriptor
	for id, desc := range c.descriptors {
		if prefix == "" || strings.HasPrefix(id, prefix) {
			result = append(result, desc)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

// GetSchema returns the descriptor for the given schema ID.
func (c *Catalog) GetSchema(id string) (SchemaDescriptor, error) {
	if err := c.ensureLoaded(); err != nil {
		return SchemaDescriptor{}, err
	}

	c.mu.RLock()
	desc, ok := c.descriptors[id]
	c.mu.RUnlock()
	if !ok {
		return SchemaDescriptor{}, fmt.Errorf("schema %q not found", id)
	}
	return desc, nil
}

// CompareSchema compares the catalog schema to the provided schema (JSON or YAML).
func (c *Catalog) CompareSchema(id string, other []byte) ([]SchemaDiff, error) {
	desc, err := c.GetSchema(id)
	if err != nil {
		return nil, err
	}

	canonicalCatalog, err := loadAndNormalize(desc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema %s: %w", id, err)
	}

	canonicalOther, err := normalizeSchemaBytes(other)
	if err != nil {
		return nil, fmt.Errorf("failed to parse comparison schema: %w", err)
	}

	if string(canonicalCatalog) == string(canonicalOther) {
		return nil, nil
	}

	return []SchemaDiff{
		{
			Path:    id,
			Message: "schemas differ",
		},
	}, nil
}

// ValidatorByID returns (and caches) a validator for the schema ID.
func (c *Catalog) ValidatorByID(id string) (*Validator, error) {
	desc, err := c.GetSchema(id)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	if v, ok := c.validators[id]; ok {
		c.mu.RUnlock()
		return v, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.validators[id]; ok {
		return v, nil
	}

	validator, err := newValidatorFromDescriptor(desc, c.metaDir)
	if err != nil {
		return nil, err
	}
	c.validators[id] = validator
	return validator, nil
}

func (c *Catalog) ensureLoaded() error {
	c.mu.RLock()
	if c.loaded {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return nil
	}

	err := filepath.WalkDir(c.baseDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			// skip meta directory
			if filepath.Base(path) == metaDirName {
				return filepath.SkipDir
			}
			return nil
		}

		if !isSchemaFile(d.Name()) {
			return nil
		}

		desc, err := c.buildDescriptor(path)
		if err != nil {
			return err
		}
		if desc.ID != "" {
			c.descriptors[desc.ID] = desc
		}
		return nil
	})
	if err != nil {
		return err
	}
	c.loaded = true
	return nil
}

func (c *Catalog) buildDescriptor(path string) (SchemaDescriptor, error) {
	rel, err := filepath.Rel(c.baseDir, path)
	if err != nil {
		return SchemaDescriptor{}, err
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 3 {
		return SchemaDescriptor{}, nil
	}

	version := parts[len(parts)-2]
	name := strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
	name = strings.TrimSuffix(name, ".schema")
	category := strings.Join(parts[:len(parts)-2], "/")

	data, err := loadAndNormalize(path)
	if err != nil {
		return SchemaDescriptor{}, err
	}

	var meta struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Schema      string `json:"$schema"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return SchemaDescriptor{}, fmt.Errorf("failed to parse schema metadata for %s: %w", path, err)
	}

	return SchemaDescriptor{
		ID:          buildSchemaID(category, version, name),
		Category:    category,
		Version:     version,
		Name:        name,
		Path:        filepath.Clean(path),
		Draft:       meta.Schema,
		Title:       meta.Title,
		Description: meta.Description,
	}, nil
}

func buildSchemaID(category, version, name string) string {
	return fmt.Sprintf("%s/%s/%s", category, version, name)
}

func isSchemaFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}

func resolveDefaultBaseDir() string {
	candidate := defaultSchemaBaseDir
	if pathExists(candidate) {
		return candidate
	}

	if cwd, err := os.Getwd(); err == nil {
		current := cwd
		for i := 0; i < 4; i++ {
			path := filepath.Join(current, defaultSchemaBaseDir)
			if pathExists(path) {
				return path
			}
			next := filepath.Dir(current)
			if next == current {
				break
			}
			current = next
		}
	}
	return defaultSchemaBaseDir
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var (
	defaultCatalogOnce sync.Once
	defaultCatalogInst *Catalog
)

func globalCatalog() *Catalog {
	defaultCatalogOnce.Do(func() {
		defaultCatalogInst = NewCatalog(resolveDefaultBaseDir())
	})
	return defaultCatalogInst
}

func loadAndNormalize(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return normalizeSchemaBytes(raw)
}

func normalizeSchemaBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty schema payload")
	}

	if isJSON(data) {
		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			return nil, err
		}
		return json.Marshal(value)
	}

	var yamlValue interface{}
	if err := yaml.Unmarshal(data, &yamlValue); err != nil {
		return nil, err
	}
	return json.Marshal(yamlValue)
}

func isJSON(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}
