package docscribe

import (
	"fmt"
	"strings"
)

// ParseError represents an error that occurred while parsing document content.
// This is typically returned when frontmatter YAML is malformed or when
// document structure cannot be parsed correctly.
type ParseError struct {
	// Message describes what went wrong
	Message string

	// LineNumber is the 1-based line number where the error occurred (0 if unknown)
	LineNumber int

	// Column is the 1-based column number where the error occurred (0 if unknown)
	Column int

	// Context provides a snippet of the problematic content
	Context string

	// Underlying is the wrapped error that caused this parse failure
	Underlying error
}

func (e *ParseError) Error() string {
	var sb strings.Builder

	sb.WriteString("parse error")

	if e.LineNumber > 0 {
		if e.Column > 0 {
			sb.WriteString(fmt.Sprintf(" at line %d, column %d", e.LineNumber, e.Column))
		} else {
			sb.WriteString(fmt.Sprintf(" at line %d", e.LineNumber))
		}
	}

	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}

	if e.Context != "" {
		sb.WriteString("\n  context: ")
		sb.WriteString(e.Context)
	}

	if e.Underlying != nil {
		sb.WriteString("\n  cause: ")
		sb.WriteString(e.Underlying.Error())
	}

	return sb.String()
}

func (e *ParseError) Unwrap() error {
	return e.Underlying
}

// FormatError represents an error when content doesn't match the expected format.
// This is returned when attempting to process content as a specific format
// (e.g., trying to parse JSON that isn't valid JSON).
type FormatError struct {
	// Expected describes the format that was expected
	Expected string

	// Actual describes what was detected or found
	Actual string

	// Message provides additional context about the format mismatch
	Message string

	// Underlying is the wrapped error if one exists
	Underlying error
}

func (e *FormatError) Error() string {
	var sb strings.Builder

	sb.WriteString("format error")

	if e.Expected != "" || e.Actual != "" {
		sb.WriteString(": expected ")
		if e.Expected != "" {
			sb.WriteString(e.Expected)
		} else {
			sb.WriteString("unknown")
		}

		if e.Actual != "" {
			sb.WriteString(", got ")
			sb.WriteString(e.Actual)
		}
	}

	if e.Message != "" {
		if e.Expected != "" || e.Actual != "" {
			sb.WriteString(" - ")
		} else {
			sb.WriteString(": ")
		}
		sb.WriteString(e.Message)
	}

	if e.Underlying != nil {
		sb.WriteString("\n  cause: ")
		sb.WriteString(e.Underlying.Error())
	}

	return sb.String()
}

func (e *FormatError) Unwrap() error {
	return e.Underlying
}

// newParseError creates a ParseError with the given message.
func newParseError(message string) *ParseError {
	return &ParseError{
		Message: message,
	}
}

// newParseErrorWithLine creates a ParseError with line number context.
func newParseErrorWithLine(message string, lineNumber int) *ParseError {
	return &ParseError{
		Message:    message,
		LineNumber: lineNumber,
	}
}

// newParseErrorWithLocation creates a ParseError with line and column context.
func newParseErrorWithLocation(message string, lineNumber, column int) *ParseError {
	return &ParseError{
		Message:    message,
		LineNumber: lineNumber,
		Column:     column,
	}
}

// wrapParseError wraps an underlying error as a ParseError.
func wrapParseError(message string, err error) *ParseError {
	// Try to extract line number from underlying error
	lineNumber := extractLineNumberFromError(err)

	return &ParseError{
		Message:    message,
		LineNumber: lineNumber,
		Underlying: err,
	}
}

// newFormatError creates a FormatError with expected and actual formats.
func newFormatError(expected, actual, message string) *FormatError {
	return &FormatError{
		Expected: expected,
		Actual:   actual,
		Message:  message,
	}
}

// wrapFormatError wraps an underlying error as a FormatError.
func wrapFormatError(expected, actual string, err error) *FormatError {
	return &FormatError{
		Expected:   expected,
		Actual:     actual,
		Underlying: err,
	}
}

// extractLineNumberFromError attempts to extract a line number from an error message.
// This is useful for YAML parsing errors that include line information.
// Returns 0 if no line number can be extracted.
func extractLineNumberFromError(err error) int {
	if err == nil {
		return 0
	}

	errStr := err.Error()

	// yaml.v3 errors often include line numbers in various formats:
	// - "line 5: ..."
	// - "yaml: line 5: ..."
	// Try to extract the line number
	if strings.Contains(errStr, "line ") {
		var lineNum int
		// Try multiple parsing patterns
		if n, _ := fmt.Sscanf(errStr, "yaml: line %d:", &lineNum); n > 0 {
			return lineNum
		}
		if n, _ := fmt.Sscanf(errStr, "line %d:", &lineNum); n > 0 {
			return lineNum
		}
		// Look for "line X" pattern anywhere in the string
		parts := strings.Split(errStr, "line ")
		if len(parts) > 1 {
			if n, _ := fmt.Sscanf(parts[1], "%d", &lineNum); n > 0 {
				return lineNum
			}
		}
	}

	return 0
}
