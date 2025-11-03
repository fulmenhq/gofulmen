package foundry

import (
	"github.com/fulmenhq/crucible"
	cruciblefoundry "github.com/fulmenhq/crucible/foundry"
)

// version is the current version of gofulmen.
// This should match the VERSION file in the repository root.
// Updated during release process.
const version = "0.1.8"

// GofulmenVersion returns the gofulmen library version.
// This is sourced from the version constant which should match the VERSION file.
func GofulmenVersion() string {
	return "v" + version
}

// ExitCodesVersion returns the version of the exit codes catalog.
// This is sourced from the Crucible catalog metadata.
func ExitCodesVersion() string {
	return cruciblefoundry.ExitCodesVersion
}

// ExitCodesCatalogVersion is an alias for ExitCodesVersion for consistency with
// other Fulmen helper libraries (pyfulmen, tsfulmen).
func ExitCodesCatalogVersion() string {
	return ExitCodesVersion()
}

// CrucibleVersion returns the Crucible module version that gofulmen is using.
// This allows applications to verify which version of the catalog they're consuming.
func CrucibleVersion() string {
	return crucible.Version
}

// ExitCodesLastReviewed returns the last review date of the exit codes catalog.
// This is sourced from the Crucible foundry package metadata.
func ExitCodesLastReviewed() string {
	// This would ideally come from Crucible's generated constants
	// For now, we parse it from the catalog if needed
	// The Crucible v0.2.3 exit_codes.go has this in a comment
	return "2025-10-31"
}

// ProvInfo returns comprehensive provenance information for the exit codes catalog.
// This is useful for logging startup diagnostics and compliance auditing.
type ProvInfo struct {
	GofulmenVersion       string `json:"gofulmen_version"`
	CrucibleVersion       string `json:"crucible_version"`
	ExitCodesVersion      string `json:"exit_codes_version"`
	ExitCodesLastReviewed string `json:"exit_codes_last_reviewed,omitempty"`
}

// GetProvenanceInfo returns comprehensive provenance information for diagnostics and auditing.
//
// Example usage:
//
//	prov := foundry.GetProvenanceInfo()
//	logger.Info("exit codes provenance",
//	    zap.String("gofulmen_version", prov.GofulmenVersion),
//	    zap.String("crucible_version", prov.CrucibleVersion),
//	    zap.String("catalog_version", prov.ExitCodesVersion))
func GetProvenanceInfo() ProvInfo {
	return ProvInfo{
		GofulmenVersion:       GofulmenVersion(),
		CrucibleVersion:       CrucibleVersion(),
		ExitCodesVersion:      ExitCodesVersion(),
		ExitCodesLastReviewed: ExitCodesLastReviewed(),
	}
}
