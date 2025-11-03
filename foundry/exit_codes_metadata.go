package foundry

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/fulmenhq/crucible"
	"gopkg.in/yaml.v3"
)

// ExitCodeInfo provides comprehensive metadata for an exit code from the Foundry catalog.
type ExitCodeInfo struct {
	Code               int    `json:"code"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Context            string `json:"context"`
	Category           string `json:"category"`
	RetryHint          string `json:"retry_hint,omitempty"`
	BSDEquivalent      string `json:"bsd_equivalent,omitempty"`
	SimplifiedBasic    *int   `json:"simplified_basic,omitempty"`
	SimplifiedSeverity *int   `json:"simplified_severity,omitempty"`
}

// catalogData represents the structure of exit-codes.yaml
type catalogData struct {
	Schema           string            `yaml:"$schema"`
	Description      string            `yaml:"description"`
	Version          string            `yaml:"version"`
	Categories       []category        `yaml:"categories"`
	SimplifiedModes  []simplifiedMode  `yaml:"simplified_modes"`
	BSDCompatibility *bsdCompatibility `yaml:"bsd_compatibility,omitempty"`
}

type category struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Range       codeRange   `yaml:"range"`
	Codes       []codeEntry `yaml:"codes"`
}

type codeRange struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type codeEntry struct {
	Code          int    `yaml:"code"`
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	Context       string `yaml:"context"`
	RetryHint     string `yaml:"retry_hint,omitempty"`
	BSDEquivalent string `yaml:"bsd_equivalent,omitempty"`
}

type simplifiedMode struct {
	ID          string              `yaml:"id"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Mappings    []simplifiedMapping `yaml:"mappings"`
}

type simplifiedMapping struct {
	SimplifiedCode int    `yaml:"simplified_code"`
	SimplifiedName string `yaml:"simplified_name"`
	MapsFrom       []int  `yaml:"maps_from"`
}

type bsdCompatibility struct {
	Purpose  string       `yaml:"purpose"`
	Mappings []bsdMapping `yaml:"mappings"`
}

type bsdMapping struct {
	BSDCode    int    `yaml:"bsd_code"`
	BSDName    string `yaml:"bsd_name"`
	FulmenCode int    `yaml:"fulmen_code"`
	FulmenName string `yaml:"fulmen_name"`
	Category   string `yaml:"category"`
	Notes      string `yaml:"notes,omitempty"`
}

var (
	catalog *catalogData
	once    sync.Once
	loadErr error

	// Precomputed maps for fast lookups
	codeInfoMap map[int]*ExitCodeInfo
	nameInfoMap map[string]*ExitCodeInfo
)

// init loads and parses the exit codes catalog from Crucible's embedded assets.
// Panics if catalog loading fails, as exit codes are fundamental infrastructure.
func init() {
	loadCatalog()
}

// loadCatalog reads and parses the exit-codes.yaml from Crucible's embedded ConfigFS.
func loadCatalog() {
	once.Do(func() {
		// Read the catalog file using Crucible's GetConfig
		data, err := crucible.GetConfig("library/foundry/exit-codes.yaml")
		if err != nil {
			loadErr = fmt.Errorf("failed to read exit-codes catalog from Crucible: %w (crucible version: %s, gofulmen: %s, GOOS: %s, GOARCH: %s)",
				err, crucibleVersion(), gofulmenVersion(), runtime.GOOS, runtime.GOARCH)
			panic(loadErr)
		}

		// Parse YAML
		var cat catalogData
		if err := yaml.Unmarshal(data, &cat); err != nil {
			loadErr = fmt.Errorf("failed to parse exit-codes catalog YAML: %w (crucible version: %s, gofulmen: %s)",
				err, crucibleVersion(), gofulmenVersion())
			panic(loadErr)
		}

		catalog = &cat

		// Build lookup maps
		buildLookupMaps()
	})
}

// buildLookupMaps constructs efficient lookup structures from the parsed catalog.
func buildLookupMaps() {
	codeInfoMap = make(map[int]*ExitCodeInfo)
	nameInfoMap = make(map[string]*ExitCodeInfo)

	// Build simplified mode mappings first
	basicMappings := make(map[int]int)    // fulmen -> basic
	severityMappings := make(map[int]int) // fulmen -> severity

	for _, mode := range catalog.SimplifiedModes {
		for _, mapping := range mode.Mappings {
			for _, fulmenCode := range mapping.MapsFrom {
				switch mode.ID {
				case "basic":
					basicMappings[fulmenCode] = mapping.SimplifiedCode
				case "severity":
					severityMappings[fulmenCode] = mapping.SimplifiedCode
				}
			}
		}
	}

	// Process all categories and codes
	for _, cat := range catalog.Categories {
		for _, code := range cat.Codes {
			info := &ExitCodeInfo{
				Code:          code.Code,
				Name:          code.Name,
				Description:   code.Description,
				Context:       code.Context,
				Category:      cat.ID,
				RetryHint:     code.RetryHint,
				BSDEquivalent: code.BSDEquivalent,
			}

			// Add simplified mappings if they exist
			if basicCode, ok := basicMappings[code.Code]; ok {
				info.SimplifiedBasic = &basicCode
			}
			if severityCode, ok := severityMappings[code.Code]; ok {
				info.SimplifiedSeverity = &severityCode
			}

			codeInfoMap[code.Code] = info
			nameInfoMap[code.Name] = info
		}
	}
}

// GetExitCodeInfo returns metadata for the specified exit code.
// Returns (info, true) if found, (zero, false) if not in catalog.
func GetExitCodeInfo(code ExitCode) (ExitCodeInfo, bool) {
	info, ok := codeInfoMap[code]
	if !ok {
		return ExitCodeInfo{}, false
	}
	return *info, true
}

// LookupExitCode returns metadata for the specified exit code name (e.g., "EXIT_SUCCESS").
// Returns (info, true) if found, (zero, false) if not in catalog.
func LookupExitCode(name string) (ExitCodeInfo, bool) {
	info, ok := nameInfoMap[name]
	if !ok {
		return ExitCodeInfo{}, false
	}
	return *info, true
}

// ListExitCodes returns all exit codes from the catalog, sorted by code number.
// Use ListOptions to filter results.
func ListExitCodes(opts ...ListOption) []ExitCodeInfo {
	config := &listConfig{}
	for _, opt := range opts {
		opt(config)
	}

	codes := make([]int, 0, len(codeInfoMap))
	for code := range codeInfoMap {
		codes = append(codes, code)
	}

	// Sort codes numerically
	sort.Ints(codes)

	result := make([]ExitCodeInfo, 0, len(codes))
	for _, code := range codes {
		info := *codeInfoMap[code]

		// Apply filters
		if config.excludeSignalCodes && code >= 128 && code <= 165 {
			continue
		}

		result = append(result, info)
	}

	return result
}

// ListOption configures how ListExitCodes filters results.
type ListOption func(*listConfig)

type listConfig struct {
	excludeSignalCodes bool
}

// WithoutSignalCodes excludes POSIX signal exit codes (128-165) from the results.
// Useful on Windows where signal codes are not supported.
func WithoutSignalCodes() ListOption {
	return func(c *listConfig) {
		c.excludeSignalCodes = true
	}
}

// Helper functions for version reporting
func crucibleVersion() string {
	return CrucibleVersion()
}

func gofulmenVersion() string {
	return GofulmenVersion()
}
