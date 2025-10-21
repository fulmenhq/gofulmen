package docscribe

import (
	"bytes"
)

// InspectDocument performs fast document analysis without full parsing.
// This is optimized for speed (<1ms target for 100KB documents) and provides
// quick insights into document structure and characteristics.
//
// The inspection includes:
//   - Frontmatter detection (presence check only, no parsing)
//   - Header counting (how many headers exist)
//   - Format detection (markdown, yaml, json, etc.)
//   - Line counting
//   - Section estimation (based on header hierarchy)
//
// This function does not parse frontmatter YAML or extract full header details.
// For complete parsing, use ParseFrontmatter or ExtractHeaders instead.
//
// Example:
//
//	info, err := documentation.InspectDocument(content)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Format: %s\n", info.Format)
//	fmt.Printf("Has frontmatter: %v\n", info.HasFrontmatter)
//	fmt.Printf("Headers: %d, Estimated sections: %d\n",
//	    info.HeaderCount, info.EstimatedSections)
//
// Returns DocumentInfo with inspection results, or an error if content cannot be processed.
func InspectDocument(content []byte) (*DocumentInfo, error) {
	info := &DocumentInfo{}

	// 1. Detect format (uses existing heuristics)
	info.Format = DetectFormat(content)

	// 2. Check for frontmatter (fast check without parsing)
	info.HasFrontmatter = hasFrontmatter(content)

	// 3. Count lines
	info.LineCount = countLines(content)

	// 4. Quick header count and section estimation
	// Only do this for markdown content
	if info.Format == FormatMarkdown || info.Format == FormatMultiMarkdown {
		headerCount, sectionCount := analyzeHeaderStructure(content)
		info.HeaderCount = headerCount
		info.EstimatedSections = sectionCount
	} else {
		info.HeaderCount = 0
		info.EstimatedSections = 0
	}

	return info, nil
}

// countLines counts the number of lines in the content.
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}

	// Count newlines
	count := bytes.Count(content, []byte("\n"))

	// If content doesn't end with newline, add 1 for the last line
	if len(content) > 0 && content[len(content)-1] != '\n' {
		count++
	}

	return count
}

// analyzeHeaderStructure performs a fast header analysis without full parsing.
// Returns (headerCount, estimatedSections).
//
// Section estimation logic:
//   - H1 headers typically denote major sections
//   - H2 headers under H1s are subsections (count as separate sections if substantial)
//   - Estimate is conservative: count H1s + significant H2s
func analyzeHeaderStructure(content []byte) (int, int) {
	lines := bytes.Split(content, []byte("\n"))

	headerCount := 0
	h1Count := 0
	h2Count := 0
	inCodeBlock := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Track code blocks
		if isCodeBlockFence(line) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// Check for ATX headers (fast check without full regex)
		trimmed := bytes.TrimLeft(line, " \t")
		if len(trimmed) > 0 && trimmed[0] == '#' {
			// Count leading # symbols
			level := 0
			for j := 0; j < len(trimmed) && trimmed[j] == '#'; j++ {
				level++
			}

			// Must have space after # and be valid header (1-6 levels)
			if level >= 1 && level <= 6 && len(trimmed) > level && trimmed[level] == ' ' {
				headerCount++
				switch level {
				case 1:
					h1Count++
				case 2:
					h2Count++
				}
			}
		}

		// Check for Setext headers (need to look at next line)
		if i+1 < len(lines) {
			text := bytes.TrimSpace(line)
			nextLine := bytes.TrimSpace(lines[i+1])

			if len(text) > 0 && len(nextLine) > 0 {
				// Check if next line is all = or all -
				if isSetextUnderlineFast(nextLine, '=') {
					headerCount++
					h1Count++
					i++ // Skip underline
				} else if isSetextUnderlineFast(nextLine, '-') {
					headerCount++
					h2Count++
					i++ // Skip underline
				}
			}
		}
	}

	// Estimate sections:
	// - Each H1 is a major section
	// - H2s are subsections, but only count them as separate sections if there are many
	// - Conservative estimate: H1s + (significant H2s / 3)
	estimatedSections := h1Count
	if h2Count > 0 {
		// Add some H2s as sections (assume every 3 H2s represents a major subsection)
		estimatedSections += (h2Count + 2) / 3
	}

	// If no H1s but has H2s, use H2s as section markers
	if h1Count == 0 && h2Count > 0 {
		estimatedSections = h2Count
	}

	// Ensure at least 1 section if we found headers
	if estimatedSections == 0 && headerCount > 0 {
		estimatedSections = 1
	}

	return headerCount, estimatedSections
}

// isSetextUnderlineFast is a faster version of isSetextUnderline for inspection.
// It does a simple check without full validation.
func isSetextUnderlineFast(line []byte, char byte) bool {
	if len(line) == 0 {
		return false
	}

	// Must have at least one of the target character
	hasChar := false
	for _, b := range line {
		if b == char {
			hasChar = true
		} else if b != ' ' && b != '\t' {
			// Found non-whitespace, non-target character
			return false
		}
	}

	return hasChar
}
