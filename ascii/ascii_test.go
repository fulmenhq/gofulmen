package ascii

import (
	"strings"
	"testing"
)

func TestDrawBox(t *testing.T) {
	content := "Hello\nWorld"
	box := DrawBox(content, 10)

	lines := strings.Split(strings.TrimSpace(box), "\n")
	if len(lines) != 4 { // top, content line 1, content line 2, bottom
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	// Check that it starts and ends with box characters
	if !strings.HasPrefix(lines[0], "┌") || !strings.HasSuffix(lines[0], "┐") {
		t.Errorf("Top border incorrect: %s", lines[0])
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ASCII", "hello", 5},
		{"Spaces", "hello world", 11},
		{"Unicode", "café", 4},
		{"Emoji", "🚀", 2},    // Emojis are width 2
		{"CJK", "こんにちは", 10}, // CJK characters are width 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := StringWidth(tt.input)
			if width != tt.expected {
				t.Errorf("StringWidth(%q) = %d, expected %d", tt.input, width, tt.expected)
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	s := "Hello\nWorld Café"
	analysis := Analyze(s)

	if analysis.Length != len(s) {
		t.Errorf("Length mismatch")
	}
	if analysis.LineCount != 2 {
		t.Errorf("Expected 2 lines, got %d", analysis.LineCount)
	}
	if !analysis.HasUnicode {
		t.Errorf("Should detect Unicode")
	}
	if !analysis.IsMultiline {
		t.Errorf("Should be multiline")
	}
}

func TestTerminalOverrides(t *testing.T) {
	tests := []struct {
		name          string
		termProgram   string
		testString    string
		expectedWidth int
		hasOverride   bool
	}{
		{
			name:          "Ghostty with emoji override",
			termProgram:   "ghostty",
			testString:    "⚠️",
			expectedWidth: 2,
			hasOverride:   true,
		},
		{
			name:          "iTerm2 with emoji override",
			termProgram:   "iTerm.app",
			testString:    "☠️",
			expectedWidth: 2,
			hasOverride:   true,
		},
		{
			name:          "macOS Terminal without override",
			termProgram:   "Apple_Terminal",
			testString:    "Hello",
			expectedWidth: 5,
			hasOverride:   false,
		},
		{
			name:          "Unknown terminal fallback",
			termProgram:   "unknown",
			testString:    "test",
			expectedWidth: 4,
			hasOverride:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM_PROGRAM", tt.termProgram)

			if err := loadTerminalCatalog(); err != nil {
				t.Fatalf("Failed to load terminal catalog: %v", err)
			}
			detectCurrentTerminal()

			config := GetTerminalConfig()
			if tt.hasOverride && config == nil {
				t.Errorf("Expected terminal config for %s, got nil", tt.termProgram)
			}
			if !tt.hasOverride && config != nil && len(config.Overrides) > 0 {
				t.Errorf("Did not expect overrides for %s", tt.termProgram)
			}

			width := StringWidth(tt.testString)
			if width != tt.expectedWidth {
				t.Errorf("StringWidth(%q) = %d, expected %d", tt.testString, width, tt.expectedWidth)
			}
		})
	}
}

func TestStringWidthWithMultipleOverrides(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")

	if err := loadTerminalCatalog(); err != nil {
		t.Fatalf("Failed to load terminal catalog: %v", err)
	}
	detectCurrentTerminal()

	testString := "⚠️ Warning ☠️"
	width := StringWidth(testString)

	expectedWidth := 2 + 1 + 7 + 1 + 2
	if width != expectedWidth {
		t.Errorf("StringWidth(%q) = %d, expected %d (with Ghostty overrides)", testString, width, expectedWidth)
	}
}

func TestBYOC_SetTerminalOverrides(t *testing.T) {
	// Save original state
	originalCatalog := terminalCatalog
	defer func() {
		terminalCatalog = originalCatalog
		detectCurrentTerminal()
	}()

	// Layer 3: BYOC - External app provides custom config
	customConfig := &TerminalOverrides{
		Version: "1.0.0",
		Terminals: map[string]TerminalConfig{
			"customterm": {
				Name: "Custom Terminal",
				Overrides: map[string]int{
					"🔧": 3, // Custom width
				},
			},
		},
	}

	SetTerminalOverrides(customConfig)

	// Verify custom config is used
	configs := GetAllTerminalConfigs()
	if cfg, exists := configs["customterm"]; !exists {
		t.Error("Custom terminal config not found after SetTerminalOverrides")
	} else if cfg.Name != "Custom Terminal" {
		t.Errorf("Expected custom terminal name 'Custom Terminal', got %q", cfg.Name)
	}
}

func TestBYOC_SetTerminalConfig(t *testing.T) {
	// Save original state
	originalCatalog := terminalCatalog
	defer func() {
		terminalCatalog = originalCatalog
		detectCurrentTerminal()
	}()

	// Layer 3: BYOC - Set config for specific terminal
	SetTerminalConfig("myterm", TerminalConfig{
		Name: "My Terminal",
		Overrides: map[string]int{
			"🎯": 2,
		},
	})

	configs := GetAllTerminalConfigs()
	if cfg, exists := configs["myterm"]; !exists {
		t.Error("Custom terminal config not found after SetTerminalConfig")
	} else if cfg.Name != "My Terminal" {
		t.Errorf("Expected terminal name 'My Terminal', got %q", cfg.Name)
	}
}

func TestMaxContentWidth(t *testing.T) {
	contents := []string{
		"Short",
		"Medium length",
		"Very long content here",
	}

	maxWidth := MaxContentWidth(contents)
	expected := StringWidth("Very long content here")

	if maxWidth != expected {
		t.Errorf("MaxContentWidth() = %d, expected %d", maxWidth, expected)
	}
}

func TestMaxContentWidth_MultiLine(t *testing.T) {
	contents := []string{
		"Line 1\nLine 2",
		"Short",
		"This is a very long single line",
	}

	maxWidth := MaxContentWidth(contents)
	expected := StringWidth("This is a very long single line")

	if maxWidth != expected {
		t.Errorf("MaxContentWidth() = %d, expected %d", maxWidth, expected)
	}
}

func TestDrawBoxWithMinWidth(t *testing.T) {
	content := "Short"
	minWidth := 20

	box := DrawBox(content, minWidth)
	lines := strings.Split(strings.TrimSpace(box), "\n")

	// Top border should be minWidth + 2 (padding) + 2 (borders) = minWidth + 4 total chars
	topBorder := lines[0]
	expectedTopWidth := minWidth + 2 + 2 // content padding (2) + borders (2)

	if StringWidth(topBorder) != expectedTopWidth {
		t.Errorf("Top border width = %d, expected %d", StringWidth(topBorder), expectedTopWidth)
	}
}

func TestDrawBoxWithOptions_MinWidth(t *testing.T) {
	content := "Hi"
	opts := BoxOptions{MinWidth: 30}

	box := DrawBoxWithOptions(content, opts)
	lines := strings.Split(strings.TrimSpace(box), "\n")

	// Content line should have MinWidth worth of content space
	contentLine := lines[1]
	// Should be: │ + space + content + padding + space + │
	expectedWidth := 1 + 1 + 30 + 1 + 1 // borders + padding + minWidth

	if StringWidth(contentLine) != expectedWidth {
		t.Errorf("Content line width = %d, expected %d", StringWidth(contentLine), expectedWidth)
	}
}

func TestDrawBoxWithOptions_AlignedBoxes(t *testing.T) {
	contents := []string{
		"Short",
		"Medium length text",
		"Very long content line here",
	}

	maxWidth := MaxContentWidth(contents)

	boxes := make([]string, len(contents))
	for i, content := range contents {
		boxes[i] = DrawBox(content, maxWidth)
	}

	// All boxes should have the same width
	var widths []int
	for _, box := range boxes {
		lines := strings.Split(strings.TrimSpace(box), "\n")
		topBorder := lines[0]
		widths = append(widths, StringWidth(topBorder))
	}

	// Check all widths are equal
	firstWidth := widths[0]
	for i, width := range widths {
		if width != firstWidth {
			t.Errorf("Box %d has width %d, expected %d (all boxes should align)", i, width, firstWidth)
		}
	}
}

func TestReloadTerminalOverrides(t *testing.T) {
	// Use SetTerminalOverrides with completely custom config
	customConfig := &TerminalOverrides{
		Version: "1.0.0",
		Terminals: map[string]TerminalConfig{
			"temp": {Name: "Temporary"},
		},
	}
	SetTerminalOverrides(customConfig)

	// Verify only custom config exists
	configs := GetAllTerminalConfigs()
	if len(configs) != 1 {
		t.Errorf("Expected 1 config after SetTerminalOverrides, got %d", len(configs))
	}
	if _, exists := configs["temp"]; !exists {
		t.Error("Temporary config should exist before reload")
	}

	// Reload to reset to defaults
	if err := ReloadTerminalOverrides(); err != nil {
		t.Fatalf("Failed to reload terminal overrides: %v", err)
	}

	// After reload, should have default configs
	configs = GetAllTerminalConfigs()
	if _, exists := configs["temp"]; exists {
		t.Error("Temporary config should not exist after reload")
	}

	// Verify defaults are back (at least 3: ghostty, iTerm.app, Apple_Terminal)
	if len(configs) < 3 {
		t.Errorf("Expected at least 3 default configs after reload, got %d", len(configs))
	}
	if _, exists := configs["ghostty"]; !exists {
		t.Error("Default ghostty config should exist after reload")
	}

	// Cleanup - reload again to ensure defaults for other tests
	_ = ReloadTerminalOverrides()
}
