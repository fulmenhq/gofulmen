package similarity

import (
	"testing"
)

// TestNormalize_Whitespace tests whitespace trimming
func TestNormalize_Whitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     NormalizeOptions
		expected string
	}{
		{"leading spaces", "  hello", NormalizeOptions{}, "hello"},
		{"trailing spaces", "hello  ", NormalizeOptions{}, "hello"},
		{"both sides", "  hello  ", NormalizeOptions{}, "hello"},
		{"tabs and newlines", "\t\nhello\n\t", NormalizeOptions{}, "hello"},
		{"multiple types", " \t\n hello \n\t ", NormalizeOptions{}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input, tt.opts)
			if got != tt.expected {
				t.Errorf("Normalize(%q, %+v) = %q, want %q", tt.input, tt.opts, got, tt.expected)
			}
		})
	}
}

// TestNormalize_CaseFolding tests case folding
func TestNormalize_CaseFolding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     NormalizeOptions
		expected string
	}{
		{"simple lowercase", "Hello World", NormalizeOptions{}, "hello world"},
		{"all uppercase", "SCREAMING", NormalizeOptions{}, "screaming"},
		{"mixed case", "MixedCaseIdentifier", NormalizeOptions{}, "mixedcaseidentifier"},
		{"already lowercase", "lowercase", NormalizeOptions{}, "lowercase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input, tt.opts)
			if got != tt.expected {
				t.Errorf("Normalize(%q, %+v) = %q, want %q", tt.input, tt.opts, got, tt.expected)
			}
		})
	}
}

// TestNormalize_AccentStripping tests accent removal
func TestNormalize_AccentStripping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     NormalizeOptions
		expected string
	}{
		{"acute accent", "café", NormalizeOptions{StripAccents: true}, "cafe"},
		{"diaeresis", "naïve", NormalizeOptions{StripAccents: true}, "naive"},
		{"umlaut with case", "Zürich", NormalizeOptions{StripAccents: true}, "zurich"},
		{"multiple accents", "résumé", NormalizeOptions{StripAccents: true}, "resume"},
		{"preserve when false", "café", NormalizeOptions{StripAccents: false}, "café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input, tt.opts)
			if got != tt.expected {
				t.Errorf("Normalize(%q, %+v) = %q, want %q", tt.input, tt.opts, got, tt.expected)
			}
		})
	}
}

// TestNormalize_Combined tests combined operations
func TestNormalize_Combined(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     NormalizeOptions
		expected string
	}{
		{
			"trim + casefold + strip",
			"  Café Münchën  ",
			NormalizeOptions{StripAccents: true},
			"cafe munchen",
		},
		{
			"whitespace + case only",
			"  HELLO WORLD  ",
			NormalizeOptions{StripAccents: false},
			"hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.input, tt.opts)
			if got != tt.expected {
				t.Errorf("Normalize(%q, %+v) = %q, want %q", tt.input, tt.opts, got, tt.expected)
			}
		})
	}
}

// TestCasefold_Simple tests simple case folding
func TestCasefold_Simple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		locale   string
		expected string
	}{
		{"simple hello", "Hello", "", "hello"},
		{"all caps", "HELLO", "", "hello"},
		{"mixed", "HeLLo WoRLd", "", "hello world"},
		{"already lower", "hello", "", "hello"},
		{"empty string", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Casefold(tt.input, tt.locale)
			if got != tt.expected {
				t.Errorf("Casefold(%q, %q) = %q, want %q", tt.input, tt.locale, got, tt.expected)
			}
		})
	}
}

// TestCasefold_Turkish tests Turkish locale special cases
func TestCasefold_Turkish(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		locale   string
		expected string
	}{
		{"dotted I", "İstanbul", "tr", "istanbul"},
		{"dotless I", "TITLE", "tr", "tıtle"},
		{"mixed text", "İzmir ISTANBUL", "tr", "izmir ıstanbul"},
		{"TR uppercase", "İstanbul", "TR", "istanbul"}, // TR also works
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Casefold(tt.input, tt.locale)
			if got != tt.expected {
				t.Errorf("Casefold(%q, %q) = %q, want %q", tt.input, tt.locale, got, tt.expected)
			}
		})
	}
}

// TestStripAccents_Basic tests basic accent stripping
func TestStripAccents_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"acute", "café", "cafe"},
		{"grave", "où", "ou"},
		{"circumflex", "château", "chateau"},
		{"diaeresis", "naïve", "naive"},
		{"umlaut", "Zürich", "Zurich"},
		{"tilde", "mañana", "manana"},
		{"multiple", "résumé", "resume"},
		{"no accents", "hello", "hello"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripAccents(tt.input)
			if got != tt.expected {
				t.Errorf("StripAccents(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestStripAccents_Complex tests more complex cases
func TestStripAccents_Complex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"french sentence", "Très bien, merci!", "Tres bien, merci!"},
		{"german", "Schön über Äpfel", "Schon uber Apfel"},
		{"spanish", "Años señor niño", "Anos senor nino"},
		{"mixed", "café naïve Zürich", "cafe naive Zurich"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripAccents(tt.input)
			if got != tt.expected {
				t.Errorf("StripAccents(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestEqualsIgnoreCase tests case-insensitive comparison
func TestEqualsIgnoreCase(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		opts     NormalizeOptions
		expected bool
	}{
		{"same case", "hello", "hello", NormalizeOptions{}, true},
		{"different case", "Hello", "hello", NormalizeOptions{}, true},
		{"all caps vs lower", "HELLO", "hello", NormalizeOptions{}, true},
		{"different strings", "hello", "world", NormalizeOptions{}, false},
		{"with accents preserved", "café", "Café", NormalizeOptions{StripAccents: false}, true},
		{"with accents stripped", "café", "cafe", NormalizeOptions{StripAccents: true}, true},
		{"accents mismatch", "café", "cafe", NormalizeOptions{StripAccents: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EqualsIgnoreCase(tt.a, tt.b, tt.opts)
			if got != tt.expected {
				t.Errorf("EqualsIgnoreCase(%q, %q, %+v) = %v, want %v",
					tt.a, tt.b, tt.opts, got, tt.expected)
			}
		})
	}
}
