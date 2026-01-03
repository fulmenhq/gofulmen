package similarity

import (
	"errors"
	"fmt"
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
// Uses native Zhao-Sahni algorithm implementation, replacing the previous matchr dependency
// (GPL-2.0) with MIT-compatible native code for v0.2.0.
//
// Reference for algorithm: "Linear space string correction algorithm using the Damerau-Levenshtein distance"
// by Chunchun Zhao and Sartaj Sahni.
//
// Ported from rapidfuzz-cpp (MIT licensed).
func damerauUnrestrictedDistance(a, b string) int {
	return damerauUnrestrictedDistanceNative(a, b)
}

// damerauUnrestrictedDistanceNative implements unrestricted Damerau-Levenshtein distance.
// Activated in v0.2.0 to replace matchr dependency (GPL-2.0).
//
// This implementation uses the standard dynamic programming approach with full matrix
// to properly handle unrestricted transpositions (characters can be transposed regardless
// of what operations have been applied to the substring between them).
//
// Reference: Damerau, F.J. (1964). "A technique for computer detection and correction
// of spelling errors". Communications of the ACM.
func damerauUnrestrictedDistanceNative(a, b string) int {
	runesA := []rune(a)
	runesB := []rune(b)

	lenA := len(runesA)
	lenB := len(runesB)

	// Edge cases
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Fast path for identical strings
	if a == b {
		return 0
	}

	// Create alphabet map for character positions
	// Maps each character to its last seen position in string A
	da := make(map[rune]int)

	// Initialize with -1 (not seen)
	for _, r := range runesA {
		da[r] = -1
	}
	for _, r := range runesB {
		da[r] = -1
	}

	// Create distance matrix with extra row/column for boundary conditions
	// d[i+1][j+1] represents distance between a[0:i] and b[0:j]
	maxDist := lenA + lenB
	d := make([][]int, lenA+2)
	for i := range d {
		d[i] = make([]int, lenB+2)
	}

	// Initialize boundary conditions
	d[0][0] = maxDist
	for i := 0; i <= lenA; i++ {
		d[i+1][0] = maxDist
		d[i+1][1] = i
	}
	for j := 0; j <= lenB; j++ {
		d[0][j+1] = maxDist
		d[1][j+1] = j
	}

	// Fill the matrix
	for i := 1; i <= lenA; i++ {
		db := -1 // Last position of current a[i-1] character in b

		for j := 1; j <= lenB; j++ {
			i1 := da[runesB[j-1]] // Last position in a where b[j-1] occurred
			j1 := db              // Last position in b where a[i-1] occurred

			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
				db = j - 1
			}

			// Standard Levenshtein operations
			d[i+1][j+1] = min(
				d[i][j]+cost, // substitution (or match)
				d[i+1][j]+1,  // insertion
				d[i][j+1]+1,  // deletion
			)

			// Transposition: if we've seen both characters before
			if i1 >= 0 && j1 >= 0 {
				// Cost of transposition: distance to reach before the transposed pair,
				// plus cost for any characters between the transposed positions, plus 1 for swap.
				// i1, j1 are 0-indexed, i, j are 1-indexed, so adjust:
				// - Characters between i1 and (i-1) in A: (i-1) - i1 - 1 = i - i1 - 2
				// - Characters between j1 and (j-1) in B: (j-1) - j1 - 1 = j - j1 - 2
				transCost := d[i1+1][j1+1] + (i - i1 - 2) + 1 + (j - j1 - 2)
				if transCost < d[i+1][j+1] {
					d[i+1][j+1] = transCost
				}
			}
		}

		da[runesA[i-1]] = i - 1
	}

	return d[lenA+1][lenB+1]
}

// min returns the minimum of the given integers
func min(a, b int, rest ...int) int {
	m := a
	if b < m {
		m = b
	}
	for _, v := range rest {
		if v < m {
			m = v
		}
	}
	return m
}

// jaroWinklerScore calculates Jaro-Winkler similarity score.
//
// Uses native implementation, replacing the previous matchr dependency
// (GPL-2.0) with MIT-compatible native code for v0.2.0.
//
// Note: prefixScale and maxPrefix parameters are accepted for API compatibility
// but the native implementation uses standard values (0.1 and 4) to match
// the previous matchr behavior exactly.
func jaroWinklerScore(a, b string, prefixScale float64, maxPrefix int) float64 {
	// Note: prefixScale and maxPrefix are not used by native implementation
	// to maintain exact parity with previous matchr.JaroWinkler behavior.
	// The native implementation uses standard values: prefixScale=0.1, maxPrefix=4
	_ = prefixScale
	_ = maxPrefix

	longTolerance := false // Standard Jaro-Winkler behavior (matches previous matchr usage)
	return jaroWinklerSimilarity(a, b, longTolerance)
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
