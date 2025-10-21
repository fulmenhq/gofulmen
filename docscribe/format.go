package docscribe

import (
	"bytes"
	"strings"
)

// DetectFormat performs heuristic-based format detection on content.
// It analyzes the structure and characteristics of the content to determine
// its format without relying on file extensions.
//
// Possible return values:
//   - "markdown": Markdown content (may include frontmatter)
//   - "yaml": Single YAML document
//   - "json": JSON document
//   - "toml": TOML configuration
//   - "text": Plain text
//   - "multi-yaml": YAML stream with multiple documents
//   - "multi-markdown": Concatenated markdown documents
//
// Detection heuristics (in order of precedence):
//  1. JSON: Starts with { or [
//  2. Multi-YAML: Contains multiple --- separators not in frontmatter pattern
//  3. YAML: YAML-like structure (key: value patterns)
//  4. Markdown: Contains markdown syntax (headers, lists, links)
//  5. TOML: Contains TOML section headers [section]
//  6. Text: Default fallback
//
// Example:
//
//	format := documentation.DetectFormat(content)
//	switch format {
//	case documentation.FormatJSON:
//	    processJSON(content)
//	case documentation.FormatMarkdown:
//	    processMarkdown(content)
//	default:
//	    processPlainText(content)
//	}
func DetectFormat(content []byte) string {
	// Handle empty content
	if len(content) == 0 {
		return FormatText
	}

	// Trim leading whitespace for detection
	trimmed := bytes.TrimLeft(content, " \t\n\r")
	if len(trimmed) == 0 {
		return FormatText
	}

	// 1. Check for JSON (starts with { or [)
	if trimmed[0] == '{' || trimmed[0] == '[' {
		return FormatJSON
	}

	// 2. Check for multi-document formats
	if multiFormat := detectMultiDocumentFormat(content); multiFormat != "" {
		return multiFormat
	}

	// 3. Check for TOML (contains [section] headers)
	if looksLikeTOML(content) {
		return FormatTOML
	}

	// 4. Check for YAML vs Markdown
	// Both can have frontmatter, so we need to look deeper
	yamlScore := calculateYAMLScore(content)
	markdownScore := calculateMarkdownScore(content)

	if yamlScore > markdownScore && yamlScore > 2 {
		return FormatYAML
	}

	if markdownScore > 0 {
		return FormatMarkdown
	}

	// 5. Default to text
	return FormatText
}

// detectMultiDocumentFormat checks if content contains multiple documents.
// Returns "multi-yaml" or "multi-markdown" if detected, empty string otherwise.
//
// To qualify as multi-document, we need at least 2 separators (not counting frontmatter)
// with actual content between them. A single --- is likely just a markdown horizontal rule.
func detectMultiDocumentFormat(content []byte) string {
	lines := bytes.Split(content, []byte("\n"))
	separatorCount := 0
	hasFrontmatter := false

	// Check if starts with frontmatter
	if len(lines) > 0 {
		if isFrontmatterDelimiter(lines[0]) {
			hasFrontmatter = true
		}
	}

	// Count --- separators that aren't frontmatter delimiters
	// Also track if there's content between separators
	inCodeBlock := false
	frontmatterClosed := false
	lastSeparatorIdx := -1
	hasContentBetweenSeparators := false

	for i, line := range lines {
		// Track code blocks
		if isCodeBlockFence(line) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// Check for separator
		if isFrontmatterDelimiter(line) {
			// If we have frontmatter, the first two --- are frontmatter delimiters
			if hasFrontmatter && !frontmatterClosed {
				if i == 0 {
					// Opening delimiter
					continue
				}
				// This is the closing delimiter
				frontmatterClosed = true
				lastSeparatorIdx = i
				continue
			}
			// This is a document separator
			separatorCount++

			// Check if there was content since last separator
			if lastSeparatorIdx >= 0 {
				if hasNonEmptyContentBetween(lines, lastSeparatorIdx, i) {
					hasContentBetweenSeparators = true
				}
			}
			lastSeparatorIdx = i
		}
	}

	// Require at least 2 separators with actual content between them
	// A single --- is likely just a markdown horizontal rule, not a document separator
	if separatorCount >= 2 && hasContentBetweenSeparators {
		// Check if it's YAML or markdown based on content
		if looksLikeYAMLStream(content) {
			return FormatMultiYAML
		}
		return FormatMultiMarkdown
	}

	return ""
}

// hasNonEmptyContentBetween checks if there's actual content between two line indices.
// Returns false if the range only contains empty lines or whitespace.
func hasNonEmptyContentBetween(lines [][]byte, startIdx, endIdx int) bool {
	for i := startIdx + 1; i < endIdx; i++ {
		if len(bytes.TrimSpace(lines[i])) > 0 {
			return true
		}
	}
	return false
}

// looksLikeYAMLStream checks if content resembles a YAML stream.
func looksLikeYAMLStream(content []byte) bool {
	lines := bytes.Split(content, []byte("\n"))
	yamlPatternCount := 0

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}

		// Look for YAML patterns: "key: value" or "- item"
		if bytes.Contains(trimmed, []byte(": ")) || bytes.HasPrefix(trimmed, []byte("- ")) {
			yamlPatternCount++
		}

		// Stop early if we have enough evidence
		if yamlPatternCount > 3 {
			return true
		}
	}

	return yamlPatternCount > 0
}

// looksLikeTOML checks if content contains TOML section headers.
func looksLikeTOML(content []byte) bool {
	lines := bytes.Split(content, []byte("\n"))

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		// TOML sections: [section] or [[array]]
		if bytes.HasPrefix(trimmed, []byte("[")) && bytes.HasSuffix(trimmed, []byte("]")) {
			// Make sure it's not a markdown link [text](url)
			if !bytes.Contains(trimmed, []byte("](")) {
				return true
			}
		}
	}

	return false
}

// calculateYAMLScore estimates how YAML-like the content is.
func calculateYAMLScore(content []byte) int {
	score := 0
	lines := bytes.Split(content, []byte("\n"))

	// Skip frontmatter if present
	startIdx := 0
	if len(lines) > 0 && isFrontmatterDelimiter(lines[0]) {
		// Find closing delimiter
		for i := 1; i < len(lines); i++ {
			if isFrontmatterDelimiter(lines[i]) {
				startIdx = i + 1
				break
			}
		}
	}

	for i := startIdx; i < len(lines) && i < startIdx+50; i++ {
		line := bytes.TrimSpace(lines[i])

		// Skip empty lines and comments
		if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
			continue
		}

		// YAML key-value pattern: "key: value"
		if bytes.Contains(line, []byte(": ")) {
			score += 2
		}

		// YAML list pattern: "- item"
		if bytes.HasPrefix(line, []byte("- ")) {
			score += 1
		}

		// YAML array notation
		if bytes.HasPrefix(line, []byte("[")) && bytes.HasSuffix(line, []byte("]")) {
			score += 1
		}
	}

	return score
}

// calculateMarkdownScore estimates how markdown-like the content is.
func calculateMarkdownScore(content []byte) int {
	score := 0
	lines := bytes.Split(content, []byte("\n"))

	for i := 0; i < len(lines) && i < 50; i++ {
		line := lines[i]
		trimmed := bytes.TrimSpace(line)

		if len(trimmed) == 0 {
			continue
		}

		lineStr := string(trimmed)

		// ATX headers: # Header
		if bytes.HasPrefix(trimmed, []byte("#")) {
			if atxHeaderRegex.Match(line) {
				score += 3
			}
		}

		// Lists: - item or * item or 1. item
		if bytes.HasPrefix(trimmed, []byte("- ")) ||
			bytes.HasPrefix(trimmed, []byte("* ")) ||
			(len(lineStr) > 2 && lineStr[0] >= '0' && lineStr[0] <= '9' && lineStr[1] == '.') {
			score += 2
		}

		// Links: [text](url) or [text][ref]
		if bytes.Contains(trimmed, []byte("](")) || bytes.Contains(trimmed, []byte("][")) {
			score += 2
		}

		// Bold/italic: **bold** or *italic*
		if bytes.Contains(trimmed, []byte("**")) || bytes.Contains(trimmed, []byte("__")) {
			score += 1
		}

		// Code blocks: ``` or ~~~
		if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
			score += 2
		}

		// Blockquotes: > quote
		if bytes.HasPrefix(trimmed, []byte("> ")) {
			score += 1
		}

		// Setext underlines (check next line)
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(string(lines[i+1]))
			if isSetextUnderline(nextLine, '=') || isSetextUnderline(nextLine, '-') {
				score += 3
			}
		}
	}

	return score
}
