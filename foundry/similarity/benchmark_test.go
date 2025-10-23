package similarity

import (
	"strings"
	"testing"
)

// BenchmarkDistance_Short benchmarks distance calculation for short strings (4-8 chars)
func BenchmarkDistance_Short(b *testing.B) {
	a := "kitten"
	b2 := "sitting"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Distance(a, b2)
	}
}

// BenchmarkDistance_Medium benchmarks distance calculation for medium strings (~30 chars)
func BenchmarkDistance_Medium(b *testing.B) {
	a := "The quick brown fox jumps over"
	b2 := "The quick brawn fox jumped oven"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Distance(a, b2)
	}
}

// BenchmarkDistance_128Chars benchmarks distance calculation for 128-character strings
// Target: ≤0.5ms p95 per Crucible standard
func BenchmarkDistance_128Chars(b *testing.B) {
	// Create two similar 128-character strings
	a := strings.Repeat("abcdefgh", 16)  // "abcdefgh" x 16 = 128 chars
	b2 := strings.Repeat("abcdxfgh", 16) // Similar with one char different per block

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Distance(a, b2)
	}
}

// BenchmarkDistance_LongDifferent benchmarks worst-case: long completely different strings
func BenchmarkDistance_LongDifferent(b *testing.B) {
	a := strings.Repeat("a", 128)
	b2 := strings.Repeat("b", 128)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Distance(a, b2)
	}
}

// BenchmarkDistance_Unicode benchmarks distance with Unicode characters
func BenchmarkDistance_Unicode(b *testing.B) {
	a := "Café München 🎉"
	b2 := "Cafe Munchen 🎊"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Distance(a, b2)
	}
}

// BenchmarkScore_Short benchmarks normalized score for short strings
func BenchmarkScore_Short(b *testing.B) {
	a := "test"
	b2 := "best"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Score(a, b2)
	}
}

// BenchmarkScore_Medium benchmarks normalized score for medium strings
func BenchmarkScore_Medium(b *testing.B) {
	a := "The quick brown fox jumps over"
	b2 := "The quick brawn fox jumped oven"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Score(a, b2)
	}
}

// BenchmarkScore_128Chars benchmarks normalized score for 128-character strings
// Target: ≤0.5ms p95 per Crucible standard
func BenchmarkScore_128Chars(b *testing.B) {
	a := strings.Repeat("abcdefgh", 16)
	b2 := strings.Repeat("abcdxfgh", 16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Score(a, b2)
	}
}

// BenchmarkNormalize_Simple benchmarks basic normalization (trim + casefold)
func BenchmarkNormalize_Simple(b *testing.B) {
	input := "  Hello World  "
	opts := NormalizeOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Normalize(input, opts)
	}
}

// BenchmarkNormalize_WithAccentStrip benchmarks normalization with accent stripping
func BenchmarkNormalize_WithAccentStrip(b *testing.B) {
	input := "  Café München Résumé  "
	opts := NormalizeOptions{StripAccents: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Normalize(input, opts)
	}
}

// BenchmarkNormalize_Turkish benchmarks Turkish locale normalization
func BenchmarkNormalize_Turkish(b *testing.B) {
	input := "İSTANBUL"
	opts := NormalizeOptions{Locale: "tr"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Normalize(input, opts)
	}
}

// BenchmarkSuggest_SmallCandidateList benchmarks suggestion with 5 candidates
func BenchmarkSuggest_SmallCandidateList(b *testing.B) {
	input := "confg"
	candidates := []string{"config", "configure", "conform", "confirm", "confluence"}
	opts := SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 3,
		Normalize:      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Suggest(input, candidates, opts)
	}
}

// BenchmarkSuggest_MediumCandidateList benchmarks suggestion with 20 candidates
func BenchmarkSuggest_MediumCandidateList(b *testing.B) {
	input := "schem"
	candidates := []string{
		"schema", "scheme", "schedule", "school", "science",
		"scream", "screen", "script", "scroll", "search",
		"season", "second", "secret", "section", "secure",
		"select", "senior", "sensor", "server", "service",
	}
	opts := SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 3,
		Normalize:      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Suggest(input, candidates, opts)
	}
}

// BenchmarkSuggest_LargeCandidateList benchmarks suggestion with 100 candidates
func BenchmarkSuggest_LargeCandidateList(b *testing.B) {
	input := "test"
	// Generate 100 candidates
	candidates := make([]string, 100)
	for i := 0; i < 100; i++ {
		candidates[i] = strings.Repeat("x", i%20+1) + "test" + strings.Repeat("y", (i+5)%15)
	}

	opts := SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 3,
		Normalize:      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Suggest(input, candidates, opts)
	}
}

// BenchmarkSuggest_NoNormalization benchmarks suggestion without normalization
func BenchmarkSuggest_NoNormalization(b *testing.B) {
	input := "test"
	candidates := []string{"test1", "test2", "test3", "best", "rest", "nest"}
	opts := SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 3,
		Normalize:      false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Suggest(input, candidates, opts)
	}
}
