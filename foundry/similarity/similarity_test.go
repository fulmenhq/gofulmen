package similarity

import (
	"testing"
)

// TestDistance_Basic tests basic ASCII string distance calculations
func TestDistance_Basic(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"empty strings", "", "", 0},
		{"identical", "test", "test", 0},
		{"empty vs non-empty", "", "hello", 5},
		{"kitten to sitting", "kitten", "sitting", 3},
		{"saturday to sunday", "saturday", "sunday", 3},
		{"book to back", "book", "back", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Distance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

// TestDistance_Unicode tests Unicode string distance calculations
func TestDistance_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"accented characters", "café", "cafe", 1},
		{"diacritic difference", "naïve", "naive", 1},
		{"emoji difference", "🎉", "🎊", 1},
		{"hello with emoji", "hello😀", "hello😃", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Distance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

// TestDistance_Symmetry verifies that Distance(a, b) == Distance(b, a)
func TestDistance_Symmetry(t *testing.T) {
	tests := []struct {
		a string
		b string
	}{
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"café", "cafe"},
		{"hello", ""},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			distAB := Distance(tt.a, tt.b)
			distBA := Distance(tt.b, tt.a)
			if distAB != distBA {
				t.Errorf("Distance not symmetric: Distance(%q, %q) = %d, Distance(%q, %q) = %d",
					tt.a, tt.b, distAB, tt.b, tt.a, distBA)
			}
		})
	}
}

// TestScore_Basic tests basic score calculations
func TestScore_Basic(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
		delta    float64 // Tolerance for floating point comparison
	}{
		{"empty strings", "", "", 1.0, 0.0001},
		{"identical", "test", "test", 1.0, 0.0001},
		{"empty vs non-empty", "", "hello", 0.0, 0.0001},
		{"kitten to sitting", "kitten", "sitting", 0.5714285714285714, 0.0001},
		{"saturday to sunday", "saturday", "sunday", 0.625, 0.0001},
		{"book to back", "book", "back", 0.5, 0.0001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.a, tt.b)
			if !floatNearlyEqual(got, tt.expected, tt.delta) {
				t.Errorf("Score(%q, %q) = %f, want %f (±%f)", tt.a, tt.b, got, tt.expected, tt.delta)
			}
		})
	}
}

// TestScore_Unicode tests Unicode score calculations
func TestScore_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
		delta    float64
	}{
		{"accented characters", "café", "cafe", 0.75, 0.0001},
		{"diacritic difference", "naïve", "naive", 0.8, 0.0001},
		{"ASCII vs accented", "hello", "hëllo", 0.8, 0.0001},
		{"different emoji", "🎉", "🎊", 0.0, 0.0001},
		{"emoji suffix", "hello😀", "hello😃", 0.8333333333333334, 0.0001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.a, tt.b)
			if !floatNearlyEqual(got, tt.expected, tt.delta) {
				t.Errorf("Score(%q, %q) = %f, want %f (±%f)", tt.a, tt.b, got, tt.expected, tt.delta)
			}
		})
	}
}

// TestScore_Range verifies scores are in [0.0, 1.0] range
func TestScore_Range(t *testing.T) {
	tests := []struct {
		a string
		b string
	}{
		{"", ""},
		{"hello", "hello"},
		{"hello", "world"},
		{"", "test"},
		{"test", ""},
		{"café", "coffee"},
		{"🎉🎊", "🎈🎁"},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			score := Score(tt.a, tt.b)
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score(%q, %q) = %f, must be in [0.0, 1.0]", tt.a, tt.b, score)
			}
		})
	}
}

// TestScore_Symmetry verifies that Score(a, b) == Score(b, a)
func TestScore_Symmetry(t *testing.T) {
	tests := []struct {
		a string
		b string
	}{
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"café", "cafe"},
		{"hello", "world"},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			scoreAB := Score(tt.a, tt.b)
			scoreBA := Score(tt.b, tt.a)
			if !floatNearlyEqual(scoreAB, scoreBA, 0.0001) {
				t.Errorf("Score not symmetric: Score(%q, %q) = %f, Score(%q, %q) = %f",
					tt.a, tt.b, scoreAB, tt.b, tt.a, scoreBA)
			}
		})
	}
}

// TestScore_EdgeCases tests additional edge cases for Score
func TestScore_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
		delta    float64
	}{
		// Cases where b is longer than a (exercises lenB > lenA branch)
		{"short vs long", "abc", "abcdef", 0.5, 0.0001},
		{"empty vs long", "", "longstring", 0.0, 0.0001},
		{"one vs many", "a", "abcdefghij", 0.1, 0.0001}, // distance=9, max=10, score=1-9/10=0.1

		// Cases where a is longer than b
		{"long vs short", "abcdef", "abc", 0.5, 0.0001},
		{"long vs empty", "longstring", "", 0.0, 0.0001},
		{"many vs one", "abcdefghij", "a", 0.1, 0.0001}, // distance=9, max=10, score=1-9/10=0.1

		// Single character cases
		{"single identical", "a", "a", 1.0, 0.0001},
		{"single different", "a", "b", 0.0, 0.0001},

		// Very different strings
		{"completely different", "abc", "xyz", 0.0, 0.0001},
		{"no overlap", "hello", "12345", 0.0, 0.0001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.a, tt.b)
			if !floatNearlyEqual(got, tt.expected, tt.delta) {
				t.Errorf("Score(%q, %q) = %f, want %f (±%f)", tt.a, tt.b, got, tt.expected, tt.delta)
			}
		})
	}
}

// TestDistance_EdgeCases tests additional edge cases for Distance
func TestDistance_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		// Test swap path: ensure b is shorter, so algorithm swaps
		{"force swap - a much longer", "abcdefghij", "abc", 7},
		{"force swap - single char vs many", "abcdefghij", "x", 10},
		{"force swap - empty vs long", "verylongstring", "", 14},

		// All operations
		{"only insertions", "abc", "abcxyz", 3},
		{"only deletions", "abcxyz", "abc", 3},
		{"mixed operations", "kitten", "sitting", 3},

		// Prefix/suffix cases
		{"common prefix", "abcdef", "abcxyz", 3},
		{"common suffix", "xyzdef", "abcdef", 3},
		{"prefix match", "hello", "helloworld", 5},
		{"suffix match", "world", "helloworld", 5},

		// Single vs multiple
		{"single to double", "a", "aa", 1},
		{"double to single", "aa", "a", 1},
		{"single to triple", "a", "aaa", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Distance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

// TestDistance_LongStrings tests with longer strings
func TestDistance_LongStrings(t *testing.T) {
	// Build long strings of repeated characters
	longA := ""
	for i := 0; i < 100; i++ {
		longA += "a"
	}
	longB := ""
	for i := 0; i < 100; i++ {
		longB += "b"
	}

	// Completely different long strings
	dist := Distance(longA, longB)
	if dist != 100 {
		t.Errorf("Distance between 100-char different strings = %d, want 100", dist)
	}

	// One character different
	almostSame := longA[:99] + "b"
	dist = Distance(longA, almostSame)
	if dist != 1 {
		t.Errorf("Distance with 1 char diff in 100-char strings = %d, want 1", dist)
	}

	// Test with Unicode long strings
	longUnicode := ""
	for i := 0; i < 50; i++ {
		longUnicode += "café"
	}
	longUnicodeDiff := ""
	for i := 0; i < 50; i++ {
		longUnicodeDiff += "cafe"
	}
	dist = Distance(longUnicode, longUnicodeDiff)
	if dist != 50 { // 50 accent differences
		t.Errorf("Distance between Unicode long strings = %d, want 50", dist)
	}
}

// Helper function for floating point comparison with tolerance
func floatNearlyEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= epsilon
}
