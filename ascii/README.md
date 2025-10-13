# ASCII Art Library

Gofulmen's `ascii` package provides utilities for creating formatted ASCII art and terminal output. It's designed for CLI tools, status displays, and any application that needs consistent, visually appealing text formatting in terminal environments.

## Purpose

Terminal output formatting is essential for CLI applications to provide clear, readable information to users. The `ascii` library addresses common formatting needs by providing:

- **Box drawing**: Unicode box-drawing characters for structured output
- **Text alignment**: Proper padding and alignment within boxes
- **Unicode safety**: Correct handling of multi-byte characters
- **Consistent formatting**: Standardized appearance across different terminals with terminal-specific overrides
- **Simple API**: Easy-to-use functions for common formatting tasks

## Key Features

- **Unicode box drawing**: Proper box-drawing characters (┌┐└┘─│)
- **Automatic alignment**: Content is properly centered and padded
- **Multi-line support**: Handles multiple lines of text within boxes
- **Unicode-aware**: Correctly handles emojis, accented characters, and multi-byte sequences
- **Terminal-specific width handling**: Adapts to different terminal emoji rendering behaviors
- **Automated calibration**: Tools to automatically detect and fix width issues
- **Interactive calibration**: Manual calibration for fine-tuning terminal configurations

## Basic Usage

### Drawing ASCII Boxes

```go
package main

import (
    "fmt"
    "github.com/fulmenhq/gofulmen/ascii"
)

func main() {
    // Simple single-line box
    lines := []string{"Hello, World!"}
    ascii.DrawBox(lines)

    // Multi-line box with various content
    lines = []string{
        "Welcome to gofulmen",
        "",
        "ASCII Art Library Demo",
        "Unicode: 🌟 αβγδε",
    }
    ascii.DrawBox(lines)
}
```

Output:

```
┌─────────────────┐
│ Hello, World!   │
└─────────────────┘

┌──────────────────────┐
│ Welcome to gofulmen  │
│                      │
│ ASCII Art Library Demo│
│ Unicode: 🌟 αβγδε     │
└──────────────────────┘
```

### Creating Titled Boxes

```go
package main

import (
    "fmt"
    "strings"
    "github.com/fulmenhq/gofulmen/ascii"
)

func createTitledBox(title, content string, icon string) {
    // Create a separator line matching title width
    titleLine := fmt.Sprintf("%s %s", icon, title)
    separator := strings.Repeat("═", len([]rune(titleLine))+2) // Account for padding

    lines := []string{
        titleLine,
        separator,
        content,
    }
    ascii.DrawBox(lines)
}

func main() {
    // Simple titled box
    createTitledBox("Database Status", "✅ Connected to PostgreSQL 15.3", "🗄️")
}
```

## API Reference

### ascii.DrawBox(lines []string)

Draws a properly aligned ASCII box around the given lines of text.

**Parameters:**

- `lines`: Slice of strings to display within the box

**Behavior:**

- Trims trailing spaces from each line
- Calculates maximum line length for consistent box width
- Adds appropriate padding (2 spaces on each side)
- Uses Unicode box-drawing characters
- Prints directly to stdout

### ascii.StringWidth(s string) int

Calculates the display width of a string, accounting for Unicode characters.

**Parameters:**

- `s`: The string to measure

**Returns:**

- The visual width of the string in terminal columns

### ascii.Analyze(s string) StringAnalysis

Provides analysis of a string's properties.

**Parameters:**

- `s`: The string to analyze

**Returns:**

- `StringAnalysis` struct with length, width, unicode detection, etc.

## Terminal Compatibility

The ASCII library includes terminal-specific overrides for optimal rendering across different terminal emulators:

### Built-in Support

- **Ghostty**: Width overrides for emoji rendering (⚠️, ☠️, 🛠️, etc.)
- **iTerm2**: Width overrides for emoji rendering
- **macOS Terminal**: Standard Unicode width handling

### Custom Overrides

You can provide custom terminal overrides in `~/.config/fulmen/terminal-overrides.yaml`:

```yaml
version: "1.0.0"
terminals:
  ghostty:
    name: "Ghostty"
    overrides:
      "⚠️": 2
      "🔧": 2
    notes: "Custom overrides for my Ghostty config"
```

The library automatically detects your terminal via `$TERM_PROGRAM` and applies the appropriate overrides.

## Testing

The package includes comprehensive unit tests covering box drawing, string width calculations, and Unicode handling.

```bash
go test ./ascii/...
```
