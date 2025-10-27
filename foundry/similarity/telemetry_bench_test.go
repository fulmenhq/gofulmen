package similarity

import (
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry"
	teltest "github.com/fulmenhq/gofulmen/telemetry/testing"
)

// BenchmarkDistanceWithAlgorithm_NoTelemetry benchmarks without telemetry (baseline)
func BenchmarkDistanceWithAlgorithm_NoTelemetry(b *testing.B) {
	// Ensure telemetry is disabled
	DisableTelemetry()

	tests := []struct {
		name string
		a    string
		b    string
		algo Algorithm
	}{
		{"levenshtein_tiny", "hello", "world", AlgorithmLevenshtein},
		{"levenshtein_short", "the quick brown fox", "the slow brown dog", AlgorithmLevenshtein},
		{"osa_tiny", "hello", "ehllo", AlgorithmDamerauOSA},
		{"osa_short", "algorithm", "lagorithm", AlgorithmDamerauOSA},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = DistanceWithAlgorithm(tt.a, tt.b, tt.algo)
			}
		})
	}
}

// BenchmarkDistanceWithAlgorithm_WithTelemetry benchmarks with telemetry enabled
func BenchmarkDistanceWithAlgorithm_WithTelemetry(b *testing.B) {
	// Setup fake collector (minimal overhead)
	collector := teltest.NewFakeCollector()
	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		b.Fatalf("failed to create telemetry system: %v", err)
	}

	EnableTelemetry(sys)
	defer DisableTelemetry()

	tests := []struct {
		name string
		a    string
		b    string
		algo Algorithm
	}{
		{"levenshtein_tiny", "hello", "world", AlgorithmLevenshtein},
		{"levenshtein_short", "the quick brown fox", "the slow brown dog", AlgorithmLevenshtein},
		{"osa_tiny", "hello", "ehllo", AlgorithmDamerauOSA},
		{"osa_short", "algorithm", "lagorithm", AlgorithmDamerauOSA},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			// Reset collector before each sub-benchmark
			collector.Reset()
			for i := 0; i < b.N; i++ {
				_, _ = DistanceWithAlgorithm(tt.a, tt.b, tt.algo)
			}
		})
	}
}

// BenchmarkScoreWithAlgorithm_NoTelemetry benchmarks score without telemetry (baseline)
func BenchmarkScoreWithAlgorithm_NoTelemetry(b *testing.B) {
	// Ensure telemetry is disabled
	DisableTelemetry()

	tests := []struct {
		name string
		a    string
		b    string
		algo Algorithm
	}{
		{"levenshtein_tiny", "hello", "world", AlgorithmLevenshtein},
		{"levenshtein_identical", "hello", "hello", AlgorithmLevenshtein}, // Fast path
		{"jaro_winkler", "martha", "marhta", AlgorithmJaroWinkler},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ScoreWithAlgorithm(tt.a, tt.b, tt.algo, nil)
			}
		})
	}
}

// BenchmarkScoreWithAlgorithm_WithTelemetry benchmarks score with telemetry enabled
func BenchmarkScoreWithAlgorithm_WithTelemetry(b *testing.B) {
	// Setup fake collector (minimal overhead)
	collector := teltest.NewFakeCollector()
	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		b.Fatalf("failed to create telemetry system: %v", err)
	}

	EnableTelemetry(sys)
	defer DisableTelemetry()

	tests := []struct {
		name string
		a    string
		b    string
		algo Algorithm
	}{
		{"levenshtein_tiny", "hello", "world", AlgorithmLevenshtein},
		{"levenshtein_identical", "hello", "hello", AlgorithmLevenshtein}, // Fast path
		{"jaro_winkler", "martha", "marhta", AlgorithmJaroWinkler},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			// Reset collector before each sub-benchmark
			collector.Reset()
			for i := 0; i < b.N; i++ {
				_, _ = ScoreWithAlgorithm(tt.a, tt.b, tt.algo, nil)
			}
		})
	}
}

// BenchmarkTelemetryOverhead_FastPath measures overhead for fast path (identical strings)
func BenchmarkTelemetryOverhead_FastPath(b *testing.B) {
	b.Run("NoTelemetry", func(b *testing.B) {
		DisableTelemetry()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = ScoreWithAlgorithm("test", "test", AlgorithmLevenshtein, nil)
		}
	})

	b.Run("WithTelemetry", func(b *testing.B) {
		collector := teltest.NewFakeCollector()
		sys, _ := telemetry.NewSystem(&telemetry.Config{
			Enabled: true,
			Emitter: collector,
		})
		EnableTelemetry(sys)
		defer DisableTelemetry()

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = ScoreWithAlgorithm("test", "test", AlgorithmLevenshtein, nil)
		}
	})
}
