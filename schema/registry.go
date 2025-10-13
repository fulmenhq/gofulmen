package schema

import (
	"fmt"
	"path/filepath"
	"sync"
)

// SchemaRegistry manages versioned schemas
type SchemaRegistry struct {
	baseDir string
	cache   map[string][]byte
	mutex   sync.RWMutex
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry(baseDir string) *SchemaRegistry {
	return &SchemaRegistry{
		baseDir: baseDir,
		cache:   make(map[string][]byte),
	}
}

// NewDefaultSchemaRegistry creates a registry pointing to ./crucible/schemas
func NewDefaultSchemaRegistry() *SchemaRegistry {
	return NewSchemaRegistry("./crucible/schemas")
}

// LoadSchema loads a schema by name and version
func (r *SchemaRegistry) LoadSchema(name, version string) ([]byte, error) {
	key := fmt.Sprintf("%s/%s", name, version)

	r.mutex.RLock()
	if data, exists := r.cache[key]; exists {
		r.mutex.RUnlock()
		return data, nil
	}
	r.mutex.RUnlock()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double-check after acquiring write lock
	if data, exists := r.cache[key]; exists {
		return data, nil
	}

	// Load from file
	dir := filepath.Join(r.baseDir, name, version)
	data, err := LoadSchemaFromDir(dir, "schema")
	if err != nil {
		return nil, fmt.Errorf("failed to load schema %s/%s: %w", name, version, err)
	}

	r.cache[key] = data
	return data, nil
}

// GetValidator gets a validator for a schema
func (r *SchemaRegistry) GetValidator(name, version string) (*Validator, error) {
	data, err := r.LoadSchema(name, version)
	if err != nil {
		return nil, err
	}
	return NewValidator(data)
}

// ClearCache clears the schema cache
func (r *SchemaRegistry) ClearCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cache = make(map[string][]byte)
}
