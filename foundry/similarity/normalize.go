package similarity

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizeOptions configures the text normalization behavior.
//
// Normalization applies a pipeline of transformations:
//  1. Trim leading/trailing whitespace
//  2. Apply Unicode case folding (locale-aware)
//  3. Optionally strip diacritical marks (accents)
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
type NormalizeOptions struct {
	// StripAccents controls whether to remove diacritical marks.
	// Default: false (preserve accents).
	//
	// When true, applies Unicode NFD decomposition, filters combining marks
	// (category Mn), and recomposes to NFC.
	//
	// Examples with StripAccents=true:
	//   - "café" → "cafe"
	//   - "naïve" → "naive"
	//   - "Zürich" → "zurich"
	StripAccents bool

	// Locale specifies the locale for case folding.
	// Default: "" (simple Unicode case folding).
	//
	// Supported locales:
	//   - "" (empty): Simple Unicode case folding (default)
	//   - "tr" or "TR": Turkish locale (handles İ→i, I→ı correctly)
	//
	// Note: Limited locale support in initial implementation. Additional
	// locales may be added in future versions.
	Locale string
}

// Normalize applies the full normalization pipeline to a string.
//
// The pipeline performs these operations in order:
//  1. Trim leading and trailing whitespace
//  2. Apply case folding (using opts.Locale)
//  3. Optionally strip accents (if opts.StripAccents is true)
//
// Examples:
//   - Normalize("  Hello  ", NormalizeOptions{}) returns "hello"
//   - Normalize("Café", NormalizeOptions{StripAccents: true}) returns "cafe"
//   - Normalize("İstanbul", NormalizeOptions{Locale: "tr"}) returns "istanbul"
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func Normalize(value string, opts NormalizeOptions) string {
	// Step 1: Trim leading/trailing whitespace
	result := strings.TrimSpace(value)

	// Step 2: Apply case folding
	result = Casefold(result, opts.Locale)

	// Step 3: Optionally strip accents
	if opts.StripAccents {
		result = StripAccents(result)
	}

	return result
}

// Casefold converts a string to lowercase using Unicode case folding.
//
// Case folding is more comprehensive than simple lowercasing and handles
// locale-specific rules when a locale is specified.
//
// Parameters:
//   - value: The string to case fold
//   - locale: Locale for case folding ("" for simple, "tr" for Turkish, etc.)
//
// Examples:
//   - Casefold("Hello", "") returns "hello"
//   - Casefold("İstanbul", "tr") returns "istanbul" (Turkish dotted I)
//   - Casefold("TITLE", "tr") returns "tıtle" (Turkish dotless ı)
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func Casefold(value string, locale string) string {
	// Handle Turkish locale special cases
	if locale == "tr" || locale == "TR" {
		return turkishCasefold(value)
	}

	// Default: Simple Unicode case folding
	return strings.ToLower(value)
}

// turkishCasefold applies Turkish-specific case folding rules.
//
// Turkish has special case mapping rules:
//   - İ (U+0130 Latin Capital Letter I with Dot) → i
//   - I (U+0049 Latin Capital Letter I) → ı (U+0131 Latin Small Letter Dotless I)
func turkishCasefold(value string) string {
	var builder strings.Builder
	builder.Grow(len(value)) // Pre-allocate for efficiency

	for _, r := range value {
		switch r {
		case 'İ': // U+0130 - Turkish dotted capital I
			builder.WriteRune('i') // Lowercase i
		case 'I': // U+0049 - ASCII capital I
			builder.WriteRune('ı') // U+0131 - Turkish dotless lowercase ı
		default:
			// Use standard Unicode lowercasing for other characters
			builder.WriteRune(unicode.ToLower(r))
		}
	}

	return builder.String()
}

// StripAccents removes diacritical marks from a string.
//
// Uses Unicode normalization to decompose characters into base characters
// and combining marks, filters out combining marks (category Mn), and
// recomposes the result.
//
// Algorithm:
//  1. Decompose to NFD (Normalization Form Decomposed)
//  2. Filter out characters in Unicode category Mn (Nonspacing_Mark)
//  3. Recompose to NFC (Normalization Form Composed)
//
// Examples:
//   - StripAccents("café") returns "cafe"
//   - StripAccents("naïve") returns "naive"
//   - StripAccents("Zürich") returns "Zurich"
//   - StripAccents("résumé") returns "resume"
//
// Note: Uses golang.org/x/text/unicode/norm package (approved).
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func StripAccents(value string) string {
	// Step 1: Decompose to NFD (separate base chars from combining marks)
	decomposed := norm.NFD.String(value)

	// Step 2: Filter out combining diacritical marks (Unicode category Mn)
	var builder strings.Builder
	builder.Grow(len(decomposed)) // Pre-allocate

	for _, r := range decomposed {
		// Keep character if it's NOT a nonspacing mark (category Mn)
		if !unicode.Is(unicode.Mn, r) {
			builder.WriteRune(r)
		}
	}

	// Step 3: Recompose to NFC
	return norm.NFC.String(builder.String())
}

// EqualsIgnoreCase compares two strings for equality using normalization.
//
// Both strings are normalized using the provided options before comparison.
// This enables case-insensitive and optionally accent-insensitive comparison.
//
// Examples:
//   - EqualsIgnoreCase("Hello", "hello", NormalizeOptions{}) returns true
//   - EqualsIgnoreCase("Café", "cafe", NormalizeOptions{StripAccents: true}) returns true
//   - EqualsIgnoreCase("Hello", "World", NormalizeOptions{}) returns false
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func EqualsIgnoreCase(a, b string, opts NormalizeOptions) bool {
	return Normalize(a, opts) == Normalize(b, opts)
}
