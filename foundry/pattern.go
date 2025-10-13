package foundry

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
)

// PatternKind defines the type of pattern matching to use.
type PatternKind string

const (
	// PatternKindRegex indicates a regular expression pattern.
	PatternKindRegex PatternKind = "regex"

	// PatternKindGlob indicates a glob/wildcard pattern (e.g., "*.json").
	PatternKindGlob PatternKind = "glob"

	// PatternKindLiteral indicates an exact string match.
	PatternKindLiteral PatternKind = "literal"
)

// PatternFlags contains language-specific regex flags from Crucible catalog.
//
// The flags map keys are language names (e.g., "go", "python", "typescript")
// and values are flag configurations specific to that language.
type PatternFlags map[string]map[string]bool

// Pattern represents an immutable pattern definition from Foundry catalog.
//
// Patterns support three kinds of matching:
//   - regex: Regular expression matching with language-specific flags
//   - glob: File glob/wildcard matching (e.g., "**/*.json")
//   - literal: Exact string matching
//
// Patterns are loaded from Crucible configuration and provide compiled
// regex instances with lazy initialization for performance.
type Pattern struct {
	// ID is the unique pattern identifier (e.g., "ansi-email", "slug").
	ID string

	// Name is the human-readable pattern name.
	Name string

	// Kind indicates the pattern type: regex, glob, or literal.
	Kind PatternKind

	// Pattern is the pattern string (regex, glob pattern, or literal string).
	Pattern string

	// Description provides documentation for the pattern's purpose.
	Description string

	// Examples contains valid example strings that match this pattern.
	Examples []string

	// NonExamples contains invalid example strings that should not match.
	NonExamples []string

	// Flags contains language-specific regex flags (e.g., ignoreCase, unicode).
	Flags PatternFlags

	// compiledRegex is the lazily-compiled regex (for regex kind only).
	compiledRegex *regexp.Regexp

	// compileOnce ensures regex is compiled only once.
	compileOnce sync.Once

	// compileErr stores any error from regex compilation.
	compileErr error
}

// Match tests if the given value matches this pattern.
//
// The matching behavior depends on the pattern kind:
//   - regex: Full match from start to end (like regexp.MatchString)
//   - glob: Filename pattern match using filepath.Match
//   - literal: Exact string equality
//
// Returns true if value matches the pattern.
//
// Example:
//
//	pattern := &Pattern{Kind: PatternKindRegex, Pattern: "^[a-z]+$"}
//	if pattern.Match("hello") {
//	    // Matched
//	}
func (p *Pattern) Match(value string) (bool, error) {
	switch p.Kind {
	case PatternKindRegex:
		regex, err := p.CompiledRegex()
		if err != nil {
			return false, fmt.Errorf("pattern compilation failed: %w", err)
		}
		return regex.MatchString(value), nil

	case PatternKindLiteral:
		return value == p.Pattern, nil

	case PatternKindGlob:
		matched, err := filepath.Match(p.Pattern, value)
		if err != nil {
			return false, fmt.Errorf("glob match failed: %w", err)
		}
		return matched, nil

	default:
		return false, fmt.Errorf("unknown pattern kind: %s", p.Kind)
	}
}

// MustMatch is like Match but panics if an error occurs.
//
// This is useful when you know the pattern is valid (e.g., loaded from
// validated Crucible configuration) and want simpler error handling.
//
// Example:
//
//	if emailPattern.MustMatch("user@example.com") {
//	    // Valid email
//	}
func (p *Pattern) MustMatch(value string) bool {
	matched, err := p.Match(value)
	if err != nil {
		panic(fmt.Sprintf("pattern match failed: %v", err))
	}
	return matched
}

// Search checks if the pattern appears anywhere in the value (not just full match).
//
// For regex patterns, this uses FindString instead of MatchString.
// For literal patterns, this uses string containment.
// For glob patterns, this falls back to Match (globs don't have "search" semantic).
//
// Example:
//
//	pattern := &Pattern{Kind: PatternKindRegex, Pattern: `\d+`}
//	if pattern.Search("Order #12345 shipped") {
//	    // Found digits in the string
//	}
func (p *Pattern) Search(value string) (bool, error) {
	switch p.Kind {
	case PatternKindRegex:
		regex, err := p.CompiledRegex()
		if err != nil {
			return false, fmt.Errorf("pattern compilation failed: %w", err)
		}
		return regex.FindString(value) != "", nil

	case PatternKindLiteral:
		return contains(value, p.Pattern), nil

	case PatternKindGlob:
		// Glob doesn't have "search" semantic - use match
		return p.Match(value)

	default:
		return false, fmt.Errorf("unknown pattern kind: %s", p.Kind)
	}
}

// CompiledRegex returns the compiled regular expression for regex patterns.
//
// The regex is compiled lazily on first access and cached for performance.
// Go-specific flags from the Crucible catalog are applied during compilation.
//
// Returns an error if:
//   - Pattern kind is not "regex"
//   - Regex compilation fails
//
// Example:
//
//	regex, err := pattern.CompiledRegex()
//	if err != nil {
//	    // Handle error
//	}
//	if regex.MatchString(input) {
//	    // Matched
//	}
func (p *Pattern) CompiledRegex() (*regexp.Regexp, error) {
	if p.Kind != PatternKindRegex {
		return nil, fmt.Errorf("pattern kind %s is not regex", p.Kind)
	}

	p.compileOnce.Do(func() {
		// Extract Go-specific flags
		goFlags, hasGoFlags := p.Flags["go"]

		patternStr := p.Pattern

		// Apply Go regex flags as inline flags
		if hasGoFlags {
			var flags string

			// Case insensitive: (?i)
			if goFlags["ignoreCase"] {
				flags += "i"
			}

			// Multiline: (?m)
			if goFlags["multiline"] {
				flags += "m"
			}

			// Dotall: (?s)
			if goFlags["dotall"] {
				flags += "s"
			}

			// Unicode is default in Go, so no flag needed

			if flags != "" {
				patternStr = fmt.Sprintf("(?%s)%s", flags, patternStr)
			}
		}

		p.compiledRegex, p.compileErr = regexp.Compile(patternStr)
	})

	if p.compileErr != nil {
		return nil, p.compileErr
	}

	return p.compiledRegex, nil
}

// Describe returns a formatted description of the pattern with examples.
//
// This is useful for documentation, debugging, or interactive pattern exploration.
//
// Example output:
//
//	RFC 5322 Email (ansi-email)
//	Pattern: ^[\p{L}\p{N}._%+-]+@[\p{L}\p{N}.-]+\.[\p{L}]{2,}$
//	Description: Internationalized email address with Unicode letters
//
//	Valid examples:
//	  - user@example.com
//	  - 你好@例子.公司
func (p *Pattern) Describe() string {
	result := fmt.Sprintf("%s (%s)\n", p.Name, p.ID)
	result += fmt.Sprintf("Pattern: %s\n", p.Pattern)

	if p.Description != "" {
		result += fmt.Sprintf("Description: %s\n", p.Description)
	}

	if len(p.Examples) > 0 {
		result += "\nValid examples:\n"
		for _, example := range p.Examples {
			result += fmt.Sprintf("  - %s\n", example)
		}
	}

	if len(p.NonExamples) > 0 {
		result += "\nInvalid examples:\n"
		for _, example := range p.NonExamples {
			result += fmt.Sprintf("  - %s\n", example)
		}
	}

	return result
}

// contains checks if substr is contained in s.
// This is a helper to avoid importing strings package.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

// findSubstring performs a simple substring search.
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
