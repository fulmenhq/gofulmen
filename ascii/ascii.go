package ascii

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

type BoxChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	Cross       string
}

func DefaultBoxChars() BoxChars {
	return BoxChars{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Cross:       "┼",
	}
}

// BoxOptions configures box drawing behavior
type BoxOptions struct {
	MinWidth int       // Minimum width (default: 0, uses content width)
	MaxWidth int       // Maximum width (default: 0, unlimited). Content exceeding this will cause error.
	Chars    *BoxChars // Custom box characters (default: DefaultBoxChars())
}

// DrawBox draws a box around the given content with a minimum width.
// The box will expand to fit content, but never be smaller than the width parameter.
//
// Parameters:
//   - content: Multi-line string to box
//   - width: Minimum width (0 = auto-size to content)
//
// Example:
//
//	box := DrawBox("Hello\nWorld", 20)  // Box will be at least 20 chars wide
func DrawBox(content string, width int) string {
	return DrawBoxWithOptions(content, BoxOptions{MinWidth: width})
}

// DrawBoxWithOptions draws a box with advanced options for width constraints and styling
//
// Example - Aligned boxes:
//
//	boxes := []string{"Short", "Medium length", "Very long content here"}
//	maxWidth := MaxContentWidth(boxes)
//	for _, content := range boxes {
//	    fmt.Print(DrawBoxWithOptions(content, BoxOptions{MinWidth: maxWidth}))
//	}
//
// Example - Width limits:
//
//	box := DrawBoxWithOptions(content, BoxOptions{
//	    MinWidth: 40,
//	    MaxWidth: 80,  // Error if content > 80 chars
//	})
func DrawBoxWithOptions(content string, opts BoxOptions) string {
	chars := DefaultBoxChars()
	if opts.Chars != nil {
		chars = *opts.Chars
	}

	lines := strings.Split(content, "\n")

	// Find content width
	contentWidth := 0
	for _, line := range lines {
		lineWidth := StringWidth(line)
		if lineWidth > contentWidth {
			contentWidth = lineWidth
		}
	}

	// Apply width constraints
	boxWidth := contentWidth
	if opts.MinWidth > 0 && boxWidth < opts.MinWidth {
		boxWidth = opts.MinWidth
	}
	if opts.MaxWidth > 0 && contentWidth > opts.MaxWidth {
		// Content exceeds max width - this is an error condition
		// For now, we'll just use max width and truncate, but we could panic or return error
		boxWidth = opts.MaxWidth
	}

	var result strings.Builder

	result.WriteString(chars.TopLeft)
	result.WriteString(strings.Repeat(chars.Horizontal, boxWidth+2))
	result.WriteString(chars.TopRight)
	result.WriteString("\n")

	for _, line := range lines {
		lineWidth := StringWidth(line)

		result.WriteString(chars.Vertical)
		result.WriteString(" ")

		// Truncate if exceeds max width
		if opts.MaxWidth > 0 && lineWidth > opts.MaxWidth {
			// TODO: Proper truncation respecting rune boundaries
			result.WriteString(line[:opts.MaxWidth])
			lineWidth = opts.MaxWidth
		} else {
			result.WriteString(line)
		}

		padding := boxWidth - lineWidth
		if padding > 0 {
			result.WriteString(strings.Repeat(" ", padding))
		}
		result.WriteString(" ")
		result.WriteString(chars.Vertical)
		result.WriteString("\n")
	}

	result.WriteString(chars.BottomLeft)
	result.WriteString(strings.Repeat(chars.Horizontal, boxWidth+2))
	result.WriteString(chars.BottomRight)
	result.WriteString("\n")

	return result.String()
}

// MaxContentWidth returns the maximum display width across multiple content strings
// This is useful for aligning multiple boxes to the same width
//
// Example:
//
//	contents := []string{
//	    "Layer 1: Short",
//	    "Layer 2: Medium length text",
//	    "Layer 3: Very long content here",
//	}
//	maxWidth := MaxContentWidth(contents)
//	for _, content := range contents {
//	    fmt.Print(DrawBox(content, maxWidth))
//	}
func MaxContentWidth(contents []string) int {
	maxWidth := 0
	for _, content := range contents {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			width := StringWidth(line)
			if width > maxWidth {
				maxWidth = width
			}
		}
	}
	return maxWidth
}

// StringWidth returns the display width of a string, accounting for Unicode characters
// and terminal-specific overrides
func StringWidth(s string) int {
	// If we have terminal-specific overrides, apply them
	if currentTerminalConfig != nil && len(currentTerminalConfig.Overrides) > 0 {
		baseWidth := runewidth.StringWidth(s)
		adjustment := 0

		// Check each override character
		for emoji, expectedWidth := range currentTerminalConfig.Overrides {
			count := strings.Count(s, emoji)
			if count > 0 {
				currentWidth := runewidth.StringWidth(emoji)
				adjustment += count * (expectedWidth - currentWidth)
			}
		}

		if adjustment != 0 {
			return baseWidth + adjustment
		}
	}

	// Fall back to go-runewidth
	return runewidth.StringWidth(s)
}
