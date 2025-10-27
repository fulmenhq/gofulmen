package similarity

import "github.com/fulmenhq/gofulmen/telemetry"

// telemetrySystem holds the optional telemetry system for similarity operations.
// nil if telemetry is disabled (default).
var telemetrySystem *telemetry.System

// EnableTelemetry enables counter-only telemetry for similarity operations.
//
// Implements ADR-0008 Pattern 1 (counter-only) for performance-sensitive operations.
// Counters track:
//   - Algorithm usage (which algorithms are called)
//   - String length distribution (bucketed)
//   - Fast path hits (identical strings)
//   - Edge cases (empty strings)
//   - API misuse errors (wrong algorithm for API)
//
// Does NOT include:
//   - Histograms (operation duration) - violates ADR-0008 for hot-loop code
//   - Tracing (span creation) - performance-sensitive context
//
// Example usage:
//
//	// In application initialization
//	sys, _ := telemetry.NewSystem(telemetry.DefaultConfig())
//	similarity.EnableTelemetry(sys)
//
//	// Now all similarity operations emit counters
//	distance, _ := similarity.DistanceWithAlgorithm("hello", "world", similarity.AlgorithmLevenshtein)
//	// Emits: foundry.similarity.distance.calls{algorithm=levenshtein}
//	// Emits: foundry.similarity.string_length{bucket=tiny,algorithm=levenshtein}
func EnableTelemetry(sys *telemetry.System) {
	telemetrySystem = sys
}

// DisableTelemetry disables telemetry for similarity operations.
func DisableTelemetry() {
	telemetrySystem = nil
}

// isTelemetryEnabled returns true if telemetry is enabled.
func isTelemetryEnabled() bool {
	return telemetrySystem != nil
}

// emitCounter emits a counter metric if telemetry is enabled.
// Safe to call even if telemetry is disabled (no-op).
func emitCounter(name string, value float64, tags map[string]string) {
	if !isTelemetryEnabled() {
		return
	}

	// Counter emission is best-effort; we don't propagate errors to avoid
	// breaking similarity operations if telemetry has issues.
	_ = telemetrySystem.Counter(name, value, tags)
}

// lengthBucket categorizes string length for performance analysis.
// Buckets chosen to align with benchmark test coverage.
func lengthBucket(s string) string {
	n := len([]rune(s))
	switch {
	case n == 0:
		return "empty"
	case n <= 10:
		return "tiny" // <= 10 chars (fast path, minimal memory)
	case n <= 50:
		return "short" // 11-50 chars (common for CLI args, names)
	case n <= 200:
		return "medium" // 51-200 chars (typical text snippets)
	case n <= 1000:
		return "long" // 201-1000 chars (paragraphs, performance-sensitive)
	default:
		return "very_long" // > 1000 chars (memory optimization critical)
	}
}

// emitAlgorithmCounter emits a counter for algorithm usage.
func emitAlgorithmCounter(api string, algorithm Algorithm) {
	emitCounter("foundry.similarity."+api+".calls", 1, map[string]string{
		"algorithm": string(algorithm),
	})
}

// emitStringLengthCounter emits a counter for string length distribution.
func emitStringLengthCounter(algorithm Algorithm, a, b string) {
	// Use max length to represent the "difficulty" of the operation
	bucketA := lengthBucket(a)
	bucketB := lengthBucket(b)

	// Emit for the longer string (represents worst-case complexity)
	bucket := bucketA
	if len([]rune(b)) > len([]rune(a)) {
		bucket = bucketB
	}

	emitCounter("foundry.similarity.string_length", 1, map[string]string{
		"bucket":    bucket,
		"algorithm": string(algorithm),
	})
}

// emitFastPathCounter emits a counter when identical strings are detected.
func emitFastPathCounter(reason string) {
	emitCounter("foundry.similarity.fast_path", 1, map[string]string{
		"reason": reason,
	})
}

// emitEdgeCaseCounter emits a counter for edge cases.
func emitEdgeCaseCounter(caseType string) {
	emitCounter("foundry.similarity.edge_case", 1, map[string]string{
		"case": caseType,
	})
}

// emitErrorCounter emits a counter for API misuse errors.
func emitErrorCounter(errorType string, algorithm Algorithm, correctAPI string) {
	emitCounter("foundry.similarity.error", 1, map[string]string{
		"type":        errorType,
		"algorithm":   string(algorithm),
		"correct_api": correctAPI,
	})
}
