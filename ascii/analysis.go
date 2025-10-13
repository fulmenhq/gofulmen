package ascii

import (
	"strings"
	"unicode"
)

// AnalyzeString provides analysis of a string's properties
type StringAnalysis struct {
	Length      int
	Width       int
	HasUnicode  bool
	LineCount   int
	WordCount   int
	IsMultiline bool
}

// Analyze performs analysis on the given string
func Analyze(s string) StringAnalysis {
	lines := strings.Split(s, "\n")
	words := strings.Fields(s)

	hasUnicode := false
	for _, r := range s {
		if r > 127 {
			hasUnicode = true
			break
		}
	}

	return StringAnalysis{
		Length:      len(s),
		Width:       StringWidth(s),
		HasUnicode:  hasUnicode,
		LineCount:   len(lines),
		WordCount:   len(words),
		IsMultiline: len(lines) > 1,
	}
}

// IsPrintable checks if a rune is printable
func IsPrintable(r rune) bool {
	return unicode.IsPrint(r)
}

// Sanitize removes non-printable characters
func Sanitize(s string) string {
	var result strings.Builder
	for _, r := range s {
		if IsPrintable(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
