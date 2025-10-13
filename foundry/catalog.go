package foundry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Catalog provides immutable access to Foundry pattern datasets.
//
// The catalog loads patterns, MIME types, and HTTP status groups from
// Crucible configuration using lazy loading for performance. All data
// is cached after first access.
//
// Example:
//
//	catalog := NewCatalog()
//	pattern, _ := catalog.GetPattern("ansi-email")
//	if pattern.MustMatch("user@example.com") {
//	    // Valid email
//	}
type Catalog struct {
	configBasePath string

	// Lazy-loaded data with mutex protection
	patterns     map[string]*Pattern
	patternsOnce sync.Once
	patternsErr  error

	mimeTypes     map[string]*MimeType
	mimeTypesOnce sync.Once
	mimeTypesErr  error

	httpGroups      []*HTTPStatusGroup
	httpGroupsOnce  sync.Once
	httpGroupsErr   error
	httpCodeToGroup map[int]string
	httpHelper      *HTTPStatusHelper
}

// NewCatalog creates a new Catalog instance.
//
// The catalog uses lazy loading - data is only loaded when first accessed.
// It automatically finds the repository root by looking for config/crucible-go.
//
// Example:
//
//	catalog := NewCatalog()
func NewCatalog() *Catalog {
	// Find repository root by looking for config/crucible-go directory
	configPath := findConfigPath()
	return &Catalog{
		configBasePath: configPath,
	}
}

// findConfigPath searches for the config/crucible-go directory starting from the current directory
// and walking up the directory tree.
func findConfigPath() string {
	currentDir, err := os.Getwd()
	if err != nil {
		// Fallback to relative path
		return "config/crucible-go"
	}

	// Try current directory first
	checkPath := filepath.Join(currentDir, "config", "crucible-go")
	if _, err := os.Stat(checkPath); err == nil {
		return filepath.Join(currentDir, "config", "crucible-go")
	}

	// Walk up the directory tree
	for i := 0; i < 10; i++ { // Limit to 10 levels up
		currentDir = filepath.Dir(currentDir)
		if currentDir == "/" || currentDir == "." {
			break
		}

		checkPath = filepath.Join(currentDir, "config", "crucible-go")
		if _, err := os.Stat(checkPath); err == nil {
			return filepath.Join(currentDir, "config", "crucible-go")
		}
	}

	// Fallback to relative path
	return "config/crucible-go"
}

// GetDefaultCatalog returns a singleton catalog.
//
// This is a convenience function for applications that don't need custom
// configuration loading.
//
// Example:
//
//	catalog := GetDefaultCatalog()
//	pattern, _ := catalog.GetPattern("slug")
func GetDefaultCatalog() *Catalog {
	defaultCatalogOnce.Do(func() {
		defaultCatalog = NewCatalog()
	})
	return defaultCatalog
}

var (
	defaultCatalog     *Catalog
	defaultCatalogOnce sync.Once
)

// loadYAML loads a YAML file and returns the parsed data.
func (c *Catalog) loadYAML(relPath string) (map[string]interface{}, error) {
	fullPath := filepath.Join(c.configBasePath, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML from %s: %w", fullPath, err)
	}

	return result, nil
}

// loadPatterns loads patterns from Crucible configuration (lazy loading).
func (c *Catalog) loadPatterns() error {
	c.patternsOnce.Do(func() {
		data, err := c.loadYAML("library/foundry/patterns.yaml")
		if err != nil {
			c.patternsErr = fmt.Errorf("failed to load patterns config: %w", err)
			return
		}

		patternsData, ok := data["patterns"].([]interface{})
		if !ok {
			c.patternsErr = fmt.Errorf("patterns config has invalid format")
			return
		}

		patterns := make(map[string]*Pattern)

		for _, item := range patternsData {
			patternMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			pattern := &Pattern{}

			if id, ok := patternMap["id"].(string); ok {
				pattern.ID = id
			}
			if name, ok := patternMap["name"].(string); ok {
				pattern.Name = name
			}
			if kind, ok := patternMap["kind"].(string); ok {
				pattern.Kind = PatternKind(kind)
			}
			if p, ok := patternMap["pattern"].(string); ok {
				pattern.Pattern = p
			}
			if desc, ok := patternMap["description"].(string); ok {
				pattern.Description = desc
			}

			// Parse examples
			if examples, ok := patternMap["examples"].([]interface{}); ok {
				pattern.Examples = make([]string, 0, len(examples))
				for _, ex := range examples {
					if s, ok := ex.(string); ok {
						pattern.Examples = append(pattern.Examples, s)
					}
				}
			}

			// Parse non_examples
			if nonExamples, ok := patternMap["non_examples"].([]interface{}); ok {
				pattern.NonExamples = make([]string, 0, len(nonExamples))
				for _, ex := range nonExamples {
					if s, ok := ex.(string); ok {
						pattern.NonExamples = append(pattern.NonExamples, s)
					}
				}
			}

			// Parse flags
			if flags, ok := patternMap["flags"].(map[string]interface{}); ok {
				pattern.Flags = make(PatternFlags)
				for lang, langFlags := range flags {
					if flagMap, ok := langFlags.(map[string]interface{}); ok {
						pattern.Flags[lang] = make(map[string]bool)
						for flagName, flagValue := range flagMap {
							if b, ok := flagValue.(bool); ok {
								pattern.Flags[lang][flagName] = b
							}
						}
					}
				}
			}

			if pattern.ID != "" {
				patterns[pattern.ID] = pattern
			}
		}

		c.patterns = patterns
	})

	return c.patternsErr
}

// loadMimeTypes loads MIME types from Crucible configuration (lazy loading).
func (c *Catalog) loadMimeTypes() error {
	c.mimeTypesOnce.Do(func() {
		data, err := c.loadYAML("library/foundry/mime-types.yaml")
		if err != nil {
			c.mimeTypesErr = fmt.Errorf("failed to load mime-types config: %w", err)
			return
		}

		typesData, ok := data["types"].([]interface{})
		if !ok {
			c.mimeTypesErr = fmt.Errorf("mime-types config has invalid format")
			return
		}

		mimeTypes := make(map[string]*MimeType)

		for _, item := range typesData {
			typeMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			mimeType := &MimeType{}

			if id, ok := typeMap["id"].(string); ok {
				mimeType.ID = id
			}
			if mime, ok := typeMap["mime"].(string); ok {
				mimeType.Mime = mime
			}
			if name, ok := typeMap["name"].(string); ok {
				mimeType.Name = name
			}
			if desc, ok := typeMap["description"].(string); ok {
				mimeType.Description = desc
			}

			// Parse extensions
			if extensions, ok := typeMap["extensions"].([]interface{}); ok {
				mimeType.Extensions = make([]string, 0, len(extensions))
				for _, ext := range extensions {
					if s, ok := ext.(string); ok {
						mimeType.Extensions = append(mimeType.Extensions, s)
					}
				}
			}

			if mimeType.ID != "" {
				mimeTypes[mimeType.ID] = mimeType
			}
		}

		c.mimeTypes = mimeTypes
	})

	return c.mimeTypesErr
}

// loadHTTPGroups loads HTTP status groups from Crucible configuration (lazy loading).
func (c *Catalog) loadHTTPGroups() error {
	c.httpGroupsOnce.Do(func() {
		data, err := c.loadYAML("library/foundry/http-statuses.yaml")
		if err != nil {
			c.httpGroupsErr = fmt.Errorf("failed to load http-statuses config: %w", err)
			return
		}

		groupsData, ok := data["groups"].([]interface{})
		if !ok {
			c.httpGroupsErr = fmt.Errorf("http-statuses config has invalid format")
			return
		}

		var groups []*HTTPStatusGroup
		codeToGroup := make(map[int]string)

		for _, item := range groupsData {
			groupMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			group := &HTTPStatusGroup{}

			if id, ok := groupMap["id"].(string); ok {
				group.ID = id
			}
			if name, ok := groupMap["name"].(string); ok {
				group.Name = name
			}
			if desc, ok := groupMap["description"].(string); ok {
				group.Description = desc
			}

			// Parse codes
			if codes, ok := groupMap["codes"].([]interface{}); ok {
				group.Codes = make([]HTTPStatusCode, 0, len(codes))
				for _, codeItem := range codes {
					if codeMap, ok := codeItem.(map[string]interface{}); ok {
						code := HTTPStatusCode{}
						if value, ok := codeMap["value"].(int); ok {
							code.Value = value
						}
						if reason, ok := codeMap["reason"].(string); ok {
							code.Reason = reason
						}
						if code.Value > 0 {
							group.Codes = append(group.Codes, code)
							codeToGroup[code.Value] = group.ID
						}
					}
				}
			}

			if group.ID != "" {
				groups = append(groups, group)
			}
		}

		c.httpGroups = groups
		c.httpCodeToGroup = codeToGroup
		c.httpHelper = NewHTTPStatusHelper(groups)
	})

	return c.httpGroupsErr
}

// GetPattern retrieves a pattern by ID.
//
// Returns nil if the pattern is not found.
//
// Example:
//
//	pattern, err := catalog.GetPattern("ansi-email")
//	if err != nil {
//	    // Handle error
//	}
//	if pattern != nil && pattern.MustMatch("user@example.com") {
//	    // Valid email
//	}
func (c *Catalog) GetPattern(id string) (*Pattern, error) {
	if err := c.loadPatterns(); err != nil {
		return nil, err
	}
	return c.patterns[id], nil
}

// GetAllPatterns returns all available patterns.
//
// Returns a map of pattern ID to Pattern instance.
func (c *Catalog) GetAllPatterns() (map[string]*Pattern, error) {
	if err := c.loadPatterns(); err != nil {
		return nil, err
	}
	// Return a copy to prevent external modification
	result := make(map[string]*Pattern, len(c.patterns))
	for k, v := range c.patterns {
		result[k] = v
	}
	return result, nil
}

// GetMimeType retrieves a MIME type by ID.
//
// Returns nil if the MIME type is not found.
//
// Example:
//
//	mimeType, err := catalog.GetMimeType("json")
//	if err != nil {
//	    // Handle error
//	}
//	if mimeType != nil {
//	    fmt.Println(mimeType.Mime) // "application/json"
//	}
func (c *Catalog) GetMimeType(id string) (*MimeType, error) {
	if err := c.loadMimeTypes(); err != nil {
		return nil, err
	}
	return c.mimeTypes[id], nil
}

// GetMimeTypeByExtension retrieves a MIME type by file extension.
//
// The extension can be provided with or without a leading dot.
// Returns nil if no matching MIME type is found.
//
// Example:
//
//	mimeType, err := catalog.GetMimeTypeByExtension("json")
//	if err != nil {
//	    // Handle error
//	}
//	if mimeType != nil {
//	    fmt.Println(mimeType.Mime) // "application/json"
//	}
func (c *Catalog) GetMimeTypeByExtension(extension string) (*MimeType, error) {
	if err := c.loadMimeTypes(); err != nil {
		return nil, err
	}

	for _, mimeType := range c.mimeTypes {
		if mimeType.MatchesExtension(extension) {
			return mimeType, nil
		}
	}

	return nil, nil
}

// GetAllMimeTypes returns all available MIME types.
//
// Returns a map of MIME type ID to MimeType instance.
func (c *Catalog) GetAllMimeTypes() (map[string]*MimeType, error) {
	if err := c.loadMimeTypes(); err != nil {
		return nil, err
	}
	// Return a copy to prevent external modification
	result := make(map[string]*MimeType, len(c.mimeTypes))
	for k, v := range c.mimeTypes {
		result[k] = v
	}
	return result, nil
}

// GetHTTPStatusGroup retrieves an HTTP status group by ID.
//
// Returns nil if the group is not found.
//
// Example:
//
//	group, err := catalog.GetHTTPStatusGroup("success")
//	if err != nil {
//	    // Handle error
//	}
//	if group != nil && group.Contains(200) {
//	    // Status code is in success group
//	}
func (c *Catalog) GetHTTPStatusGroup(id string) (*HTTPStatusGroup, error) {
	if err := c.loadHTTPGroups(); err != nil {
		return nil, err
	}

	for _, group := range c.httpGroups {
		if group.ID == id {
			return group, nil
		}
	}

	return nil, nil
}

// GetHTTPStatusGroupForCode retrieves the HTTP status group for a specific status code.
//
// Returns nil if the status code is not recognized.
//
// Example:
//
//	group, err := catalog.GetHTTPStatusGroupForCode(404)
//	if err != nil {
//	    // Handle error
//	}
//	if group != nil {
//	    fmt.Println(group.ID) // "client-error"
//	}
func (c *Catalog) GetHTTPStatusGroupForCode(statusCode int) (*HTTPStatusGroup, error) {
	if err := c.loadHTTPGroups(); err != nil {
		return nil, err
	}

	groupID, exists := c.httpCodeToGroup[statusCode]
	if !exists {
		return nil, nil
	}

	return c.GetHTTPStatusGroup(groupID)
}

// GetAllHTTPStatusGroups returns all HTTP status groups.
//
// Returns a slice of HTTPStatusGroup instances.
func (c *Catalog) GetAllHTTPStatusGroups() ([]*HTTPStatusGroup, error) {
	if err := c.loadHTTPGroups(); err != nil {
		return nil, err
	}
	// Return a copy to prevent external modification
	result := make([]*HTTPStatusGroup, len(c.httpGroups))
	copy(result, c.httpGroups)
	return result, nil
}

// GetHTTPStatusHelper returns an HTTP status helper for convenient status code checks.
//
// The helper provides methods like IsSuccess(), IsClientError(), etc.
//
// Example:
//
//	helper, err := catalog.GetHTTPStatusHelper()
//	if err != nil {
//	    // Handle error
//	}
//	if helper.IsSuccess(200) {
//	    // Success response
//	}
func (c *Catalog) GetHTTPStatusHelper() (*HTTPStatusHelper, error) {
	if err := c.loadHTTPGroups(); err != nil {
		return nil, err
	}
	return c.httpHelper, nil
}
