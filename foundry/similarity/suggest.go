package similarity

// ═══════════════════════════════════════════════════════════════════════════════
// TODO(v0.1.4+): ADD DAMERAU-LEVENSHTEIN & JARO-WINKLER METRICS
// ═══════════════════════════════════════════════════════════════════════════════
// Planned expansion (next iteration):
// 1. Add DistanceMetric enum (Levenshtein, DamerauLevenshtein, JaroWinkler)
// 2. Add Metric field to SuggestOptions (default: Levenshtein for backward compat)
// 3. Implement DamerauDistance() function
// 4. Implement JaroWinklerSimilarity() function
// 5. Update Suggest() to switch on metric type
// 6. Enable skipped tests in fixtures_test.go
//
// See: .plans/active/v0.1.3/012-similarity-expansion-roadmap.md
// ═══════════════════════════════════════════════════════════════════════════════

// scoredCandidate is an internal type used during suggestion ranking.
type scoredCandidate struct {
	originalValue   string
	normalizedValue string
	score           float64
}

// Suggestion represents a ranked fuzzy match result.
//
// Suggestions are returned by the Suggest function and include both the
// original candidate value and its similarity score.
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
type Suggestion struct {
	// Value is the candidate string that matched.
	// This is the original value from the candidates list.
	Value string

	// Score is the similarity score in the range [0.0, 1.0].
	// Higher scores indicate closer matches:
	//   - 1.0 = identical match
	//   - 0.6 = default threshold for "similar enough"
	//   - 0.0 = completely different
	Score float64
}

// SuggestOptions configures the suggestion ranking behavior.
//
// Default values follow empirical best practices:
//   - MinScore: 0.6 (balances helpful suggestions vs. noise)
//   - MaxSuggestions: 3 (avoids overwhelming users)
//   - Normalize: true (most use cases benefit from case-insensitive matching)
//
// Note: The Normalize field defaults to true per the Crucible standard,
// but Go's zero value for bool is false. Callers should explicitly set
// Normalize or use DefaultSuggestOptions() helper.
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
type SuggestOptions struct {
	// MinScore is the minimum similarity score threshold.
	// Candidates with score < MinScore are filtered out.
	// Default: 0.6
	//
	// Range: [0.0, 1.0]
	MinScore float64

	// MaxSuggestions is the maximum number of suggestions to return.
	// Results are sorted by score (descending), then alphabetically.
	// Default: 3
	//
	// Must be: >= 1
	MaxSuggestions int

	// Normalize controls whether to normalize input and candidates before scoring.
	// When true, applies case folding for case-insensitive matching.
	// Default: true (per Crucible standard)
	//
	// Note: Use pointer to distinguish unset from explicit false, or
	// use DefaultSuggestOptions() to get correct defaults.
	Normalize bool
}

// DefaultSuggestOptions returns SuggestOptions with Crucible standard defaults.
//
// Returns:
//   - MinScore: 0.6
//   - MaxSuggestions: 3
//   - Normalize: true
//
// Usage:
//
//	opts := similarity.DefaultSuggestOptions()
//	suggestions := similarity.Suggest(input, candidates, opts)
func DefaultSuggestOptions() SuggestOptions {
	return SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 3,
		Normalize:      true,
	}
}

// Suggest generates ranked fuzzy match suggestions from a candidate list.
//
// The algorithm performs these steps:
//  1. Normalize input and candidates (if opts.Normalize is true)
//  2. Calculate similarity score for each candidate
//  3. Filter candidates with score >= opts.MinScore
//  4. Sort by score (descending), then alphabetically for ties
//  5. Return top opts.MaxSuggestions results
//
// Parameters:
//   - input: The string to find matches for (e.g., user typo)
//   - candidates: List of valid values to match against
//   - opts: Configuration for scoring and filtering
//
// Returns:
//   - Slice of Suggestion sorted by score (highest first)
//   - Empty slice if no candidates meet the threshold
//
// Examples:
//
//	candidates := []string{"docscribe", "crucible", "foundry"}
//	suggestions := Suggest("docscrib", candidates, DefaultSuggestOptions())
//	// Returns: [{Value: "docscribe", Score: 0.8889}]
//
//	suggestions := Suggest("xyz", []string{"abc", "def"}, DefaultSuggestOptions())
//	// Returns: [] (no candidates meet 0.6 threshold)
//
// Use Cases:
//   - "Did you mean...?" error messages
//   - CLI command suggestions for typos
//   - Document/asset discovery with fuzzy search
//   - Configuration key suggestions
//
// Conformance: Implements Crucible Foundry Similarity Standard v1.0.0 (2025.10.2).
func Suggest(input string, candidates []string, opts SuggestOptions) []Suggestion {
	// Apply defaults if not set
	minScore := opts.MinScore
	if minScore == 0 {
		minScore = 0.6 // Crucible default
	}
	maxSuggestions := opts.MaxSuggestions
	if maxSuggestions == 0 {
		maxSuggestions = 3 // Crucible default
	}
	// Note: opts.Normalize defaults to false (Go zero value)
	// Callers should use DefaultSuggestOptions() or set explicitly

	// Handle empty input or candidates
	if len(candidates) == 0 {
		return []Suggestion{}
	}

	// Prepare normalized versions if requested
	normalizedInput := input
	normalizedCandidates := make([]string, len(candidates))
	copy(normalizedCandidates, candidates)

	if opts.Normalize {
		// Normalize input
		normalizedInput = Normalize(input, NormalizeOptions{})

		// Normalize all candidates
		for i, candidate := range candidates {
			normalizedCandidates[i] = Normalize(candidate, NormalizeOptions{})
		}
	}

	// Score all candidates
	scored := make([]scoredCandidate, 0, len(candidates))
	for i, candidate := range candidates {
		score := Score(normalizedInput, normalizedCandidates[i])

		// Filter by minimum score
		if score >= minScore {
			scored = append(scored, scoredCandidate{
				originalValue:   candidate,
				normalizedValue: normalizedCandidates[i],
				score:           score,
			})
		}
	}

	// If no candidates meet threshold, return empty
	if len(scored) == 0 {
		return []Suggestion{}
	}

	// Sort by score (descending), then alphabetically for ties
	// Using insertion sort for small slices (typically < 10 candidates)
	for i := 1; i < len(scored); i++ {
		key := scored[i]
		j := i - 1

		// Move elements that are "less than" key to the right
		for j >= 0 && shouldSwap(scored[j], key) {
			scored[j+1] = scored[j]
			j--
		}
		scored[j+1] = key
	}

	// Return top maxSuggestions
	limit := maxSuggestions
	if limit > len(scored) {
		limit = len(scored)
	}

	results := make([]Suggestion, limit)
	for i := 0; i < limit; i++ {
		results[i] = Suggestion{
			Value: scored[i].originalValue,
			Score: scored[i].score,
		}
	}

	return results
}

// shouldSwap returns true if a should come after b in the sorted order.
// Sort order: score descending, then alphabetically ascending for ties.
func shouldSwap(a, b scoredCandidate) bool {
	// If scores are different, higher score comes first
	if a.score != b.score {
		return a.score < b.score // a has lower score, so b should come first
	}

	// Scores are equal, sort alphabetically
	return a.originalValue > b.originalValue // a is later alphabetically, so b should come first
}
