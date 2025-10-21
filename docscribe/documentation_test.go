package docscribe

import (
	"os"
	"path/filepath"
	"testing"
)

// Test fixtures directory
const fixturesDir = "../test/fixtures/docscribe"

// loadFixture is a helper to load test fixture files
func loadFixture(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join(fixturesDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}
	return content
}

// TestParseFrontmatter tests frontmatter parsing with various inputs
func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name            string
		fixture         string
		wantBody        string
		wantHasMetadata bool
		wantTitle       string
		wantError       bool
	}{
		{
			name:            "valid frontmatter",
			fixture:         "valid-frontmatter.md",
			wantHasMetadata: true,
			wantTitle:       "Test Document",
			wantError:       false,
		},
		{
			name:            "no frontmatter",
			fixture:         "no-frontmatter.md",
			wantHasMetadata: false,
			wantError:       false,
		},
		{
			name:            "empty frontmatter",
			fixture:         "empty-frontmatter.md",
			wantHasMetadata: true,
			wantError:       false,
		},
		{
			name:      "malformed YAML",
			fixture:   "malformed-yaml.md",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			body, metadata, err := ParseFrontmatter(content)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.wantHasMetadata {
				if metadata == nil {
					t.Error("Expected metadata, got nil")
					return
				}
				if tt.wantTitle != "" {
					if title, ok := metadata["title"].(string); !ok || title != tt.wantTitle {
						t.Errorf("Expected title %q, got %q", tt.wantTitle, title)
					}
				}
			} else {
				if metadata != nil {
					t.Errorf("Expected nil metadata, got %v", metadata)
				}
			}

			if body == "" && tt.wantBody != "" {
				t.Error("Expected non-empty body")
			}
		})
	}
}

// TestExtractMetadata tests metadata extraction
func TestExtractMetadata(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		wantNil   bool
		wantError bool
	}{
		{"valid frontmatter", "valid-frontmatter.md", false, false},
		{"no frontmatter", "no-frontmatter.md", true, false},
		{"empty frontmatter", "empty-frontmatter.md", false, false},
		{"malformed YAML", "malformed-yaml.md", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			metadata, err := ExtractMetadata(content)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.wantNil && metadata != nil {
				t.Errorf("Expected nil metadata, got %v", metadata)
			}

			if !tt.wantNil && metadata == nil {
				t.Error("Expected metadata, got nil")
			}
		})
	}
}

// TestStripFrontmatter tests frontmatter stripping
func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		wantContains  string
		wantNotPrefix string
	}{
		{
			name:          "valid frontmatter",
			fixture:       "valid-frontmatter.md",
			wantContains:  "# Test Document",
			wantNotPrefix: "---",
		},
		{
			name:         "no frontmatter",
			fixture:      "no-frontmatter.md",
			wantContains: "# Document Without Frontmatter",
		},
		{
			name:          "empty frontmatter",
			fixture:       "empty-frontmatter.md",
			wantContains:  "# Document With Empty Frontmatter",
			wantNotPrefix: "---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			result := StripFrontmatter(content)

			if tt.wantContains != "" {
				if len(result) == 0 || !contains(result, tt.wantContains) {
					t.Errorf("Expected result to contain %q", tt.wantContains)
				}
			}

			if tt.wantNotPrefix != "" {
				if hasPrefix(result, tt.wantNotPrefix) {
					t.Errorf("Expected result not to start with %q", tt.wantNotPrefix)
				}
			}
		})
	}
}

// TestExtractHeaders tests header extraction
func TestExtractHeaders(t *testing.T) {
	tests := []struct {
		name       string
		fixture    string
		wantCount  int
		wantLevels []int
	}{
		{
			name:      "valid frontmatter",
			fixture:   "valid-frontmatter.md",
			wantCount: 4,
		},
		{
			name:       "no frontmatter",
			fixture:    "no-frontmatter.md",
			wantCount:  3,
			wantLevels: []int{1, 2, 2},
		},
		{
			name:      "setext headers",
			fixture:   "setext-headers.md",
			wantCount: 4,
		},
		{
			name:       "all header levels",
			fixture:    "all-header-levels.md",
			wantCount:  6,
			wantLevels: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:      "code blocks (headers inside ignored)",
			fixture:   "code-blocks.md",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			headers, err := ExtractHeaders(content)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(headers) != tt.wantCount {
				t.Errorf("Expected %d headers, got %d", tt.wantCount, len(headers))
			}

			if tt.wantLevels != nil {
				for i, wantLevel := range tt.wantLevels {
					if i >= len(headers) {
						break
					}
					if headers[i].Level != wantLevel {
						t.Errorf("Header %d: expected level %d, got %d", i, wantLevel, headers[i].Level)
					}
				}
			}

			// Verify all headers have anchors and line numbers
			for i, h := range headers {
				if h.Anchor == "" {
					t.Errorf("Header %d missing anchor", i)
				}
				if h.LineNumber == 0 {
					t.Errorf("Header %d missing line number", i)
				}
			}
		})
	}
}

// TestGenerateAnchor tests anchor slug generation
func TestGenerateAnchor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"API v2.0 (Beta)", "api-v2-0-beta"},
		{"What's New?", "what-s-new"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Trailing hyphens---", "trailing-hyphens"},
		{"UPPERCASE", "uppercase"},
		{"mixed_CASE_text", "mixed_case_text"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := generateAnchor(tt.input)
			if got != tt.want {
				t.Errorf("generateAnchor(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestDetectFormat tests format detection
func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		want    string
	}{
		{"markdown with frontmatter", "valid-frontmatter.md", FormatMarkdown},
		{"markdown no frontmatter", "no-frontmatter.md", FormatMarkdown},
		{"JSON", "json-content.json", FormatJSON},
		{"YAML", "yaml-content.yaml", FormatYAML},
		{"TOML", "toml-content.toml", FormatJSON},
		{"YAML stream", "yaml-stream.yaml", FormatMultiYAML},
		{"multi markdown", "multi-markdown.md", FormatMultiYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			got := DetectFormat(content)

			if got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDetectFormatEdgeCases tests format detection edge cases
func TestDetectFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"empty content", "", FormatText},
		{"whitespace only", "   \n\n  \t  ", FormatText},
		{"JSON array", "[1, 2, 3]", FormatJSON},
		{"JSON object", `{"key": "value"}`, FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFormat([]byte(tt.content))
			if got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestInspectDocument tests document inspection
func TestInspectDocument(t *testing.T) {
	tests := []struct {
		name            string
		fixture         string
		wantFormat      string
		wantFrontmatter bool
		wantHeaderCount int
		wantMinSections int
	}{
		{
			name:            "markdown with frontmatter",
			fixture:         "valid-frontmatter.md",
			wantFormat:      FormatMarkdown,
			wantFrontmatter: true,
			wantHeaderCount: 4,
			wantMinSections: 1,
		},
		{
			name:            "no frontmatter",
			fixture:         "no-frontmatter.md",
			wantFormat:      FormatMarkdown,
			wantFrontmatter: false,
			wantHeaderCount: 3,
			wantMinSections: 1,
		},
		{
			name:            "JSON",
			fixture:         "json-content.json",
			wantFormat:      FormatJSON,
			wantFrontmatter: false,
			wantHeaderCount: 0,
		},
		{
			name:            "YAML",
			fixture:         "yaml-content.yaml",
			wantFormat:      FormatYAML,
			wantFrontmatter: false,
			wantHeaderCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			info, err := InspectDocument(content)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if info.Format != tt.wantFormat {
				t.Errorf("Format = %q, want %q", info.Format, tt.wantFormat)
			}

			if info.HasFrontmatter != tt.wantFrontmatter {
				t.Errorf("HasFrontmatter = %v, want %v", info.HasFrontmatter, tt.wantFrontmatter)
			}

			if info.HeaderCount != tt.wantHeaderCount {
				t.Errorf("HeaderCount = %d, want %d", info.HeaderCount, tt.wantHeaderCount)
			}

			if tt.wantMinSections > 0 && info.EstimatedSections < tt.wantMinSections {
				t.Errorf("EstimatedSections = %d, want >= %d", info.EstimatedSections, tt.wantMinSections)
			}

			if info.LineCount == 0 {
				t.Error("LineCount should not be 0")
			}
		})
	}
}

// TestSplitDocuments tests multi-document splitting
func TestSplitDocuments(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		wantCount int
		wantError bool
	}{
		{
			name:      "single document with frontmatter",
			fixture:   "valid-frontmatter.md",
			wantCount: 1,
		},
		{
			name:      "horizontal rule not split",
			fixture:   "horizontal-rule.md",
			wantCount: 1,
		},
		{
			name:      "multi markdown",
			fixture:   "multi-markdown.md",
			wantCount: 3,
		},
		{
			name:      "YAML stream",
			fixture:   "yaml-stream.yaml",
			wantCount: 3,
		},
		{
			name:      "code blocks ignored",
			fixture:   "code-blocks.md",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := loadFixture(t, tt.fixture)
			docs, err := SplitDocuments(content)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(docs) != tt.wantCount {
				t.Errorf("Expected %d documents, got %d", tt.wantCount, len(docs))
			}

			// Verify each document is non-empty
			for i, doc := range docs {
				if len(doc) == 0 {
					t.Errorf("Document %d is empty", i)
				}
			}
		})
	}
}

// TestSplitDocumentsEdgeCases tests edge cases in document splitting
func TestSplitDocumentsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
	}{
		{
			name:      "empty content",
			content:   "",
			wantCount: 0,
		},
		{
			name:      "single ---",
			content:   "---",
			wantCount: 1,
		},
		{
			name:      "two --- no content",
			content:   "---\n---",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := SplitDocuments([]byte(tt.content))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(docs) != tt.wantCount {
				t.Errorf("Expected %d documents, got %d", tt.wantCount, len(docs))
			}
		})
	}
}

// TestErrorTypes tests error type construction and messages
func TestErrorTypes(t *testing.T) {
	t.Run("ParseError", func(t *testing.T) {
		err := &ParseError{
			Message:    "test error",
			LineNumber: 10,
			Column:     5,
			Context:    "test context",
		}
		errMsg := err.Error()
		if !contains(errMsg, "line 10") || !contains(errMsg, "column 5") {
			t.Errorf("Error message missing line/column: %s", errMsg)
		}
	})

	t.Run("FormatError", func(t *testing.T) {
		err := &FormatError{
			Expected: "yaml",
			Actual:   "json",
			Message:  "format mismatch",
		}
		errMsg := err.Error()
		if !contains(errMsg, "yaml") || !contains(errMsg, "json") {
			t.Errorf("Error message missing format info: %s", errMsg)
		}
	})

	t.Run("newParseError", func(t *testing.T) {
		err := newParseError("test")
		if err.Message != "test" {
			t.Error("newParseError failed")
		}
	})

	t.Run("newParseErrorWithLine", func(t *testing.T) {
		err := newParseErrorWithLine("test", 5)
		if err.LineNumber != 5 {
			t.Error("newParseErrorWithLine failed")
		}
	})

	t.Run("newParseErrorWithLocation", func(t *testing.T) {
		err := newParseErrorWithLocation("test", 5, 10)
		if err.LineNumber != 5 || err.Column != 10 {
			t.Error("newParseErrorWithLocation failed")
		}
	})

	t.Run("newFormatError", func(t *testing.T) {
		err := newFormatError("yaml", "json", "mismatch")
		if err.Expected != "yaml" {
			t.Error("newFormatError failed")
		}
	})

	t.Run("wrapFormatError", func(t *testing.T) {
		underlying := newParseError("underlying")
		err := wrapFormatError("yaml", "json", underlying)
		if err.Unwrap() != underlying {
			t.Error("wrapFormatError failed")
		}
	})
}

// TestCountLines tests line counting edge cases
func TestCountLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty", "", 0},
		{"single line no newline", "test", 1},
		{"single line with newline", "test\n", 1},
		{"multiple lines", "line1\nline2\nline3", 3},
		{"multiple lines trailing newline", "line1\nline2\nline3\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLines([]byte(tt.content))
			if got != tt.want {
				t.Errorf("countLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCodeBlockFenceTypes tests different code fence types
func TestCodeBlockFenceTypes(t *testing.T) {
	content := []byte("# Header\n\n```\ncode\n```\n\n~~~\ncode\n~~~\n\n## Another")
	headers, err := ExtractHeaders(content)
	if err != nil {
		t.Fatal(err)
	}
	// Should only find the headers, not code inside fences
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}
}

// TestOnlyEmptyLines tests the onlyEmptyLines helper
func TestOnlyEmptyLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  bool
	}{
		{"all empty", []string{"", "  ", "\t"}, true},
		{"has content", []string{"", "content", ""}, false},
		{"empty slice", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := make([][]byte, len(tt.lines))
			for i, l := range tt.lines {
				lines[i] = []byte(l)
			}
			got := onlyEmptyLines(lines)
			if got != tt.want {
				t.Errorf("onlyEmptyLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWrapParseError tests error wrapping
func TestWrapParseError(t *testing.T) {
	underlying := &ParseError{Message: "yaml: line 5: bad indent"}
	wrapped := wrapParseError("frontmatter", underlying)
	if wrapped.Unwrap() != underlying {
		t.Error("wrapParseError should wrap underlying error")
	}
	if wrapped.LineNumber != 5 {
		t.Errorf("Expected line number 5, got %d", wrapped.LineNumber)
	}
}

// TestExtractLineNumber tests line number extraction from errors
func TestExtractLineNumber(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"yaml error with line", &ParseError{Message: "yaml: line 10: error"}, 10},
		{"plain message", &ParseError{Message: "some error"}, 0},
		{"nil error", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLineNumberFromError(tt.err)
			if got != tt.want {
				t.Errorf("extractLineNumberFromError() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
