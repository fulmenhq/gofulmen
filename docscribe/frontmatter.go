package docscribe

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	frontmatterDelimiter = "---"
)

// ParseFrontmatter extracts YAML frontmatter and returns both the clean content
// and the parsed metadata. This is the primary function for processing documents
// that may contain frontmatter.
//
// The function looks for YAML frontmatter delimited by "---" at the start of the
// document. If frontmatter is found, it returns:
//   - body: The document content with frontmatter removed
//   - metadata: The parsed YAML frontmatter as a map
//   - error: nil on success, ParseError if YAML is malformed
//
// If no frontmatter is present, returns:
//   - body: The original content unchanged
//   - metadata: nil
//   - error: nil
//
// Example input:
//
//	---
//	title: "My Document"
//	author: "Jane Doe"
//	tags: ["golang", "documentation"]
//	---
//	# My Document
//
//	This is the content.
//
// Returns:
//   - body: "# My Document\n\nThis is the content."
//   - metadata: map[string]interface{}{"title": "My Document", "author": "Jane Doe", ...}
func ParseFrontmatter(content []byte) (string, map[string]interface{}, error) {
	// Fast path: check if content could have frontmatter
	if !hasFrontmatter(content) {
		return string(content), nil, nil
	}

	// Extract frontmatter block and body
	yamlBlock, body, found := extractFrontmatterBlock(content)
	if !found {
		return string(content), nil, nil
	}

	// Parse the YAML frontmatter
	metadata, err := parseFrontmatterYAML(yamlBlock)
	if err != nil {
		return string(body), nil, err
	}

	return string(body), metadata, nil
}

// ExtractMetadata extracts only the YAML frontmatter metadata without returning
// the document body. This is more efficient than ParseFrontmatter when you only
// need the metadata.
//
// Returns nil if no frontmatter is present.
// Returns ParseError if frontmatter exists but YAML is malformed.
//
// Example:
//
//	metadata, err := documentation.ExtractMetadata(content)
//	if err != nil {
//	    return err
//	}
//	if metadata != nil {
//	    title := metadata["title"].(string)
//	    fmt.Printf("Document title: %s\n", title)
//	}
func ExtractMetadata(content []byte) (map[string]interface{}, error) {
	// Fast path: check if content could have frontmatter
	if !hasFrontmatter(content) {
		return nil, nil
	}

	// Extract frontmatter block
	yamlBlock, _, found := extractFrontmatterBlock(content)
	if !found {
		return nil, nil
	}

	// Parse the YAML frontmatter
	metadata, err := parseFrontmatterYAML(yamlBlock)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// StripFrontmatter removes YAML frontmatter and returns only the clean document body.
// This is the most efficient function when you only need the content without metadata.
//
// If no frontmatter is present, returns the original content unchanged.
// This function never returns an error - malformed frontmatter is simply left in place.
//
// Example:
//
//	cleanContent := documentation.StripFrontmatter(rawMarkdown)
//	// Process the markdown without frontmatter
//	renderMarkdown(cleanContent)
func StripFrontmatter(content []byte) string {
	// Fast path: check if content could have frontmatter
	if !hasFrontmatter(content) {
		return string(content)
	}

	// Extract body (ignore YAML block and errors)
	_, body, found := extractFrontmatterBlock(content)
	if !found {
		return string(content)
	}

	return string(body)
}

// hasFrontmatter performs a fast check to see if content might contain frontmatter.
// This avoids expensive parsing for content that clearly has no frontmatter.
func hasFrontmatter(content []byte) bool {
	// Must start with "---" (possibly after leading whitespace)
	trimmed := bytes.TrimLeft(content, " \t")
	return bytes.HasPrefix(trimmed, []byte(frontmatterDelimiter))
}

// extractFrontmatterBlock extracts the YAML frontmatter block and remaining body.
// Returns:
//   - yamlBlock: The YAML content between the delimiters (without "---" lines)
//   - body: The remaining content after frontmatter
//   - found: true if frontmatter was found and extracted
//
// The frontmatter must be at the very start of the document (after optional leading whitespace)
// and must be properly delimited by "---" on separate lines.
func extractFrontmatterBlock(content []byte) (yamlBlock []byte, body []byte, found bool) {
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 3 {
		// Need at minimum: "---", yaml content, "---"
		return nil, content, false
	}

	// Find the first non-whitespace line
	startIdx := 0
	for startIdx < len(lines) {
		trimmed := bytes.TrimSpace(lines[startIdx])
		if len(trimmed) > 0 {
			break
		}
		startIdx++
	}

	if startIdx >= len(lines) {
		return nil, content, false
	}

	// First non-whitespace line must be "---"
	if !isFrontmatterDelimiter(lines[startIdx]) {
		return nil, content, false
	}

	// Find closing delimiter
	closeIdx := -1
	for i := startIdx + 1; i < len(lines); i++ {
		if isFrontmatterDelimiter(lines[i]) {
			closeIdx = i
			break
		}
	}

	if closeIdx == -1 {
		// No closing delimiter found
		return nil, content, false
	}

	// Extract YAML block (between delimiters)
	yamlLines := lines[startIdx+1 : closeIdx]
	yamlBlock = bytes.Join(yamlLines, []byte("\n"))

	// Extract body (everything after closing delimiter)
	if closeIdx+1 < len(lines) {
		bodyLines := lines[closeIdx+1:]
		body = bytes.Join(bodyLines, []byte("\n"))
	} else {
		body = []byte("")
	}

	return yamlBlock, body, true
}

// isFrontmatterDelimiter checks if a line is a frontmatter delimiter.
// The delimiter must be exactly "---" with optional leading/trailing whitespace.
func isFrontmatterDelimiter(line []byte) bool {
	trimmed := strings.TrimSpace(string(line))
	return trimmed == frontmatterDelimiter
}

// parseFrontmatterYAML parses YAML frontmatter into a map.
// Returns ParseError if the YAML is malformed.
func parseFrontmatterYAML(yamlContent []byte) (map[string]interface{}, error) {
	// Handle empty frontmatter
	if len(bytes.TrimSpace(yamlContent)) == 0 {
		return make(map[string]interface{}), nil
	}

	var metadata map[string]interface{}
	err := yaml.Unmarshal(yamlContent, &metadata)
	if err != nil {
		return nil, wrapParseError("invalid frontmatter YAML", err)
	}

	return metadata, nil
}
