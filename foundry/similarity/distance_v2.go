package similarity

import (
	"errors"
	"fmt"

	"github.com/antzucaro/matchr"
)

// Algorithm represents supported string distance and similarity algorithms.
//
// Implements Crucible similarity v2.0.0 standard with multiple algorithm variants:
// - Levenshtein: Classic edit distance (insertions, deletions, substitutions)
// - Damerau OSA: Optimal String Alignment (adds adjacent transpositions, cannot edit same substring twice)
// - Damerau Unrestricted: True Damerau-Levenshtein (unrestricted transpositions)
// - Jaro-Winkler: Similarity metric optimized for short strings with common prefixes
// - Substring: Longest common substring matching
//
// Use cases:
//   - Levenshtein: General-purpose edit distance, spell checking, diff algorithms
//   - Damerau OSA: Typo correction, CLI fuzzy matching, spell checking with transpositions
//   - Damerau Unrestricted: General similarity, DNA sequencing, complex transformations
//   - Jaro-Winkler: Name matching, record linkage, prefix-heavy matching
//   - Substring: Partial string matching, search-as-you-type, path component matching
type Algorithm string

const (
	// AlgorithmLevenshtein calculates classic edit distance.
	// Allows: insertions, deletions, substitutions
	// Use for: general edit distance, spell checking, diff algorithms
	AlgorithmLevenshtein Algorithm = "levenshtein"

	// AlgorithmDamerauOSA calculates Damerau-Levenshtein distance (OSA variant).
	// Allows: insertions, deletions, substitutions, adjacent transpositions
	// Restriction: cannot edit same substring more than once
	// Use for: typo correction, CLI fuzzy matching, common typing errors
	AlgorithmDamerauOSA Algorithm = "damerau_osa"

	// AlgorithmDamerauUnrestricted calculates unrestricted Damerau-Levenshtein distance.
	// Allows: insertions, deletions, substitutions, unrestricted transpositions
	// No OSA restriction
	// Use for: general similarity, DNA sequencing, complex string transformations
	AlgorithmDamerauUnrestricted Algorithm = "damerau_unrestricted"

	// AlgorithmJaroWinkler calculates Jaro-Winkler similarity score.
	// Optimized for short strings with common prefixes
	// Use for: name matching, record linkage, person/organization names
	AlgorithmJaroWinkler Algorithm = "jaro_winkler"

	// AlgorithmSubstring finds longest common substring.
	// Returns best substring match and score
	// Use for: partial matching, search-as-you-type, path component matching
	AlgorithmSubstring Algorithm = "substring"
)

// DistanceWithAlgorithm calculates edit distance between two strings using the specified algorithm.
//
// Returns the minimum number of single-character edits required to transform string a into string b.
// The specific operations allowed depend on the algorithm:
//   - Levenshtein: insertions, deletions, substitutions
//   - Damerau OSA: adds adjacent transpositions (optimal string alignment)
//   - Damerau Unrestricted: unrestricted transpositions
//
// For similarity-based metrics (Jaro-Winkler, substring), returns an error directing
// users to ScoreWithAlgorithm().
//
// Examples:
//
//	distance, _ := DistanceWithAlgorithm("kitten", "sitting", AlgorithmLevenshtein)
//	// Returns: 3 (3 substitutions)
//
//	distance, _ := DistanceWithAlgorithm("abcd", "abdc", AlgorithmDamerauOSA)
//	// Returns: 1 (1 transposition: cd -> dc)
//
//	distance, _ := DistanceWithAlgorithm("CA", "ABC", AlgorithmDamerauOSA)
//	// Returns: 3 (OSA restriction applies)
//
//	distance, _ := DistanceWithAlgorithm("CA", "ABC", AlgorithmDamerauUnrestricted)
//	// Returns: 2 (unrestricted allows more efficient transformation)
//
// Performance: See ADR-0002 for benchmark data. Levenshtein uses optimized native
// implementation (1.24-1.76x faster than external libraries, 3-326x less memory).
//
// Conformance: Implements Crucible Foundry Similarity Standard v2.0.0 (2025.10.3).
func DistanceWithAlgorithm(a, b string, algorithm Algorithm) (int, error) {
	// Emit telemetry: algorithm usage counter (ADR-0008 Pattern 1)
	emitAlgorithmCounter("distance", algorithm)

	// Emit telemetry: string length distribution
	emitStringLengthCounter(algorithm, a, b)

	switch algorithm {
	case AlgorithmLevenshtein:
		return levenshteinDistance(a, b), nil

	case AlgorithmDamerauOSA:
		return damerauOSADistance(a, b), nil

	case AlgorithmDamerauUnrestricted:
		return damerauUnrestrictedDistance(a, b), nil

	case AlgorithmJaroWinkler:
		// Emit telemetry: API misuse error
		emitErrorCounter("wrong_api", algorithm, "ScoreWithAlgorithm")
		return 0, errors.New(
			"jaro_winkler metric produces similarity scores, not distances. " +
				"Use ScoreWithAlgorithm(a, b, AlgorithmJaroWinkler, nil) instead",
		)

	case AlgorithmSubstring:
		// Emit telemetry: API misuse error
		emitErrorCounter("wrong_api", algorithm, "SubstringMatch")
		return 0, errors.New(
			"substring metric does not produce distances. " +
				"Use SubstringMatch(needle, haystack) instead",
		)

	default:
		return 0, fmt.Errorf(
			"invalid algorithm: %q. Valid options: %s, %s, %s",
			algorithm,
			AlgorithmLevenshtein,
			AlgorithmDamerauOSA,
			AlgorithmDamerauUnrestricted,
		)
	}
}

// ScoreOptions configures similarity score calculation.
type ScoreOptions struct {
	// JaroPrefixScale is the Jaro-Winkler prefix scaling factor.
	// Higher values give more weight to matching prefixes.
	// Standard range: 0.0-0.25, default: 0.1
	// Only used for AlgorithmJaroWinkler.
	JaroPrefixScale float64

	// JaroMaxPrefix is the maximum prefix length for Jaro-Winkler bonus.
	// Standard range: 1-8, default: 4
	// Only used for AlgorithmJaroWinkler.
	JaroMaxPrefix int
}

// DefaultScoreOptions returns default options for score calculation.
func DefaultScoreOptions() *ScoreOptions {
	return &ScoreOptions{
		JaroPrefixScale: 0.1, // Standard Jaro-Winkler default
		JaroMaxPrefix:   4,   // Standard Jaro-Winkler default
	}
}

// ScoreWithAlgorithm calculates a normalized similarity score between two strings.
//
// Returns a score in the range [0.0, 1.0]:
//   - 0.0 = completely different
//   - 1.0 = identical
//
// For distance-based metrics (Levenshtein, Damerau variants):
//
//	Formula: 1.0 - distance / max(len(a), len(b))
//
// For similarity-based metrics (Jaro-Winkler, substring):
//
//	Formula: Direct similarity calculation
//
// Examples:
//
//	score, _ := ScoreWithAlgorithm("kitten", "sitting", AlgorithmLevenshtein, nil)
//	// Returns: 0.5714285714285714 (1 - 3/7)
//
//	score, _ := ScoreWithAlgorithm("abcd", "abdc", AlgorithmDamerauOSA, nil)
//	// Returns: 0.75 (1 - 1/4)
//
//	opts := &ScoreOptions{JaroPrefixScale: 0.1, JaroMaxPrefix: 4}
//	score, _ := ScoreWithAlgorithm("martha", "marhta", AlgorithmJaroWinkler, opts)
//	// Returns: 0.9611111111111111
//
//	score, _ := ScoreWithAlgorithm("hello", "hello world", AlgorithmSubstring, nil)
//	// Returns: 0.4545454545454545
//
// Performance: Targets ≤0.5ms p95 for 128-character strings. Distance-based metrics
// benefit from optimized implementations (see ADR-0002 for benchmark data).
//
// Conformance: Implements Crucible Foundry Similarity Standard v2.0.0 (2025.10.3).
func ScoreWithAlgorithm(a, b string, algorithm Algorithm, opts *ScoreOptions) (float64, error) {
	// Emit telemetry: algorithm usage counter (ADR-0008 Pattern 1)
	emitAlgorithmCounter("score", algorithm)

	// Emit telemetry: string length distribution
	emitStringLengthCounter(algorithm, a, b)

	// Fast path: identical strings
	if a == b {
		// Emit telemetry: fast path hit
		emitFastPathCounter("identical")
		return 1.0, nil
	}

	// Get lengths
	lenA := len([]rune(a))
	lenB := len([]rune(b))

	// Empty strings case
	if lenA == 0 && lenB == 0 {
		// Emit telemetry: edge case
		emitEdgeCaseCounter("both_empty")
		return 1.0, nil
	}

	// Handle similarity-based metrics
	switch algorithm {
	case AlgorithmJaroWinkler:
		if opts == nil {
			opts = DefaultScoreOptions()
		}
		return jaroWinklerScore(a, b, opts.JaroPrefixScale, opts.JaroMaxPrefix), nil

	case AlgorithmSubstring:
		_, score := substringMatch(a, b)
		return score, nil
	}

	// Handle distance-based metrics
	maxLen := lenA
	if lenB > maxLen {
		maxLen = lenB
	}

	if maxLen == 0 {
		return 1.0, nil
	}

	distance, err := DistanceWithAlgorithm(a, b, algorithm)
	if err != nil {
		return 0, err
	}

	return 1.0 - float64(distance)/float64(maxLen), nil
}

// levenshteinDistance calculates Levenshtein edit distance.
// Uses optimized native Go implementation (see ADR-0002 for benchmark data).
func levenshteinDistance(a, b string) int {
	// Use existing optimized implementation
	return Distance(a, b)
}

// damerauOSADistance calculates Damerau-Levenshtein distance (OSA variant).
// Uses native Go implementation (see osa.go).
//
// Previous implementation used matchr.OSA() but had a bug with start-of-string
// transpositions (e.g., "hello"/"ehllo" returned 2 instead of 1).
// Native implementation resolves this issue and provides better performance.
// See ADR-0003 for details.
func damerauOSADistance(a, b string) int {
	return osaDistance(a, b)
}

// damerauUnrestrictedDistance calculates unrestricted Damerau-Levenshtein distance.
//
// Wraps matchr library implementation (matchr.DamerauLevenshtein is the unrestricted variant).
//
// Note: Initial implementation attempted to port the Zhao-Sahni algorithm from rapidfuzz-cpp,
// but matchr's implementation proved simpler and already handles the unrestricted variant correctly.
// See ADR-0002 for details on implementation strategy and future string-metrics-go research project.
//
// Reference for algorithm: "Linear space string correction algorithm using the Damerau-Levenshtein distance"
// by Chunchun Zhao and Sartaj Sahni.
func damerauUnrestrictedDistance(a, b string) int {
	return matchr.DamerauLevenshtein(a, b)
}

// damerauUnrestrictedDistanceZhao is the Zhao-Sahni algorithm port from rapidfuzz-cpp.
// Preserved for reference and potential future use if matchr becomes unmaintained.
// See: https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/DamerauLevenshtein_impl.hpp
//
// Copyright notice from rapidfuzz-cpp:
// SPDX-License-Identifier: MIT
// Copyright © 2022-present Max Bachmann
//
// DEPRECATED: Use matchr.DamerauLevenshtein() instead (damerauUnrestrictedDistance wrapper above).
//
//nolint:unused // Preserved for reference, may be used if matchr becomes unmaintained
func damerauUnrestrictedDistanceZhao(a, b string) int {
	// Convert to runes for Unicode support
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

	// maxVal is used as a sentinel for "infinity" in the algorithm
	maxVal := lenA + lenB + 1

	// lastRowID tracks last occurrence of each character in string A
	lastRowID := make(map[rune]int)

	// Three rows for dynamic programming (using 1-based indexing, need extra space):
	// FR: saved values for transposition detection
	// R1: previous row
	// R: current row
	FR := make([]int, lenB+3) // +3 for 1-based indexing and j-2 access
	R1 := make([]int, lenB+3)
	R := make([]int, lenB+3)

	// Initialize with maxVal
	for i := range FR {
		FR[i] = maxVal
		R1[i] = maxVal
	}

	// Initialize first row: distances from empty string (1-based indexing)
	R[0] = maxVal
	for j := 1; j <= lenB+1; j++ {
		R[j] = j - 1
	}

	// Main loop: process each character of string A
	for i := 1; i <= lenA; i++ {
		// Swap rows
		R, R1 = R1, R

		lastColID := -1
		lastI2L1 := R[0]
		R[0] = i
		T := maxVal

		// Process each character of string B
		for j := 1; j <= lenB; j++ {
			charA := runesA[i-1]
			charB := runesB[j-1]

			// Calculate costs for different operations
			cost := 1
			if charA == charB {
				cost = 0
			}

			// Standard operations: insert, delete, substitute
			diag := R1[j-1] + cost
			left := R[j-1] + 1
			up := R1[j] + 1
			temp := min(diag, min(left, up))

			// Handle transpositions
			if charA == charB {
				lastColID = j
				if j >= 2 {
					FR[j] = R1[j-2]
				}
				T = lastI2L1
			} else {
				k, exists := lastRowID[charB]
				if !exists {
					k = -1 // Character not seen in A yet
				}
				l := lastColID

				// Check for transposition opportunities
				if (j-l) == 1 && k >= 0 {
					transpose := FR[j] + (i - k)
					temp = min(temp, transpose)
				} else if (i-k) == 1 && l >= 0 {
					transpose := T + (j - l)
					temp = min(temp, transpose)
				}
			}

			lastI2L1 = R[j]
			R[j] = temp
		}

		lastRowID[runesA[i-1]] = i
	}

	return R[lenB]
}

// jaroWinklerScore calculates Jaro-Winkler similarity score.
// Wraps matchr library implementation.
func jaroWinklerScore(a, b string, prefixScale float64, maxPrefix int) float64 {
	// matchr.JaroWinkler signature: JaroWinkler(r1, r2 string, longTolerance bool)
	// longTolerance: if true, applies additional tolerance for longer strings
	// For now, use false for strict matching (standard Jaro-Winkler behavior)
	//
	// TODO: matchr doesn't expose prefix scale/maxPrefix parameters
	// If custom parameters needed, implement Jaro-Winkler directly or use different library
	_ = prefixScale // Suppress unused warning
	_ = maxPrefix   // Suppress unused warning

	longTolerance := false // Standard Jaro-Winkler behavior
	return matchr.JaroWinkler(a, b, longTolerance)
}

// MatchRange represents a matched substring range.
type MatchRange struct {
	Start int  // Start index (inclusive, 0-based character position)
	End   int  // End index (exclusive, one past last character)
	Valid bool // Whether a match was found
}

// substringMatch finds the longest common substring and calculates similarity score.
//
// Computes the longest common substring (LCS) between needle and haystack,
// returning both the matched range in the haystack and a normalized score.
//
// Examples:
//
//	match, score := substringMatch("hello", "hello world")
//	// match.Start: 0, match.End: 5, score: 0.4545454545454545
//
//	match, score := substringMatch("world", "hello world")
//	// match.Start: 6, match.End: 11, score: 0.4545454545454545
//
//	match, score := substringMatch("xyz", "abcdef")
//	// match.Valid: false, score: 0.0
func substringMatch(needle, haystack string) (MatchRange, float64) {
	runesNeedle := []rune(needle)
	runesHaystack := []rune(haystack)

	lenNeedle := len(runesNeedle)
	lenHaystack := len(runesHaystack)

	if lenNeedle == 0 || lenHaystack == 0 {
		return MatchRange{Valid: false}, 0.0
	}

	maxLen := lenNeedle
	if lenHaystack > maxLen {
		maxLen = lenHaystack
	}

	// Dynamic programming table for LCS
	dp := make([][]int, lenNeedle+1)
	for i := range dp {
		dp[i] = make([]int, lenHaystack+1)
	}

	lcsLength := 0
	lcsEndPos := 0

	// Fill DP table
	for i := 1; i <= lenNeedle; i++ {
		for j := 1; j <= lenHaystack; j++ {
			if runesNeedle[i-1] == runesHaystack[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > lcsLength {
					lcsLength = dp[i][j]
					lcsEndPos = j
				}
			}
		}
	}

	if lcsLength == 0 {
		return MatchRange{Valid: false}, 0.0
	}

	start := lcsEndPos - lcsLength
	end := lcsEndPos
	score := float64(lcsLength) / float64(maxLen)

	return MatchRange{
		Start: start,
		End:   end,
		Valid: true,
	}, score
}

// SubstringMatch finds the longest common substring between needle and haystack.
//
// Returns the matched range in the haystack and a normalized similarity score.
// Score is calculated as: lcs_length / max(len(needle), len(haystack))
//
// Examples:
//
//	match, score := SubstringMatch("hello", "hello world")
//	// match: {Start: 0, End: 5, Valid: true}, score: 0.4545454545454545
//
//	match, score := SubstringMatch("xyz", "abcdef")
//	// match: {Valid: false}, score: 0.0
func SubstringMatch(needle, haystack string) (MatchRange, float64) {
	return substringMatch(needle, haystack)
}
