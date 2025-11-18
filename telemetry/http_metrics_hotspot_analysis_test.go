package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// BenchmarkBaseline measures pure handler performance without any middleware
func BenchmarkBaseline(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareOnly measures just the middleware wrapper overhead
func BenchmarkMiddlewareOnly(b *testing.B) {
	emitter := &performanceMockEmitter{}

	// Create middleware but with minimal processing
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("test"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			return path // No normalization
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
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkRouteNormalizationOnly measures just the route normalization overhead
func BenchmarkRouteNormalizationOnly(b *testing.B) {
	normalizer := telemetry.DefaultRouteNormalizer

	testPaths := []string{
		"/api/users/123",
		"/api/users/550e8400-e29b-41d4-a716-446655440000",
		"/api/v1/users/123/profile/456",
		"/static/css/main.css",
		"/",
		"/api/search?q=test&sort=desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		_ = normalizer("GET", path)
	}
}

// BenchmarkTagCreation measures tag map creation overhead
func BenchmarkTagCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tags := map[string]string{
			metrics.TagMethod:  "GET",
			metrics.TagRoute:   "/api/users/{id}",
			metrics.TagStatus:  "200",
			metrics.TagService: "test-service",
		}
		_ = tags
	}
}

// BenchmarkEmitterCalls measures emitter call overhead
func BenchmarkEmitterCalls(b *testing.B) {
	emitter := &performanceMockEmitter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the 5 emitter calls made per request
		_ = emitter.Counter("http_requests_total", 1, nil)
		_ = emitter.Histogram("http_request_duration_seconds", time.Nanosecond, nil)
		_ = emitter.HistogramSummary("http_request_size_bytes", telemetry.HistogramSummary{}, nil)
		_ = emitter.HistogramSummary("http_response_size_bytes", telemetry.HistogramSummary{}, nil)
		_ = emitter.Gauge("http_active_requests", 1, nil)
	}
}

// BenchmarkHistogramConstruction measures histogram summary construction overhead
func BenchmarkHistogramConstruction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate histogram summary construction
		buckets := make([]telemetry.HistogramBucket, 7)
		for j := 0; j < 6; j++ {
			buckets[j] = telemetry.HistogramBucket{
				LE:    float64(j) * 1000,
				Count: int64(j + 1),
			}
		}
		buckets[6] = telemetry.HistogramBucket{
			LE:    float64(999999),
			Count: int64(6),
		}
		summary := telemetry.HistogramSummary{
			Count:   1,
			Sum:     1024.0,
			Buckets: buckets,
		}
		_ = summary
	}
}

// BenchmarkActiveRequestsGauge measures just the active requests gauge overhead
func BenchmarkActiveRequestsGauge(b *testing.B) {
	emitter := &performanceMockEmitter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = emitter.Gauge("http_active_requests", float64(i%10), map[string]string{
			metrics.TagService: "test-service",
		})
	}
}

// BenchmarkRequestProcessing measures the full request processing pipeline
func BenchmarkRequestProcessing(b *testing.B) {
	emitter := &performanceMockEmitter{}

	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("test-service"),
		telemetry.WithRouteNormalizer(telemetry.DefaultRouteNormalizer),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/123", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkRequestProcessingNoNormalization measures request processing without route normalization
func BenchmarkRequestProcessingNoNormalization(b *testing.B) {
	emitter := &performanceMockEmitter{}

	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("test-service"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			return path // Skip normalization
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
		req := httptest.NewRequest("GET", "/api/users/123", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkRequestProcessingNoMetrics measures request processing without metrics emission
func BenchmarkRequestProcessingNoMetrics(b *testing.B) {
	// Create middleware that does everything except emit metrics
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate middleware work (tag creation, route normalization)
			tags := map[string]string{
				metrics.TagMethod:  r.Method,
				metrics.TagRoute:   "/api/users/{id}",
				metrics.TagStatus:  "200",
				metrics.TagService: "test-service",
			}
			_ = tags

			// Simulate active requests tracking
			// (In real implementation, this would be atomic)

			next.ServeHTTP(w, r)
		})
	}

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.Fatal(err)
		}
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/123", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// Performance analysis helper
func TestOverheadAnalysis(t *testing.T) {
	t.Log("=== OVERHEAD ANALYSIS ===")

	// Run targeted benchmarks to identify hotspots
	baselineResult := testing.Benchmark(BenchmarkBaseline)
	middlewareOnlyResult := testing.Benchmark(BenchmarkMiddlewareOnly)
	normalizationResult := testing.Benchmark(BenchmarkRouteNormalization)
	tagCreationResult := testing.Benchmark(BenchmarkTagCreation)
	emitterCallsResult := testing.Benchmark(BenchmarkEmitterCalls)
	histogramResult := testing.Benchmark(BenchmarkHistogramConstruction)
	activeGaugeResult := testing.Benchmark(BenchmarkActiveRequestsGauge)
	requestProcessingResult := testing.Benchmark(BenchmarkRequestProcessing)
	requestProcessingNoNormResult := testing.Benchmark(BenchmarkRequestProcessingNoNormalization)
	requestProcessingNoMetricsResult := testing.Benchmark(BenchmarkRequestProcessingNoMetrics)

	t.Logf("Baseline (no middleware): %v", baselineResult)
	t.Logf("Middleware wrapper only: %v", middlewareOnlyResult)
	t.Logf("Route normalization: %v", normalizationResult)
	t.Logf("Tag creation: %v", tagCreationResult)
	t.Logf("Emitter calls: %v", emitterCallsResult)
	t.Logf("Histogram construction: %v", histogramResult)
	t.Logf("Active requests gauge: %v", activeGaugeResult)
	t.Logf("Full request processing: %v", requestProcessingResult)
	t.Logf("Request processing (no normalization): %v", requestProcessingNoNormResult)
	t.Logf("Request processing (no metrics): %v", requestProcessingNoMetricsResult)

	// Calculate overhead contributions
	baselineNs := baselineResult.NsPerOp()
	middlewareOnlyNs := middlewareOnlyResult.NsPerOp()
	normalizationNs := normalizationResult.NsPerOp()
	tagCreationNs := tagCreationResult.NsPerOp()
	emitterCallsNs := emitterCallsResult.NsPerOp()
	histogramNs := histogramResult.NsPerOp()
	activeGaugeNs := activeGaugeResult.NsPerOp()
	requestProcessingNs := requestProcessingResult.NsPerOp()
	requestProcessingNoNormNs := requestProcessingNoNormResult.NsPerOp()
	requestProcessingNoMetricsNs := requestProcessingNoMetricsResult.NsPerOp()

	t.Logf("\n=== OVERHEAD BREAKDOWN ===")
	t.Logf("Baseline: %v", baselineNs)
	t.Logf("Middleware wrapper: %v (%.1fx)", middlewareOnlyNs, float64(middlewareOnlyNs)/float64(baselineNs)*100)
	t.Logf("Route normalization: %v (%.1fx)", normalizationNs, float64(normalizationNs)/float64(baselineNs)*100)
	t.Logf("Tag creation: %v (%.1fx)", tagCreationNs, float64(tagCreationNs)/float64(baselineNs)*100)
	t.Logf("Emitter calls: %v (%.1fx)", emitterCallsNs, float64(emitterCallsNs)/float64(baselineNs)*100)
	t.Logf("Histogram construction: %v (%.1fx)", histogramNs, float64(histogramNs)/float64(baselineNs)*100)
	t.Logf("Active requests gauge: %v (%.1fx)", activeGaugeNs, float64(activeGaugeNs)/float64(baselineNs)*100)
	t.Logf("Full request processing: %v (%.1fx)", requestProcessingNs, float64(requestProcessingNs)/float64(baselineNs)*100)
	t.Logf("Request processing (no normalization): %v (%.1fx)", requestProcessingNoNormNs, float64(requestProcessingNoNormNs)/float64(baselineNs)*100)
	t.Logf("Request processing (no metrics): %v (%.1fx)", requestProcessingNoMetricsNs, float64(requestProcessingNoMetricsNs)/float64(baselineNs)*100)

	// Identify major contributors
	totalOverhead := requestProcessingNs - baselineNs
	normalizationOverhead := requestProcessingNoNormNs - requestProcessingNoMetricsNs
	metricsOverhead := requestProcessingNs - requestProcessingNoNormNs

	t.Logf("\n=== MAJOR CONTRIBUTORS ===")
	t.Logf("Total overhead: %v (%.1fx)", totalOverhead, float64(totalOverhead)/float64(baselineNs)*100)
	t.Logf("Route normalization contribution: %v (%.1fx)", normalizationOverhead, float64(normalizationOverhead)/float64(totalOverhead)*100)
	t.Logf("Metrics emission contribution: %v (%.1fx)", metricsOverhead, float64(metricsOverhead)/float64(totalOverhead)*100)
}
