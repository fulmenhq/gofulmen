package docscribe

import (
	"bytes"
)

// SplitDocuments splits multi-document content into individual documents.
// This handles both YAML streams and concatenated markdown documents.
//
// The "---" separator serves dual purposes that must be correctly distinguished:
//  1. Frontmatter delimiter in markdown (opening and closing)
//  2. Document separator in YAML streams and concatenated markdown
//
// The function uses context-aware parsing to determine the role of each "---":
//   - At start of a document: frontmatter opening delimiter
//   - After frontmatter opening: frontmatter closing delimiter
//   - After document content: document separator
//   - Inside code blocks: literal text (ignored)
//
// Examples:
//
// YAML stream (Kubernetes manifests):
//
//	apiVersion: v1
//	kind: Pod
//	---
//	apiVersion: v1
//	kind: Service
//
// Returns: ["apiVersion: v1\nkind: Pod", "apiVersion: v1\nkind: Service"]
//
// Concatenated markdown with frontmatter:
//
//	---
//	title: Doc 1
//	---
//	# Document 1
//	---
//	---
//	title: Doc 2
//	---
//	# Document 2
//
// Returns: ["---\ntitle: Doc 1\n---\n# Document 1", "---\ntitle: Doc 2\n---\n# Document 2"]
//
// Single document with frontmatter (no splitting):
//
//	---
//	title: Single Doc
//	---
//	# Content
//
// Returns: ["---\ntitle: Single Doc\n---\n# Content"] (one document)
//
// Returns a slice of document strings, or an error if splitting fails.
func SplitDocuments(content []byte) ([]string, error) {
	if len(content) == 0 {
		return []string{}, nil
	}

	lines := bytes.Split(content, []byte("\n"))
	var documents []string
	var currentDoc [][]byte

	// State tracking for context-aware parsing
	state := &splitState{
		inCodeBlock:       false,
		atDocumentStart:   true,
		inFrontmatter:     false,
		frontmatterClosed: false,
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Track code block boundaries
		if isCodeBlockFence(line) {
			state.inCodeBlock = !state.inCodeBlock
			currentDoc = append(currentDoc, line)
			continue
		}

		// Inside code blocks, treat everything as literal content
		if state.inCodeBlock {
			currentDoc = append(currentDoc, line)
			state.atDocumentStart = false
			continue
		}

		// Check if this is a "---" delimiter
		if isFrontmatterDelimiter(line) {
			action := state.classifyDelimiter(currentDoc, lines, i)

			switch action {
			case delimiterActionFrontmatterOpen:
				// This is the opening frontmatter delimiter
				state.inFrontmatter = true
				state.atDocumentStart = false
				currentDoc = append(currentDoc, line)

			case delimiterActionFrontmatterClose:
				// This is the closing frontmatter delimiter
				state.inFrontmatter = false
				state.frontmatterClosed = true
				currentDoc = append(currentDoc, line)

			case delimiterActionDocumentSeparator:
				// This is a document separator - finish current doc and start new one
				if len(currentDoc) > 0 {
					docContent := bytes.Join(currentDoc, []byte("\n"))
					if len(bytes.TrimSpace(docContent)) > 0 {
						documents = append(documents, string(docContent))
					}
				}
				// Reset for new document
				currentDoc = nil
				state.reset()

				// Check if there's a second "---" coming soon (double-delimiter pattern: ---\n\n---)
				// Skip past empty lines and check if we find another delimiter
				skipIdx := i + 1
				for skipIdx < len(lines) && len(bytes.TrimSpace(lines[skipIdx])) == 0 {
					skipIdx++
				}
				// If we found another delimiter within 2 lines, skip past it
				if skipIdx < len(lines) && skipIdx <= i+2 && isFrontmatterDelimiter(lines[skipIdx]) {
					i = skipIdx // Skip the second delimiter
				}

			case delimiterActionLiteral:
				// This is literal content (horizontal rule in markdown)
				currentDoc = append(currentDoc, line)
				state.atDocumentStart = false
			}
		} else {
			// Regular content line
			currentDoc = append(currentDoc, line)

			// Any non-empty, non-whitespace line means we're past document start
			if len(bytes.TrimSpace(line)) > 0 {
				state.atDocumentStart = false
			}
		}
	}

	// Add the last document
	if len(currentDoc) > 0 {
		docContent := bytes.Join(currentDoc, []byte("\n"))
		if len(bytes.TrimSpace(docContent)) > 0 {
			documents = append(documents, string(docContent))
		}
	}

	// If we only found one document, return it as-is (not split)
	// This handles the common case of a single document with frontmatter
	if len(documents) == 0 {
		return []string{string(content)}, nil
	}

	return documents, nil
}

// splitState tracks the parsing state for context-aware delimiter classification.
type splitState struct {
	inCodeBlock       bool // Currently inside a code block
	atDocumentStart   bool // At the very start of a new document
	inFrontmatter     bool // Inside frontmatter block
	frontmatterClosed bool // Frontmatter has been closed for current doc
}

// delimiterAction indicates how to handle a "---" delimiter.
type delimiterAction int

const (
	delimiterActionFrontmatterOpen delimiterAction = iota
	delimiterActionFrontmatterClose
	delimiterActionDocumentSeparator
	delimiterActionLiteral
)

// classifyDelimiter determines the role of a "---" delimiter based on context.
// It uses lookahead to distinguish document separators from literal horizontal rules.
func (s *splitState) classifyDelimiter(currentDoc [][]byte, allLines [][]byte, currentIdx int) delimiterAction {
	// If we're at the very start of a document (no content yet), this opens frontmatter
	if s.atDocumentStart && len(currentDoc) == 0 {
		return delimiterActionFrontmatterOpen
	}

	// If we're at document start with only empty lines, this opens frontmatter
	if s.atDocumentStart && onlyEmptyLines(currentDoc) {
		return delimiterActionFrontmatterOpen
	}

	// If we're inside frontmatter, this closes it
	if s.inFrontmatter {
		return delimiterActionFrontmatterClose
	}

	// After frontmatter is closed, we need to distinguish:
	// 1. Document separator (YAML stream or concatenated markdown)
	// 2. Literal horizontal rule (single markdown document)

	// If we have non-frontmatter content (with or without frontmatter)
	if len(currentDoc) > 0 {
		// Check if content looks like YAML - if so, this is likely a YAML stream separator
		if looksLikeYAMLContent(currentDoc) {
			return delimiterActionDocumentSeparator
		}

		// For markdown, look ahead to see if a new document starts after this delimiter
		// A new document would start with either:
		// 1. Another frontmatter block (--- followed by YAML)
		// 2. Substantial content that looks like a new document
		if looksLikeDocumentBoundary(allLines, currentIdx) {
			return delimiterActionDocumentSeparator
		}

		// Otherwise, this is a literal horizontal rule in a single markdown document
		return delimiterActionLiteral
	}

	// Default: treat as literal content
	return delimiterActionLiteral
}

// reset prepares state for parsing a new document.
func (s *splitState) reset() {
	s.atDocumentStart = true
	s.inFrontmatter = false
	s.frontmatterClosed = false
	// Note: inCodeBlock should never be true when starting a new doc
}

// onlyEmptyLines checks if all lines are empty or whitespace.
func onlyEmptyLines(lines [][]byte) bool {
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) > 0 {
			return false
		}
	}
	return true
}

// looksLikeYAMLContent performs a quick heuristic check to see if content is YAML.
func looksLikeYAMLContent(lines [][]byte) bool {
	yamlPatterns := 0
	nonEmptyLines := 0

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		// Skip frontmatter delimiters themselves
		if isFrontmatterDelimiter(line) {
			continue
		}

		nonEmptyLines++

		// YAML patterns: "key: value" or "- item"
		if bytes.Contains(trimmed, []byte(": ")) || bytes.HasPrefix(trimmed, []byte("- ")) {
			yamlPatterns++
		}
	}

	// If most non-empty lines look like YAML, it's probably YAML
	if nonEmptyLines > 0 {
		return yamlPatterns*2 > nonEmptyLines
	}

	return false
}

// looksLikeDocumentBoundary performs lookahead to determine if a "---" delimiter
// separates two documents or is just a horizontal rule in a single document.
//
// Returns true if the content after the delimiter looks like a new document starting:
//   - Another frontmatter block (--- followed by YAML key: value lines)
//   - Empty lines followed by substantial content (indicating concatenated markdown)
func looksLikeDocumentBoundary(allLines [][]byte, delimiterIdx int) bool {
	// Look ahead at the next few lines after this delimiter
	lookAheadStart := delimiterIdx + 1
	lookAheadLimit := delimiterIdx + 10 // Check up to 10 lines ahead
	if lookAheadLimit > len(allLines) {
		lookAheadLimit = len(allLines)
	}

	// Skip empty lines immediately after delimiter
	contentStart := lookAheadStart
	for contentStart < lookAheadLimit && len(bytes.TrimSpace(allLines[contentStart])) == 0 {
		contentStart++
	}

	// If we've reached the end, it's not a boundary
	if contentStart >= len(allLines) {
		return false
	}

	// Check if next non-empty line starts a new frontmatter block
	firstContentLine := allLines[contentStart]
	if isFrontmatterDelimiter(firstContentLine) {
		// Next content is another frontmatter block → this is a document boundary
		return true
	}

	// Check if the upcoming content looks like a YAML document
	upcomingLines := allLines[contentStart:lookAheadLimit]
	if looksLikeYAMLContent(upcomingLines) {
		// Next content is YAML → this is likely a YAML stream separator
		return true
	}

	// Otherwise, this is likely a horizontal rule in a single markdown document
	return false
}
