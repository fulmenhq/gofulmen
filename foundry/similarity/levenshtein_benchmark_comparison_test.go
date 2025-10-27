package similarity

// ═══════════════════════════════════════════════════════════════════════════════
// TEMPORARY FILE: Phase 1a Levenshtein Benchmark Comparison
// ═══════════════════════════════════════════════════════════════════════════════
// This file compares our current Levenshtein implementation against matchr library.
//
// Purpose: Data-driven decision on whether to keep our implementation or use matchr
// Criteria: Accuracy (must match fixtures), Performance (speed + allocations), Maintenance
//
// This file will be DELETED after benchmarking, but kept in git history for reference.
// See: .plans/active/v0.1.5/similarity-v2-fixtures.md Phase 1a
// ═══════════════════════════════════════════════════════════════════════════════

import (
	"testing"

	"github.com/antzucaro/matchr"
)

// Current implementation (from similarity.go)
func levenshteinCurrent(a, b string) int {
	return Distance(a, b)
}

// Matchr implementation
func levenshteinMatchr(a, b string) int {
	return matchr.Levenshtein(a, b)
}

// TestLevenshtein_AccuracyComparison verifies both implementations match fixtures
func TestLevenshtein_AccuracyComparison(t *testing.T) {
	fixtures := loadFixtures(t)

	var mismatchCount int
	for _, group := range fixtures.TestCases {
		if group.Category != "levenshtein" {
			continue
		}

		for _, tc := range group.Cases {
			current := levenshteinCurrent(tc.InputA, tc.InputB)
			matchrResult := levenshteinMatchr(tc.InputA, tc.InputB)

			// Both must match expected
			if current != tc.ExpectedDistance {
				t.Errorf("CURRENT: Distance(%q, %q) = %d, want %d",
					tc.InputA, tc.InputB, current, tc.ExpectedDistance)
			}
			if matchrResult != tc.ExpectedDistance {
				t.Errorf("MATCHR: Distance(%q, %q) = %d, want %d",
					tc.InputA, tc.InputB, matchrResult, tc.ExpectedDistance)
			}

			// They must match each other
			if current != matchrResult {
				mismatchCount++
				t.Errorf("MISMATCH: Distance(%q, %q): current=%d, matchr=%d",
					tc.InputA, tc.InputB, current, matchrResult)
			}
		}
	}

	if mismatchCount > 0 {
		t.Errorf("Found %d mismatches between current and matchr implementations", mismatchCount)
	} else {
		t.Logf("✓ Both implementations match all %d fixtures exactly", 10)
	}
}

// BenchmarkLevenshtein_Current benchmarks our current implementation
func BenchmarkLevenshtein_Current(b *testing.B) {
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{"short_identical", "hello", "hello"},
		{"short_different", "kitten", "sitting"},
		{"medium_ascii", "the quick brown fox jumps over the lazy dog", "the quick brown fox jumped over the lazy dog"},
		{"medium_unicode", "café-zürich-naïve", "cafe-zurich-naive"},
		{"long_100chars", longString(100, 'a'), longString(100, 'b')},
		{"long_1000chars", longString(1000, 'a'), longString(1000, 'b')},
		{"emoji", "Hello🔥World🎉", "HelloWorld"},
		{"cjk", "日本語文字列テスト", "日本文字列テスト"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = levenshteinCurrent(tc.a, tc.b)
			}
		})
	}
}

// BenchmarkLevenshtein_Matchr benchmarks matchr library
func BenchmarkLevenshtein_Matchr(b *testing.B) {
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{"short_identical", "hello", "hello"},
		{"short_different", "kitten", "sitting"},
		{"medium_ascii", "the quick brown fox jumps over the lazy dog", "the quick brown fox jumped over the lazy dog"},
		{"medium_unicode", "café-zürich-naïve", "cafe-zurich-naive"},
		{"long_100chars", longString(100, 'a'), longString(100, 'b')},
		{"long_1000chars", longString(1000, 'a'), longString(1000, 'b')},
		{"emoji", "Hello🔥World🎉", "HelloWorld"},
		{"cjk", "日本語文字列テスト", "日本文字列テスト"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = levenshteinMatchr(tc.a, tc.b)
			}
		})
	}
}

// Helper to create long test strings
func longString(length int, char rune) string {
	s := make([]rune, length)
	for i := range s {
		s[i] = char
	}
	return string(s)
}
