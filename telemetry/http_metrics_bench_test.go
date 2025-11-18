package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/stretchr/testify/require"
)

// mockEmitter for performance testing (minimal overhead)
type performanceMockEmitter struct {
	// Use simple counters to avoid map allocation overhead
	counterCalls   int64
	histogramCalls int64
	gaugeCalls     int64
}

func (m *performanceMockEmitter) Counter(name string, value float64, tags map[string]string) error {
	m.counterCalls++
	return nil
}

func (m *performanceMockEmitter) Histogram(name string, value time.Duration, tags map[string]string) error {
	m.histogramCalls++
	return nil
}

func (m *performanceMockEmitter) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
	m.histogramCalls++
	return nil
}

func (m *performanceMockEmitter) Gauge(name string, value float64, tags map[string]string) error {
	m.gaugeCalls++
	return nil
}

// BenchmarkHTTPMetricsMiddleware measures the overhead of the HTTP metrics middleware
func BenchmarkHTTPMetricsMiddleware(b *testing.B) {
	emitter := &performanceMockEmitter{}

	// Create middleware
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("benchmark-api"),
	)

	// Create a simple handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	// Reset timer to exclude setup costs
	b.ResetTimer()

	// Benchmark the middleware
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Verify metrics were collected
	require.Greater(b, emitter.counterCalls, int64(0))
	require.Greater(b, emitter.histogramCalls, int64(0))
	require.Greater(b, emitter.gaugeCalls, int64(0))
}

// BenchmarkHTTPMetricsMiddlewareWithoutMetrics measures baseline performance without middleware
func BenchmarkHTTPMetricsMiddlewareWithoutMetrics(b *testing.B) {
	// Create a simple handler without middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	})

	// Reset timer to exclude setup costs
	b.ResetTimer()

	// Benchmark the handler without middleware
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkHTTPMetricsMiddlewareWithRouteNormalization tests performance with route normalization
func BenchmarkHTTPMetricsMiddlewareWithRouteNormalization(b *testing.B) {
	emitter := &performanceMockEmitter{}

	// Create middleware with route normalization
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("benchmark-api"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			// Simulate more complex route normalization
			if len(path) > 10 {
				return "/long/{path}"
			}
			return "/short"
		}),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/very/long/path/for/normalization", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkHTTPMetricsMiddlewareWithCustomSizeBuckets tests performance with custom size buckets
func BenchmarkHTTPMetricsMiddlewareWithCustomSizeBuckets(b *testing.B) {
	emitter := &performanceMockEmitter{}

	// Create middleware with custom buckets
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("benchmark-api"),
		telemetry.WithCustomSizeBuckets(
			[]float64{512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}, // More size buckets
		),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkHTTPMetricsMiddlewareConcurrent tests concurrent performance
func BenchmarkHTTPMetricsMiddlewareConcurrent(b *testing.B) {
	emitter := &performanceMockEmitter{}

	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("benchmark-api"),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	b.ResetTimer()

	// Run concurrent benchmarks
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

// BenchmarkRouteNormalization tests just the route normalization performance
func BenchmarkRouteNormalization(b *testing.B) {
	normalizer := telemetry.DefaultRouteNormalizer

	testPaths := []string{
		"/api/users/123",
		"/api/users/550e8400-e29b-41d4-a716-446655440000",
		"/api/v1/users/123/profile/456",
		"/static/css/main.css",
		"/",
		"/api/search?q=test&sort=desc",
		"/very/long/path/with/many/segments/123/456/789",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		_ = normalizer("GET", path)
	}
}

// BenchmarkRouteNormalizationCustom tests custom route normalization performance
func BenchmarkRouteNormalizationCustom(b *testing.B) {
	customNormalizer := func(method, path string) string {
		// More complex normalization logic
		if idx := strings.IndexAny(path, "?#"); idx >= 0 {
			path = path[:idx]
		}

		// UUID detection
		if strings.Contains(path, "-") && len(path) > 30 {
			return strings.Replace(path,
				"550e8400-e29b-41d4-a716-446655440000",
				"{uuid}", 1)
		}

		// Numeric segments
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if _, err := strconv.Atoi(part); err == nil {
				parts[i] = "{id}"
			}
		}

		return strings.Join(parts, "/")
	}

	testPaths := []string{
		"/api/users/123",
		"/api/users/550e8400-e29b-41d4-a716-446655440000",
		"/api/v1/users/123/profile/456",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		_ = customNormalizer("GET", path)
	}
}

// TestHTTPMetricsMiddlewareOverhead calculates the actual overhead percentage using proper benchmarking
func TestHTTPMetricsMiddlewareOverhead(t *testing.T) {
	emitter := &performanceMockEmitter{}

	// Test with middleware
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("overhead-test"),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Fatal(err)
		}
	}))

	// Test without middleware
	baselineHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Fatal(err)
		}
	})

	// Run benchmarks to get per-operation timings
	baselineResult := testing.Benchmark(func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			baselineHandler.ServeHTTP(w, req)
		}
	})

	middlewareResult := testing.Benchmark(func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})

	// Calculate overhead using per-operation timings (proper methodology)
	baselineNs := baselineResult.NsPerOp()
	middlewareNs := middlewareResult.NsPerOp()
	overheadNs := middlewareNs - baselineNs
	overheadPercent := float64(overheadNs) / float64(baselineNs) * 100

	// Also measure wall-clock time for 10k requests (informational only)
	start := time.Now()
	for i := 0; i < 10000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
	wallClockWithMiddleware := time.Since(start)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		baselineHandler.ServeHTTP(w, req)
	}
	wallClockBaseline := time.Since(start)

	wallClockOverhead := wallClockWithMiddleware - wallClockBaseline
	wallClockOverheadPercent := float64(wallClockOverhead) / float64(wallClockBaseline) * 100

	// Log both measurements
	t.Logf("=== PER-OPERATION OVERHEAD (Benchmark Method - Accurate) ===")
	t.Logf("Baseline: %d ns/op", baselineNs)
	t.Logf("With middleware: %d ns/op", middlewareNs)
	t.Logf("Overhead: %d ns/op (%.2f%%)", overheadNs, overheadPercent)

	t.Logf("\n=== WALL-CLOCK OVERHEAD (10k requests - Informational) ===")
	t.Logf("Baseline: %v", wallClockBaseline)
	t.Logf("With middleware: %v", wallClockWithMiddleware)
	t.Logf("Overhead: %v (%.2f%%)", wallClockOverhead, wallClockOverheadPercent)
	t.Logf("Note: Wall-clock includes test harness overhead and runtime variance")

	// Verify metrics were collected
	require.Greater(t, emitter.counterCalls, int64(0))
	require.Greater(t, emitter.histogramCalls, int64(0))
	require.Greater(t, emitter.gaugeCalls, int64(0))

	// Assert overhead is reasonable using per-operation measurement (proper methodology)
	// Comprehensive HTTP metrics (5 metrics per request) should have <50% per-operation overhead
	require.Less(t, overheadPercent, 50.0, "Middleware per-operation overhead should be less than 50%%")

	// Wall-clock overhead will be higher due to test harness, but should still be reasonable
	require.Less(t, wallClockOverheadPercent, 150.0, "Middleware wall-clock overhead should be less than 150%%")
}
