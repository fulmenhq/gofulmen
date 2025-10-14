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

// DetectMimeType inspects raw bytes and returns the matching MIME type from the catalog.
//
// This function performs basic content detection based on common file signatures
// (magic numbers). Returns nil if the content cannot be identified.
//
// Example:
//
//	data := []byte(`{"key": "value"}`)
//	mimeType, err := DetectMimeType(data)
//	if err != nil {
//	    // Handle error
//	}
//	if mimeType != nil {
//	    fmt.Println(mimeType.Mime) // "application/json"
//	}
func DetectMimeType(input []byte) (*MimeType, error) {
	catalog := GetDefaultCatalog()
	if err := catalog.loadMimeTypes(); err != nil {
		return nil, err
	}

	// Basic content-based detection using magic numbers and patterns
	if len(input) == 0 {
		return nil, nil
	}

	// Trim BOM and leading whitespace for accurate signature detection
	trimmed := trimBOMAndWhitespace(input)
	if len(trimmed) == 0 {
		return nil, nil
	}

	// Check for common file signatures
	// JSON: starts with { or [
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		// Validate it's actually JSON-like
		for _, b := range trimmed[:min(len(trimmed), 50)] {
			if b == '{' || b == '[' || b == '"' || b == ':' {
				return catalog.mimeTypes["json"], nil
			}
		}
	}

	// XML: starts with <
	if len(trimmed) > 0 && trimmed[0] == '<' {
		if len(trimmed) > 5 && string(trimmed[:5]) == "<?xml" {
			return catalog.mimeTypes["xml"], nil
		}
	}

	// YAML: look for YAML-specific patterns
	lines := string(trimmed[:min(len(trimmed), 200)])
	if len(lines) > 0 {
		// Simple heuristic: if it has key: value patterns and no { or <
		hasColon := false
		for i := 0; i < len(lines)-1; i++ {
			if lines[i] == ':' && (lines[i+1] == ' ' || lines[i+1] == '\n') {
				hasColon = true
				break
			}
		}
		if hasColon && trimmed[0] != '{' && trimmed[0] != '[' && trimmed[0] != '<' {
			return catalog.mimeTypes["yaml"], nil
		}
	}

	// CSV: look for comma-separated values
	firstLine := string(input[:min(len(input), 200)])
	for idx := 0; idx < len(firstLine); idx++ {
		if firstLine[idx] == '\n' {
			firstLine = firstLine[:idx]
			break
		}
	}
	if len(firstLine) > 0 && countCommas(firstLine) >= 2 {
		return catalog.mimeTypes["csv"], nil
	}

	// Protocol Buffers: binary format (hard to detect reliably)
	// Skip for now as it requires more sophisticated detection

	// Plain text: fallback for text-like content
	if isTextContent(input[:min(len(input), 512)]) {
		return catalog.mimeTypes["plain-text"], nil
	}

	return nil, nil
}

// IsSupportedMimeType checks if the given MIME string exists in the catalog.
//
// Example:
//
//	if IsSupportedMimeType("application/json") {
//	    // MIME type is supported
//	}
func IsSupportedMimeType(mime string) bool {
	catalog := GetDefaultCatalog()
	if err := catalog.loadMimeTypes(); err != nil {
		return false
	}

	for _, mimeType := range catalog.mimeTypes {
		if mimeType.Mime == mime {
			return true
		}
	}

	return false
}

// GetMimeTypeByExtension retrieves a MIME type by file extension.
//
// The extension can be provided with or without a leading dot.
// Returns nil if no matching MIME type is found.
//
// Example:
//
//	mimeType, err := GetMimeTypeByExtension("json")
//	if err != nil {
//	    // Handle error
//	}
//	if mimeType != nil {
//	    fmt.Println(mimeType.Mime) // "application/json"
//	}
func GetMimeTypeByExtension(extension string) (*MimeType, error) {
	catalog := GetDefaultCatalog()
	return catalog.GetMimeTypeByExtension(extension)
}

// ListMimeTypes returns all MIME types from the catalog.
//
// Example:
//
//	mimeTypes, err := ListMimeTypes()
//	if err != nil {
//	    // Handle error
//	}
//	for _, mimeType := range mimeTypes {
//	    fmt.Printf("%s: %s\n", mimeType.ID, mimeType.Mime)
//	}
func ListMimeTypes() ([]*MimeType, error) {
	catalog := GetDefaultCatalog()
	if err := catalog.loadMimeTypes(); err != nil {
		return nil, err
	}

	// Convert map to slice
	result := make([]*MimeType, 0, len(catalog.mimeTypes))
	for _, mimeType := range catalog.mimeTypes {
		result = append(result, mimeType)
	}

	return result, nil
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func countCommas(s string) int {
	count := 0
	for _, c := range s {
		if c == ',' {
			count++
		}
	}
	return count
}

func isTextContent(data []byte) bool {
	// Simple heuristic: check if most bytes are printable ASCII or common UTF-8
	printableCount := 0
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			printableCount++
		} else if b >= 128 {
			// UTF-8 continuation or multi-byte character
			printableCount++
		}
	}

	// If more than 80% is printable, consider it text
	return len(data) > 0 && float64(printableCount)/float64(len(data)) > 0.8
}

// trimBOMAndWhitespace removes byte order marks (BOM) and leading whitespace.
//
// This is critical for accurate MIME detection since real-world JSON/XML files
// often start with BOM (UTF-8: EF BB BF, UTF-16: FF FE or FE FF) or whitespace.
func trimBOMAndWhitespace(data []byte) []byte {
	// Remove UTF-8 BOM (EF BB BF)
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	// Remove UTF-16 BOM (FF FE or FE FF)
	if len(data) >= 2 && ((data[0] == 0xFF && data[1] == 0xFE) || (data[0] == 0xFE && data[1] == 0xFF)) {
		data = data[2:]
	}

	// Trim leading whitespace (space, tab, newline, carriage return)
	start := 0
	for start < len(data) {
		b := data[start]
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			start++
		} else {
			break
		}
	}

	return data[start:]
}
