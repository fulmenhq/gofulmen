package foundry

import (
	"path/filepath"
	"strings"
)

// MimeType represents an immutable MIME type definition from Foundry catalog.
//
// MIME types map file extensions to standard MIME type strings (e.g.,
// "application/json" for .json files). These are loaded from Crucible
// configuration to ensure consistent MIME type handling across services.
type MimeType struct {
	// ID is the unique MIME type identifier (e.g., "json", "yaml").
	ID string

	// Mime is the MIME type string (e.g., "application/json").
	Mime string

	// Name is the human-readable name (e.g., "JSON").
	Name string

	// Extensions contains file extensions for this MIME type (without leading dots).
	Extensions []string

	// Description provides documentation for this MIME type.
	Description string
}

// MatchesExtension checks if the given file extension matches this MIME type.
//
// The extension can be provided with or without a leading dot.
// Matching is case-insensitive.
//
// Example:
//
//	mimeType := &MimeType{Extensions: []string{"json", "map"}}
//	if mimeType.MatchesExtension("json") {  // true
//	    // Matched
//	}
//	if mimeType.MatchesExtension(".JSON") { // also true (case-insensitive)
//	    // Matched
//	}
func (m *MimeType) MatchesExtension(extension string) bool {
	ext := strings.ToLower(strings.TrimPrefix(extension, "."))
	for _, e := range m.Extensions {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

// MatchesFilename checks if the given filename's extension matches this MIME type.
//
// This is a convenience method that extracts the extension from a filename
// and calls MatchesExtension.
//
// Example:
//
//	mimeType := &MimeType{Extensions: []string{"json"}}
//	if mimeType.MatchesFilename("config.json") {
//	    // Matched
//	}
func (m *MimeType) MatchesFilename(filename string) bool {
	ext := filepath.Ext(filename)
	return m.MatchesExtension(ext)
}

// GetPrimaryExtension returns the first (primary) extension for this MIME type.
//
// Returns an empty string if the MIME type has no extensions defined.
//
// Example:
//
//	mimeType := &MimeType{Extensions: []string{"yaml", "yml"}}
//	primary := mimeType.GetPrimaryExtension() // Returns "yaml"
func (m *MimeType) GetPrimaryExtension() string {
	if len(m.Extensions) == 0 {
		return ""
	}
	return m.Extensions[0]
}
