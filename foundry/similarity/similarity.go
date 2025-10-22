package similarity

import (
	"unicode/utf8"
)

// ═══════════════════════════════════════════════════════════════════════════════
// TODO(v0.1.4+): ADD NEW DISTANCE FUNCTIONS HERE
// ═══════════════════════════════════════════════════════════════════════════════
// Planned additions:
// - func DamerauDistance(a, b string) int
// - func JaroWinklerSimilarity(a, b string) float64
//
// See: .plans/active/v0.1.3/012-similarity-expansion-roadmap.md
// ═══════════════════════════════════════════════════════════════════════════════

// Distance calculates the Levenshtein edit distance between two strings.
//
// The distance represents the minimum number of single-character edits
// (insertions, deletions, or substitutions) required to transform string a
// into string b.
//
// This implementation uses the Wagner-Fischer dynamic programming algorithm
// with Unicode-aware character counting (grapheme clusters via rune counting).
//
// Examples:
//   - Distance("", "") returns 0 (identical empty strings)
//   - Distance("kitten", "sitting") returns 3 (3 substitutions)
//   - Distance("café", "cafe") returns 1 (1 character difference)
//
// Performance: Targets ≤0.5ms p95 for 128-character strings.
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func Distance(a, b string) int {
	// Convert to rune slices for Unicode-aware processing
	runesA := []rune(a)
	runesB := []rune(b)

	lenA := len(runesA)
	lenB := len(runesB)

	// Edge cases: empty strings
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Ensure we iterate over the shorter string for better cache locality
	// If B is shorter, swap so A is always the shorter one
	if lenB < lenA {
		runesA, runesB = runesB, runesA
		lenA, lenB = lenB, lenA
	}

	// Use two-row approach for O(min(m,n)) space complexity
	// We only need the previous row to calculate the current row
	prevRow := make([]int, lenA+1)
	currRow := make([]int, lenA+1)

	// Initialize first row: distance from empty string to prefixes of A
	for i := 0; i <= lenA; i++ {
		prevRow[i] = i
	}

	// Fill the matrix row by row
	for j := 1; j <= lenB; j++ {
		// First column: distance from empty string to prefix of B
		currRow[0] = j

		for i := 1; i <= lenA; i++ {
			// Cost of substitution: 0 if characters match, 1 otherwise
			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
			}

			// Calculate minimum of three operations:
			// 1. Deletion: currRow[i-1] + 1
			// 2. Insertion: prevRow[i] + 1
			// 3. Substitution: prevRow[i-1] + cost
			deletion := currRow[i-1] + 1
			insertion := prevRow[i] + 1
			substitution := prevRow[i-1] + cost

			// Take minimum
			currRow[i] = deletion
			if insertion < currRow[i] {
				currRow[i] = insertion
			}
			if substitution < currRow[i] {
				currRow[i] = substitution
			}
		}

		// Swap rows: current becomes previous for next iteration
		prevRow, currRow = currRow, prevRow
	}

	// Result is in the last column of the last processed row (now prevRow)
	return prevRow[lenA]
}

// Score calculates a normalized similarity score between two strings.
//
// The score is computed as: 1 - (distance / max(len(a), len(b)))
// where distance is the Levenshtein distance and length is measured in
// Unicode graphemes (runes).
//
// Returns a float64 in the range [0.0, 1.0]:
//   - 0.0 indicates completely different strings
//   - 1.0 indicates identical strings
//
// Examples:
//   - Score("", "") returns 1.0 (both empty)
//   - Score("kitten", "sitting") returns 0.5714... (1 - 3/7)
//   - Score("hello", "hello") returns 1.0 (identical)
//
// Performance: Targets ≤0.5ms p95 for 128-character strings.
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func Score(a, b string) float64 {
	// Fast path: identical strings (includes empty strings)
	if a == b {
		return 1.0
	}

	// Get lengths in runes (Unicode graphemes)
	lenA := utf8.RuneCountInString(a)
	lenB := utf8.RuneCountInString(b)

	// Get max length for normalization
	maxLen := lenA
	if lenB > maxLen {
		maxLen = lenB
	}

	// Calculate distance and normalize
	distance := Distance(a, b)
	return 1.0 - float64(distance)/float64(maxLen)
}
