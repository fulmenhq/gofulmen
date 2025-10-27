/*
Package similarity provides text similarity scoring and normalization utilities
following the Fulmen Helper Library Standard (2025.10.2).

# Overview

The similarity package implements standardized text comparison capabilities
for fuzzy matching, "Did you mean...?" suggestions, and Unicode-aware text
processing. It provides cross-language API parity with pyfulmen and tsfulmen
helper libraries.

# Text Similarity

The package implements Levenshtein distance calculation using the Wagner-Fischer
dynamic programming algorithm with Unicode-aware character counting.

Distance returns the edit distance between two strings:

	dist := similarity.Distance("kitten", "sitting") // Returns: 3

Score returns a normalized similarity score from 0.0 (different) to 1.0 (identical):

	score := similarity.Score("kitten", "sitting") // Returns: 0.5714...

# Normalization

Unicode-aware text normalization with optional accent stripping:

	opts := similarity.NormalizeOptions{StripAccents: true}
	normalized := similarity.Normalize("  Café  ", opts) // Returns: "cafe"

Specialized normalization functions:

	folded := similarity.Casefold("Hello", "")                     // Returns: "hello"
	stripped := similarity.StripAccents("naïve")                   // Returns: "naive"
	equal := similarity.EqualsIgnoreCase("Hello", "HELLO", opts)   // Returns: true

Turkish locale support for special case folding rules:

	folded := similarity.Casefold("İstanbul", "tr") // Returns: "istanbul"

# Suggestions

Generate ranked suggestions for fuzzy matching:

	opts := similarity.DefaultSuggestOptions()
	suggestions := similarity.Suggest("docscrib", candidates, opts)

	for _, s := range suggestions {
		fmt.Printf("%s (%.1f%% match)\n", s.Value, s.Score*100)
	}
	// Output: docscribe (88.9% match)

Configure suggestion behavior:

	opts := similarity.SuggestOptions{
		MinScore:       0.6,   // Minimum similarity threshold
		MaxSuggestions: 3,     // Maximum results to return
		Normalize:      true,  // Case-insensitive matching
	}

# Performance

Distance and Score operations target ≤0.5ms p95 latency for 128-character strings
per Crucible standard. Actual performance significantly exceeds this target:

Benchmark results (M1 Pro, 12 cores):
  - Distance (128 chars): ~28 μs/op (17x faster than target)
  - Score (128 chars):    ~28 μs/op (17x faster than target)
  - Distance (short):     ~125 ns/op
  - Normalize (simple):   ~40 ns/op
  - Suggest (20 items):   ~3 μs/op

Run benchmarks with: go test -bench=. ./foundry/similarity/

Memory allocations are minimal:
  - Short strings: 96-128 B/op, 2 allocs
  - 128-char strings: ~3.3 KB/op, 4 allocs
  - Suggest operations: scales linearly with candidate count

# Cross-Language Compatibility

This implementation conforms to the Crucible Foundry Similarity Standard v1.0.0
and maintains API parity with pyfulmen and tsfulmen helper libraries.

Shared test fixtures validate cross-language consistency:
  - docs/crucible-go/standards/library/foundry/similarity.md
  - config/crucible-go/library/foundry/similarity-fixtures.yaml
  - schemas/crucible-go/library/foundry/v1.0.0/similarity.schema.json

# Use Cases

Common integration patterns:

		// CLI command suggestions
		func suggestCommand(input string, validCommands []string) {
			suggestions := similarity.Suggest(input, validCommands,
				similarity.DefaultSuggestOptions())
			if len(suggestions) > 0 {
				fmt.Printf("Unknown command '%s'. Did you mean:\n", input)
				for _, s := range suggestions {
					fmt.Printf("  - %s\n", s.Value)
				}
			}
		}

		// Crucible asset discovery
		func findAsset(query string, assets []Asset) *Asset {
			candidates := make([]string, len(assets))
			for i, a := range assets {
				candidates[i] = a.ID
			}

			suggestions := similarity.Suggest(query, candidates,
				similarity.SuggestOptions{
					MinScore:       0.5,
					MaxSuggestions: 1,
					Normalize:      true,
				})

		if len(suggestions) > 0 {
			return findAssetByID(suggestions[0].Value)
		}
		return nil
	}

# Unified API (v2.0.0+)

The v2 API provides algorithm-specific distance and score calculations
following the Crucible Foundry Similarity Standard v2.0.0:

	// Distance-based algorithms
	dist, _ := similarity.DistanceWithAlgorithm("hello", "world",
		similarity.AlgorithmLevenshtein)

	dist, _ := similarity.DistanceWithAlgorithm("hello", "ehllo",
		similarity.AlgorithmDamerauOSA) // Optimal String Alignment

	dist, _ := similarity.DistanceWithAlgorithm("CA", "ABC",
		similarity.AlgorithmDamerauUnrestricted) // True Damerau-Levenshtein

	// Score-based algorithms (similarity from 0.0 to 1.0)
	score, _ := similarity.ScoreWithAlgorithm("martha", "marhta",
		similarity.AlgorithmJaroWinkler, nil)

	score, _ := similarity.ScoreWithAlgorithm("hello", "hello world",
		similarity.AlgorithmSubstring, nil)

Supported algorithms:
  - AlgorithmLevenshtein: Classic edit distance (insertions, deletions, substitutions)
  - AlgorithmDamerauOSA: Optimal String Alignment (adds adjacent transpositions)
  - AlgorithmDamerauUnrestricted: True Damerau-Levenshtein (unrestricted transpositions)
  - AlgorithmJaroWinkler: Similarity metric optimized for short strings with common prefixes
  - AlgorithmSubstring: Longest common substring matching

See ADR-0002 and ADR-0003 for algorithm implementation details and performance benchmarks.

# Telemetry (Optional)

The package supports opt-in counter-only telemetry following ADR-0008 Pattern 1
(performance-sensitive, hot-loop eligible). Telemetry provides production visibility
into algorithm usage, string length distribution, and API misuse without significant
performance impact.

Enable telemetry during application initialization:

	sys, _ := telemetry.NewSystem(telemetry.DefaultConfig())
	similarity.EnableTelemetry(sys)

	// Now all similarity operations emit counters:
	_, _ = similarity.DistanceWithAlgorithm("hello", "world",
		similarity.AlgorithmLevenshtein)
	// Emits: foundry.similarity.distance.calls{algorithm=levenshtein}
	// Emits: foundry.similarity.string_length{bucket=tiny,algorithm=levenshtein}

Telemetry is disabled by default (zero overhead). When enabled, overhead is ~1μs per
operation (acceptable for typical use cases like CLI suggestions and spell checking).

Metrics emitted:
  - foundry.similarity.distance.calls: Counter of DistanceWithAlgorithm calls by algorithm
  - foundry.similarity.score.calls: Counter of ScoreWithAlgorithm calls by algorithm
  - foundry.similarity.string_length: Counter of operations by string length bucket
  - foundry.similarity.fast_path: Counter of fast path hits (identical strings)
  - foundry.similarity.edge_case: Counter of edge cases (empty strings)
  - foundry.similarity.error: Counter of API misuse errors

For applications with ultra-low latency requirements, keep telemetry disabled (default).
See phase3-telemetry-backlog.md for instrumentation details and overhead analysis.

# Algorithm Details

Levenshtein Distance:
  - Wagner-Fischer dynamic programming algorithm
  - Two-row space optimization: O(min(m,n)) space
  - Early-exit optimization for large length differences
  - Unicode-aware using rune slices for grapheme counting

Normalization Pipeline:
 1. Trim leading/trailing whitespace
 2. Apply Unicode case folding (simple or locale-specific)
 3. Optionally strip accents via NFD normalization

Accent Stripping:
 1. Decompose to NFD (Unicode Normalization Form Decomposed)
 2. Filter out combining marks (Unicode category Mn)
 3. Recompose to NFC (Unicode Normalization Form Composed)

# Conformance

Standard: Crucible Foundry Similarity Standard v2.0.0 (2025.10.3)
  - v1 API (Distance, Score): Standard v1.0.0 (2025.10.2) - Stable
  - v2 API (DistanceWithAlgorithm, ScoreWithAlgorithm): Standard v2.0.0 (2025.10.3) - Stable

Module: gofulmen/foundry
Version: 0.1.5+
Status: Stable

# References

  - Levenshtein Distance: https://en.wikipedia.org/wiki/Levenshtein_distance
  - Wagner-Fischer Algorithm: https://en.wikipedia.org/wiki/Wagner–Fischer_algorithm
  - Unicode Normalization: https://unicode.org/reports/tr15/
  - Unicode Case Folding: https://www.unicode.org/reports/tr21/tr21-5.html
*/
package similarity
