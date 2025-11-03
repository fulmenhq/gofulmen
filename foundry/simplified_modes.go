package foundry

// SimplifiedMode represents a simplified exit code mapping mode.
// Simplified modes collapse the full exit code space into a smaller set
// for novice users or environments with limited exit code support.
type SimplifiedMode string

const (
	// SimplifiedModeBasic provides a minimal 3-value exit code set:
	// 0 (success), 1 (general failure), 2 (usage error).
	// Use this for environments that only support basic exit codes.
	SimplifiedModeBasic SimplifiedMode = "basic"

	// SimplifiedModeSeverity provides a 5-value severity-based exit code set:
	// 0 (success), 1 (info), 2 (warning), 3 (error), 4 (critical).
	// Use this for environments that can distinguish severity levels.
	SimplifiedModeSeverity SimplifiedMode = "severity"
)

// MapToSimplified maps a Fulmen exit code to a simplified mode code.
// Returns (mappedCode, true) if a mapping exists, (0, false) if not.
//
// Example usage:
//
//	if code, ok := foundry.MapToSimplified(foundry.ExitConfigInvalid, foundry.SimplifiedModeBasic); ok {
//	    os.Exit(code)
//	}
func MapToSimplified(code ExitCode, mode SimplifiedMode) (int, bool) {
	info, found := GetExitCodeInfo(code)
	if !found {
		return 0, false
	}

	switch mode {
	case SimplifiedModeBasic:
		if info.SimplifiedBasic != nil {
			return *info.SimplifiedBasic, true
		}
	case SimplifiedModeSeverity:
		if info.SimplifiedSeverity != nil {
			return *info.SimplifiedSeverity, true
		}
	}

	return 0, false
}

// ListSimplifiedModes returns all available simplified mode identifiers.
// This is useful for documentation and validation.
func ListSimplifiedModes() []SimplifiedMode {
	return []SimplifiedMode{
		SimplifiedModeBasic,
		SimplifiedModeSeverity,
	}
}

// GetSimplifiedModeInfo returns metadata about a simplified mode.
// Returns nil if the mode is not recognized.
// This derives information from the catalog to stay in sync with Crucible.
func GetSimplifiedModeInfo(mode SimplifiedMode) *SimplifiedModeInfo {
	// Find the mode in the catalog
	for _, catalogMode := range catalog.SimplifiedModes {
		if catalogMode.ID != string(mode) {
			continue
		}

		// Build the code info list from mappings
		codes := make([]SimplifiedCodeInfo, 0, len(catalogMode.Mappings))
		for _, mapping := range catalogMode.Mappings {
			codes = append(codes, SimplifiedCodeInfo{
				Code:        mapping.SimplifiedCode,
				Name:        mapping.SimplifiedName,
				Description: buildSimplifiedDescription(mapping.SimplifiedName),
			})
		}

		return &SimplifiedModeInfo{
			ID:          catalogMode.ID,
			Name:        catalogMode.Name,
			Description: catalogMode.Description,
			Codes:       codes,
		}
	}
	return nil
}

// buildSimplifiedDescription generates a description for a simplified code name.
func buildSimplifiedDescription(name string) string {
	descriptions := map[string]string{
		"SUCCESS":             "Operation completed successfully",
		"ERROR":               "General error or failure",
		"USAGE_ERROR":         "Command usage error or invalid arguments",
		"USER_ERROR":          "User input or argument error",
		"CONFIG_ERROR":        "Configuration or validation error",
		"RUNTIME_ERROR":       "Runtime operation error",
		"SYSTEM_ERROR":        "System resource or permission error",
		"SECURITY_ERROR":      "Security or authentication error",
		"TEST_FAILURE":        "Test execution failure",
		"OBSERVABILITY_ERROR": "Observability infrastructure error",
	}

	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return name // Fallback to name if no description
}

// SimplifiedModeInfo provides metadata about a simplified mode.
type SimplifiedModeInfo struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Codes       []SimplifiedCodeInfo `json:"codes"`
}

// SimplifiedCodeInfo provides metadata about a simplified exit code.
type SimplifiedCodeInfo struct {
	Code        int    `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
