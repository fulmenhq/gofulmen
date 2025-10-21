package docscribe

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

// ATX header pattern: matches "# ", "## ", etc. at start of line
// Group 1: the # symbols, Group 2: the header text
var atxHeaderRegex = regexp.MustCompile(`^(#{1,6})\s+(.+?)(?:\s+#*)?$`)

// ExtractHeaders extracts all markdown headers from the content with their
// hierarchy, anchors, and line numbers.
//
// This function supports both ATX-style headers (# Header) and Setext-style
// headers (underlined with === or ---). It generates URL-safe anchor slugs
// for each header and tracks the line number for source mapping.
//
// Headers inside code blocks (fenced with ``` or indented) are ignored.
//
// Example:
//
//	headers, err := documentation.ExtractHeaders(content)
//	if err != nil {
//	    return err
//	}
//	for _, h := range headers {
//	    fmt.Printf("%s %s (#%s at line %d)\n",
//	        strings.Repeat("#", h.Level), h.Text, h.Anchor, h.LineNumber)
//	}
//
// Returns a slice of Header structs, or an error if content cannot be processed.
func ExtractHeaders(content []byte) ([]Header, error) {
	var headers []Header
	lines := bytes.Split(content, []byte("\n"))

	inCodeBlock := false
	codeBlockFence := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineNum := i + 1 // 1-based line numbers

		// Track code block state
		if isCodeBlockFence(line) {
			fence := getCodeBlockFence(line)
			if !inCodeBlock {
				// Entering code block
				inCodeBlock = true
				codeBlockFence = fence
			} else if fence == codeBlockFence {
				// Exiting code block (matching fence)
				inCodeBlock = false
				codeBlockFence = ""
			}
			continue
		}

		// Skip lines inside code blocks
		if inCodeBlock {
			continue
		}

		// Try ATX-style header first (# Header)
		if header, found := parseATXHeader(line, lineNum); found {
			headers = append(headers, header)
			continue
		}

		// Try Setext-style header (underlined)
		// Need to look at next line for underline
		if i+1 < len(lines) {
			if header, found := parseSetextHeader(line, lines[i+1], lineNum); found {
				headers = append(headers, header)
				i++ // Skip the underline line
				continue
			}
		}
	}

	return headers, nil
}

// parseATXHeader parses an ATX-style header (# Header).
// Returns the Header and true if this line is an ATX header.
func parseATXHeader(line []byte, lineNum int) (Header, bool) {
	matches := atxHeaderRegex.FindSubmatch(line)
	if matches == nil {
		return Header{}, false
	}

	level := len(matches[1])       // Count the # symbols
	text := string(matches[2])     // Extract header text
	text = strings.TrimSpace(text) // Clean up whitespace

	return Header{
		Level:      level,
		Text:       text,
		Anchor:     generateAnchor(text),
		LineNumber: lineNum,
	}, true
}

// parseSetextHeader parses a Setext-style header (underlined with === or ---).
// Returns the Header and true if this is a Setext header.
func parseSetextHeader(line, nextLine []byte, lineNum int) (Header, bool) {
	// Text line must be non-empty
	text := strings.TrimSpace(string(line))
	if text == "" {
		return Header{}, false
	}

	// Next line must be all = or all -
	underline := strings.TrimSpace(string(nextLine))
	if underline == "" {
		return Header{}, false
	}

	// Check if underline is all === (H1) or all --- (H2)
	var level int
	if isSetextUnderline(underline, '=') {
		level = 1
	} else if isSetextUnderline(underline, '-') {
		level = 2
	} else {
		return Header{}, false
	}

	return Header{
		Level:      level,
		Text:       text,
		Anchor:     generateAnchor(text),
		LineNumber: lineNum,
	}, true
}

// isSetextUnderline checks if a line is a valid Setext underline.
// The line must contain only the specified character (and optional whitespace).
func isSetextUnderline(line string, char rune) bool {
	if len(line) == 0 {
		return false
	}

	// Must have at least one character
	hasChar := false
	for _, r := range line {
		if r == char {
			hasChar = true
		} else if r != ' ' && r != '\t' {
			// Found a character that's not the underline char or whitespace
			return false
		}
	}

	return hasChar
}

// generateAnchor creates a URL-safe anchor slug from header text.
// This matches common markdown renderer behavior:
//   - Convert to lowercase
//   - Replace spaces with hyphens
//   - Remove special characters except hyphens and underscores
//   - Collapse multiple hyphens
//   - Trim leading/trailing hyphens
//
// Examples:
//   - "Hello World" → "hello-world"
//   - "API v2.0 (Beta)" → "api-v20-beta"
//   - "What's New?" → "whats-new"
func generateAnchor(text string) string {
	// Convert to lowercase
	slug := strings.ToLower(text)

	// Build the slug character by character
	var result strings.Builder
	lastWasHyphen := false

	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			lastWasHyphen = false
		} else if r == '_' {
			// Keep underscores
			result.WriteRune(r)
			lastWasHyphen = false
		} else if unicode.IsSpace(r) || r == '-' || !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Replace spaces and special chars with hyphen (but avoid duplicates)
			if !lastWasHyphen && result.Len() > 0 {
				result.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}

	// Trim trailing hyphen
	anchor := strings.TrimRight(result.String(), "-")

	return anchor
}

// isCodeBlockFence checks if a line starts a code block fence.
func isCodeBlockFence(line []byte) bool {
	trimmed := bytes.TrimLeft(line, " \t")
	return bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~"))
}

// getCodeBlockFence extracts the fence characters from a code block fence line.
// Returns "```" or "~~~" to track which fence type to match for closing.
func getCodeBlockFence(line []byte) string {
	trimmed := bytes.TrimLeft(line, " \t")
	if bytes.HasPrefix(trimmed, []byte("```")) {
		return "```"
	}
	if bytes.HasPrefix(trimmed, []byte("~~~")) {
		return "~~~"
	}
	return ""
}
