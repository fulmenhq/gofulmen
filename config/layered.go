package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

// LayeredConfigOptions describes how to load and merge configuration layers.
type LayeredConfigOptions struct {
	Category     string          // Crucible category (e.g., "terminal")
	Version      string          // Version directory (e.g., "v1.0.0")
	DefaultsFile string          // Defaults file name (e.g., "terminal-overrides-defaults.yaml")
	SchemaID     string          // Catalog schema ID (e.g., "terminal/v1.0.0/schema")
	UserPaths    []string        // Explicit user override file paths (checked in order)
	DefaultsRoot string          // Optional override for defaults root (defaults to config/crucible-go)
	Catalog      *schema.Catalog // Optional catalog to use for validation
}

// LoadLayeredConfig loads defaults, applies user overrides, then applies runtime overrides.
// Returns merged configuration map and validation diagnostics (empty when valid).
func LoadLayeredConfig(opts LayeredConfigOptions, runtimeOverrides ...map[string]any) (map[string]any, []schema.Diagnostic, error) {
	if opts.Category == "" || opts.Version == "" || opts.DefaultsFile == "" {
		return nil, nil, errors.New("category, version, and defaults file are required")
	}
	if opts.SchemaID == "" {
		return nil, nil, errors.New("schema id is required for validation")
	}

	defaultsRoot := opts.DefaultsRoot
	if defaultsRoot == "" {
		defaultsRoot = defaultConfigBaseDir()
	}

	defaultsPath := filepath.Join(defaultsRoot, opts.Category, opts.Version, opts.DefaultsFile)
	defaultLayer, err := loadConfigFile(defaultsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load defaults %s: %w", defaultsPath, err)
	}

	merged := defaultLayer

	if len(opts.UserPaths) > 0 {
		for _, path := range opts.UserPaths {
			if path == "" {
				continue
			}
			data, err := loadConfigFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, nil, fmt.Errorf("load user config %s: %w", path, err)
			}
			merged = mergeMaps(merged, data)
			break
		}
	}

	for _, override := range runtimeOverrides {
		if override == nil {
			continue
		}
		merged = mergeMaps(merged, deepCopyMap(override))
	}

	payload, err := json.Marshal(merged)
	if err != nil {
		return nil, nil, fmt.Errorf("encode merged config: %w", err)
	}

	catalog := opts.Catalog
	if catalog == nil {
		catalog = schemaCatalog()
	}

	diags, err := catalog.ValidateDataByID(opts.SchemaID, payload)
	if err != nil {
		return nil, diags, fmt.Errorf("config validation failed: %w", err)
	}

	return merged, diags, nil
}

func loadConfigFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- Config path is from trusted XDG directories
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}

	switch ext := filepath.Ext(path); ext {
	case ".yaml", ".yml":
		var content any
		if err := yaml.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
		return normalizeToStringMap(content)
	case ".json":
		var content any
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
		return normalizeToStringMap(content)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}
}

func mergeMaps(base, overlay map[string]any) map[string]any {
	if base == nil {
		base = make(map[string]any)
	}
	if overlay == nil {
		return base
	}

	for key, value := range overlay {
		if value == nil {
			delete(base, key)
			continue
		}

		switch ov := value.(type) {
		case map[string]any:
			if existing, ok := base[key].(map[string]any); ok {
				base[key] = mergeMaps(existing, ov)
			} else {
				base[key] = deepCopyMap(ov)
			}
		case []any:
			base[key] = deepCopySlice(ov)
		default:
			base[key] = ov
		}
	}
	return base
}

func normalizeToStringMap(value any) (map[string]any, error) {
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, val := range v {
			nv, err := normalizeValue(val)
			if err != nil {
				return nil, err
			}
			result[key] = nv
		}
		return result, nil
	case map[any]any:
		result := make(map[string]any, len(v))
		for key, val := range v {
			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key %v", key)
			}
			nv, err := normalizeValue(val)
			if err != nil {
				return nil, err
			}
			result[strKey] = nv
		}
		return result, nil
	default:
		return nil, errors.New("config file must contain an object at top level")
	}
}

func normalizeValue(value any) (any, error) {
	switch val := value.(type) {
	case map[string]any:
		return normalizeToStringMap(val)
	case map[any]any:
		return normalizeToStringMap(val)
	case []any:
		out := make([]any, len(val))
		for i, elem := range val {
			nv, err := normalizeValue(elem)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	default:
		return val, nil
	}
}

func deepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	copy := make(map[string]any, len(src))
	for k, v := range src {
		switch tv := v.(type) {
		case map[string]any:
			copy[k] = deepCopyMap(tv)
		case []any:
			copy[k] = deepCopySlice(tv)
		default:
			copy[k] = tv
		}
	}
	return copy
}

func deepCopySlice(src []any) []any {
	if src == nil {
		return nil
	}
	out := make([]any, len(src))
	for i, val := range src {
		switch tv := val.(type) {
		case map[string]any:
			out[i] = deepCopyMap(tv)
		case []any:
			out[i] = deepCopySlice(tv)
		default:
			out[i] = tv
		}
	}
	return out
}

var (
	configDefaultsOnce sync.Once
	configDefaultsDir  string
	schemaCatalogOnce  sync.Once
	schemaCatalogInst  *schema.Catalog
)

func defaultConfigBaseDir() string {
	configDefaultsOnce.Do(func() {
		configDefaultsDir = resolveConfigBaseDir()
	})
	return configDefaultsDir
}

func resolveConfigBaseDir() string {
	candidate := filepath.Join("config", "crucible-go")
	if dirExists(candidate) {
		abs, err := filepath.Abs(candidate)
		if err == nil {
			return abs
		}
		return candidate
	}

	if cwd, err := os.Getwd(); err == nil {
		current := cwd
		for i := 0; i < 4; i++ {
			path := filepath.Join(current, "config", "crucible-go")
			if dirExists(path) {
				return path
			}
			next := filepath.Dir(current)
			if next == current {
				break
			}
			current = next
		}
	}
	// Fallback to relative path; caller will error if missing.
	return filepath.Join("config", "crucible-go")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func schemaCatalog() *schema.Catalog {
	schemaCatalogOnce.Do(func() {
		schemaCatalogInst = schema.DefaultCatalog()
	})
	return schemaCatalogInst
}
