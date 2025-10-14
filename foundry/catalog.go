package foundry

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Embed YAML catalogs from assets/ (synced from Crucible SSOT via make sync)
//
//go:embed assets/*.yaml
var configFiles embed.FS

// Catalog provides immutable access to Foundry pattern datasets.
//
// The catalog loads patterns, MIME types, and HTTP status groups from
// embedded Crucible configuration using lazy loading for performance.
// All data is cached after first access and works offline in compiled binaries.
//
// Example:
//
//	catalog := NewCatalog()
//	pattern, _ := catalog.GetPattern("ansi-email")
//	if pattern.MustMatch("user@example.com") {
//	    // Valid email
//	}
type Catalog struct {
	// Lazy-loaded data with mutex protection
	patterns     map[string]*Pattern
	patternsOnce sync.Once
	patternsErr  error

	mimeTypes     map[string]*MimeType
	mimeTypesOnce sync.Once
	mimeTypesErr  error

	countries        map[string]*Country // keyed by uppercase Alpha2
	countriesAlpha3  map[string]*Country // keyed by uppercase Alpha3
	countriesNumeric map[string]*Country // keyed by zero-padded numeric (e.g., "840")
	countriesOnce    sync.Once
	countriesErr     error

	httpGroups      []*HTTPStatusGroup
	httpGroupsOnce  sync.Once
	httpGroupsErr   error
	httpCodeToGroup map[int]string
	httpHelper      *HTTPStatusHelper
}

// NewCatalog creates a new Catalog instance.
//
// The catalog uses lazy loading - data is only loaded when first accessed.
// All configuration files are embedded at compile time for offline operation.
//
// Example:
//
//	catalog := NewCatalog()
func NewCatalog() *Catalog {
	return &Catalog{}
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

// loadYAML loads a YAML file from embedded files and returns the parsed data.
func (c *Catalog) loadYAML(filename string) (map[string]interface{}, error) {
	// Use forward slash for embed.FS paths (Windows-safe)
	// Assets are synced from Crucible SSOT via make sync
	fullPath := "assets/" + filename
	data, err := configFiles.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded file %s: %w", fullPath, err)
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
		data, err := c.loadYAML("patterns.yaml")
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
		data, err := c.loadYAML("mime-types.yaml")
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
		data, err := c.loadYAML("http-statuses.yaml")
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

// loadCountries loads countries from Crucible configuration (lazy loading).
//
// Builds three indexes for efficient lookup:
// - Alpha2 (uppercase, e.g., "US")
// - Alpha3 (uppercase, e.g., "USA")
// - Numeric (zero-padded to 3 digits, e.g., "840")
func (c *Catalog) loadCountries() error {
	c.countriesOnce.Do(func() {
		data, err := c.loadYAML("country-codes.yaml")
		if err != nil {
			c.countriesErr = fmt.Errorf("failed to load country-codes config: %w", err)
			return
		}

		countriesData, ok := data["countries"].([]interface{})
		if !ok {
			c.countriesErr = fmt.Errorf("country-codes config has invalid format")
			return
		}

		countries := make(map[string]*Country)
		countriesAlpha3 := make(map[string]*Country)
		countriesNumeric := make(map[string]*Country)

		for _, item := range countriesData {
			countryMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			country := &Country{}

			if alpha2, ok := countryMap["alpha2"].(string); ok {
				country.Alpha2 = alpha2
			}
			if alpha3, ok := countryMap["alpha3"].(string); ok {
				country.Alpha3 = alpha3
			}
			if numeric, ok := countryMap["numeric"].(string); ok {
				country.Numeric = numeric
			}
			if name, ok := countryMap["name"].(string); ok {
				country.Name = name
			}
			if officialName, ok := countryMap["officialName"].(string); ok {
				country.OfficialName = officialName
			}

			// Build primary index (Alpha2, uppercase)
			if country.Alpha2 != "" {
				normalizedAlpha2 := strings.ToUpper(country.Alpha2)
				countries[normalizedAlpha2] = country
			}

			// Build secondary index (Alpha3, uppercase)
			if country.Alpha3 != "" {
				normalizedAlpha3 := strings.ToUpper(country.Alpha3)
				countriesAlpha3[normalizedAlpha3] = country
			}

			// Build tertiary index (Numeric, zero-padded to 3 digits)
			if country.Numeric != "" {
				// Ensure numeric code is zero-padded to 3 digits
				numericCode := country.Numeric
				for len(numericCode) < 3 {
					numericCode = "0" + numericCode
				}
				countriesNumeric[numericCode] = country
			}
		}

		c.countries = countries
		c.countriesAlpha3 = countriesAlpha3
		c.countriesNumeric = countriesNumeric
	})

	return c.countriesErr
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

// GetCountry retrieves a country by its Alpha2 code.
//
// The code is normalized to uppercase for case-insensitive lookup.
// Returns nil if the country is not found.
//
// Example:
//
//	country, err := catalog.GetCountry("US")    // works
//	country, err := catalog.GetCountry("us")    // also works
//	if err != nil {
//	    // Handle error
//	}
//	if country != nil {
//	    fmt.Println(country.Name) // "United States of America"
//	}
func (c *Catalog) GetCountry(alpha2 string) (*Country, error) {
	if err := c.loadCountries(); err != nil {
		return nil, err
	}
	normalizedAlpha2 := strings.ToUpper(alpha2)
	return c.countries[normalizedAlpha2], nil
}

// GetCountryByAlpha3 retrieves a country by its Alpha3 code.
//
// The code is normalized to uppercase for case-insensitive lookup.
// Returns nil if the country is not found.
//
// Example:
//
//	country, err := catalog.GetCountryByAlpha3("USA")  // works
//	country, err := catalog.GetCountryByAlpha3("usa")  // also works
//	if err != nil {
//	    // Handle error
//	}
//	if country != nil {
//	    fmt.Println(country.Name) // "United States of America"
//	}
func (c *Catalog) GetCountryByAlpha3(alpha3 string) (*Country, error) {
	if err := c.loadCountries(); err != nil {
		return nil, err
	}
	normalizedAlpha3 := strings.ToUpper(alpha3)
	return c.countriesAlpha3[normalizedAlpha3], nil
}

// ListCountries returns all countries from the catalog.
//
// Returns a slice of Country instances.
//
// Example:
//
//	countries, err := catalog.ListCountries()
//	if err != nil {
//	    // Handle error
//	}
//	for _, country := range countries {
//	    fmt.Printf("%s: %s\n", country.Alpha2, country.Name)
//	}
func (c *Catalog) ListCountries() ([]*Country, error) {
	if err := c.loadCountries(); err != nil {
		return nil, err
	}

	// Convert map to slice
	result := make([]*Country, 0, len(c.countries))
	for _, country := range c.countries {
		result = append(result, country)
	}

	return result, nil
}
