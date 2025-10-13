package foundry

import (
	"testing"
)

func TestPattern_Match_Regex(t *testing.T) {
	pattern := &Pattern{
		ID:      "email",
		Kind:    PatternKindRegex,
		Pattern: `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`,
	}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with subdomain", "admin@mail.example.com", true},
		{"Invalid - no @", "invalid.email", false},
		{"Invalid - no domain", "user@", false},
		{"Invalid - no TLD", "user@example", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pattern.Match(tt.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Match(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPattern_Match_Literal(t *testing.T) {
	pattern := &Pattern{
		ID:      "keyword",
		Kind:    PatternKindLiteral,
		Pattern: "exact_match",
	}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Exact match", "exact_match", true},
		{"Case sensitive", "Exact_Match", false},
		{"Partial match", "exact", false},
		{"Contains but not exact", "prefix_exact_match", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pattern.Match(tt.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Match(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPattern_Match_Glob(t *testing.T) {
	pattern := &Pattern{
		ID:      "json-files",
		Kind:    PatternKindGlob,
		Pattern: "*.json",
	}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"JSON file", "config.json", true},
		{"JSON file uppercase", "DATA.JSON", false}, // Go's filepath.Match is case-sensitive
		{"YAML file", "config.yaml", false},
		{"No extension", "config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pattern.Match(tt.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Match(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPattern_MustMatch(t *testing.T) {
	pattern := &Pattern{
		ID:      "slug",
		Kind:    PatternKindRegex,
		Pattern: `^[a-z0-9]+(?:[-_][a-z0-9]+)*$`,
	}

	if !pattern.MustMatch("valid-slug") {
		t.Error("Expected MustMatch to return true for valid slug")
	}

	if pattern.MustMatch("Invalid Slug") {
		t.Error("Expected MustMatch to return false for invalid slug")
	}
}

func TestPattern_Search(t *testing.T) {
	pattern := &Pattern{
		ID:      "digits",
		Kind:    PatternKindRegex,
		Pattern: `\d+`,
	}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Contains digits", "Order #12345 shipped", true},
		{"Digits at start", "123abc", true},
		{"Digits at end", "abc123", true},
		{"No digits", "no numbers here", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pattern.Search(tt.value)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Search(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPattern_CompiledRegex(t *testing.T) {
	pattern := &Pattern{
		ID:      "slug",
		Kind:    PatternKindRegex,
		Pattern: `^[a-z]+$`,
	}

	// First call should compile
	regex1, err := pattern.CompiledRegex()
	if err != nil {
		t.Errorf("Failed to compile regex: %v", err)
	}

	// Second call should return cached regex
	regex2, err := pattern.CompiledRegex()
	if err != nil {
		t.Errorf("Failed to get cached regex: %v", err)
	}

	// Should be same instance (pointer equality)
	if regex1 != regex2 {
		t.Error("Expected compiled regex to be cached")
	}
}

func TestPattern_CompiledRegex_WithFlags(t *testing.T) {
	pattern := &Pattern{
		ID:      "case-insensitive",
		Kind:    PatternKindRegex,
		Pattern: `^hello$`,
		Flags: PatternFlags{
			"go": {"ignoreCase": true},
		},
	}

	regex, err := pattern.CompiledRegex()
	if err != nil {
		t.Errorf("Failed to compile regex with flags: %v", err)
	}

	if !regex.MatchString("HELLO") {
		t.Error("Expected case-insensitive match for HELLO")
	}

	if !regex.MatchString("hello") {
		t.Error("Expected case-insensitive match for hello")
	}
}

func TestPattern_CompiledRegex_NonRegex(t *testing.T) {
	pattern := &Pattern{
		ID:      "literal",
		Kind:    PatternKindLiteral,
		Pattern: "test",
	}

	_, err := pattern.CompiledRegex()
	if err == nil {
		t.Error("Expected error when calling CompiledRegex on non-regex pattern")
	}
}

func TestPattern_Describe(t *testing.T) {
	pattern := &Pattern{
		ID:          "email",
		Name:        "Email Address",
		Kind:        PatternKindRegex,
		Pattern:     `^[a-z]+@[a-z]+\.[a-z]+$`,
		Description: "Simple email pattern",
		Examples:    []string{"user@example.com"},
		NonExamples: []string{"invalid"},
	}

	description := pattern.Describe()

	// Verify key elements are present
	if !contains(description, "Email Address") {
		t.Error("Description should contain pattern name")
	}
	if !contains(description, "email") {
		t.Error("Description should contain pattern ID")
	}
	if !contains(description, "user@example.com") {
		t.Error("Description should contain examples")
	}
	if !contains(description, "invalid") {
		t.Error("Description should contain non-examples")
	}
}

func BenchmarkPattern_Match_Regex(b *testing.B) {
	pattern := &Pattern{
		ID:      "email",
		Kind:    PatternKindRegex,
		Pattern: `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`,
	}
	value := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern.Match(value)
	}
}
