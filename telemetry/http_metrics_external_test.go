package telemetry_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"github.com/stretchr/testify/assert"
)

// Test HTTP metrics middleware emits correct metrics
func TestHTTPMetricsMiddleware_Metrics(t *testing.T) {
	emitter := &mockEmitter{}

	middleware := telemetry.HTTPMetricsMiddleware(emitter, telemetry.WithServiceName("test-service"))

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!")) // 13 bytes
	}))

	// Make a request with 4 bytes of body
	req := httptest.NewRequest("GET", "/users/123/profile", strings.NewReader("test"))
	req.ContentLength = 4
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, World!", w.Body.String())

	// Verify metrics were emitted correctly
	assert.True(t, emitter.calledCounter, "Counter should be called")
	assert.Equal(t, metrics.HTTPRequestsTotal, emitter.counterName)
	assert.Equal(t, float64(1), emitter.counterValue)
	assert.Equal(t, map[string]string{
		metrics.TagMethod:  "GET",
		metrics.TagRoute:   "/users/{id}/profile", // normalized
		metrics.TagStatus:  "200",
		metrics.TagService: "test-service",
	}, emitter.counterTags)

	assert.True(t, emitter.calledHistogram, "Duration histogram should be called")
	assert.Equal(t, metrics.HTTPRequestDurationSeconds, emitter.histogramName)
	assert.Greater(t, emitter.histogramDuration, time.Duration(0))

	assert.True(t, emitter.calledSizeHistogramReq, "Request size histogram should be called")
	assert.Equal(t, metrics.HTTPRequestSizeBytes, emitter.sizeHistogramNameReq)
	assert.Equal(t, float64(4), emitter.sizeHistogramSummaryReq.Sum) // request body size
	// Validate bucket counts for 4-byte request (should be in 1024 bucket)
	assert.Len(t, emitter.sizeHistogramSummaryReq.Buckets, 7) // 6 default + Inf
	// For 4-byte request: first bucket (1024) should have count=1 since 4 <= 1024
	assert.Equal(t, int64(1), emitter.sizeHistogramSummaryReq.Buckets[0].Count) // First bucket >= size

	assert.True(t, emitter.calledSizeHistogramResp, "Response size histogram should be called")
	assert.Equal(t, metrics.HTTPResponseSizeBytes, emitter.sizeHistogramNameResp)
	assert.Equal(t, float64(13), emitter.sizeHistogramSummaryResp.Sum) // response body size
	// Validate bucket counts for 13-byte response (should be in 1024 bucket)
	assert.Len(t, emitter.sizeHistogramSummaryResp.Buckets, 7)                   // 6 default + Inf
	assert.Equal(t, int64(1), emitter.sizeHistogramSummaryResp.Buckets[0].Count) // 1024 bucket: 13 <= 1024

	assert.True(t, emitter.calledGauge, "Active requests gauge should be called")
	assert.Equal(t, metrics.HTTPActiveRequests, emitter.gaugeName)
	// Gauge value may be 1 (during request) or 0 (after request) depending on timing
	assert.True(t, emitter.gaugeValue == 0 || emitter.gaugeValue == 1, "Active requests gauge should be 0 or 1")

	// Validate active requests gauge has minimal label set (service only)
	expectedActiveTags := map[string]string{
		metrics.TagService: "test-service",
	}
	assert.Equal(t, expectedActiveTags, emitter.gaugeTags, "Active requests gauge should only have service tag")

	// Validate that no gauge was emitted for size metrics (only histograms should be used)
	assert.NotEqual(t, metrics.HTTPRequestSizeBytes, emitter.gaugeName, "Request size should not be emitted as gauge")
	assert.NotEqual(t, metrics.HTTPResponseSizeBytes, emitter.gaugeName, "Response size should not be emitted as gauge")
}

func TestDefaultRouteNormalizer(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users/123", "/users/{id}"},
		{"POST", "/users/123/profile/456", "/users/{id}/profile/{id}"},
		{"GET", "/api/v1/users/550e8400-e29b-41d4-a716-446655440000", "/api/v1/users/{uuid}"},
		{"GET", "/static/css/main.css", "/static/css/main.css"},
		{"GET", "/", "/"},
		{"GET", "/api/search?q=test", "/api/search"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := telemetry.DefaultRouteNormalizer(tt.method, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHTTPMetricsMiddleware_SizeBucketValidation validates histogram bucket construction
// for different request/response sizes to ensure cumulative counts are correct
func TestHTTPMetricsMiddleware_SizeBucketValidation(t *testing.T) {
	emitter := &mockEmitter{}

	middleware := telemetry.HTTPMetricsMiddleware(emitter, telemetry.WithServiceName("test-service"))

	// Test with a large response (15KB) to validate mid/upper bucket behavior
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write 15KB of data - should land in the 1048576 (1MB) bucket, not 10240 (10KB) bucket
		w.Write(make([]byte, 15*1024)) // 15KB
	}))

	// Make a request
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Validate response size histogram buckets for 15KB response
	assert.True(t, emitter.calledSizeHistogramResp, "Response size histogram should be called")
	assert.Equal(t, metrics.HTTPResponseSizeBytes, emitter.sizeHistogramNameResp)
	assert.Equal(t, float64(15*1024), emitter.sizeHistogramSummaryResp.Sum) // 15KB

	// Default buckets: [1024, 10240, 102400, 1048576, 10485760, 104857600]
	// For 15KB (15360 bytes):
	// - 1024 bucket: 0 (15KB > 1024)
	// - 10240 bucket: 0 (15KB > 10240)
	// - 102400 bucket: 1 (15KB <= 102400) - first bucket >= size
	// - 1048576 bucket: 1 (cumulative)
	// - 10485760 bucket: 1 (cumulative)
	// - 104857600 bucket: 1 (cumulative)
	// - +Inf bucket: 1 (cumulative)

	buckets := emitter.sizeHistogramSummaryResp.Buckets
	assert.Len(t, buckets, 7) // 6 default + Inf

	// Validate individual bucket counts
	assert.Equal(t, int64(0), buckets[0].Count) // 1024: 15KB > 1024
	assert.Equal(t, int64(0), buckets[1].Count) // 10240: 15KB > 10240
	assert.Equal(t, int64(1), buckets[2].Count) // 102400: 15KB <= 102400 (first containing bucket)
	assert.Equal(t, int64(1), buckets[3].Count) // 1048576: cumulative
	assert.Equal(t, int64(1), buckets[4].Count) // 10485760: cumulative
	assert.Equal(t, int64(1), buckets[5].Count) // 104857600: cumulative

	// Validate +Inf bucket
	assert.Equal(t, math.Inf(1), buckets[6].LE)
	assert.Equal(t, int64(1), buckets[6].Count) // +Inf: always 1 for single observation
}

// mockEmitter implements telemetry.MetricsEmitter for testing
type mockEmitter struct {
	calledCounter bool
	counterName   string
	counterValue  float64
	counterTags   map[string]string

	calledHistogram   bool
	histogramName     string
	histogramDuration time.Duration

	calledSizeHistogramReq  bool
	sizeHistogramNameReq    string
	sizeHistogramSummaryReq telemetry.HistogramSummary

	calledSizeHistogramResp  bool
	sizeHistogramNameResp    string
	sizeHistogramSummaryResp telemetry.HistogramSummary

	calledGauge bool
	gaugeName   string
	gaugeValue  float64
	gaugeTags   map[string]string
}

func (m *mockEmitter) Counter(name string, value float64, tags map[string]string) error {
	m.calledCounter = true
	m.counterName = name
	m.counterValue = value
	m.counterTags = tags
	return nil
}

func (m *mockEmitter) Histogram(name string, duration time.Duration, tags map[string]string) error {
	m.calledHistogram = true
	m.histogramName = name
	m.histogramDuration = duration
	return nil
}

func (m *mockEmitter) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
	if name == metrics.HTTPRequestSizeBytes {
		m.calledSizeHistogramReq = true
		m.sizeHistogramNameReq = name
		m.sizeHistogramSummaryReq = summary
	} else if name == metrics.HTTPResponseSizeBytes {
		m.calledSizeHistogramResp = true
		m.sizeHistogramNameResp = name
		m.sizeHistogramSummaryResp = summary
	}
	return nil
}

func (m *mockEmitter) Gauge(name string, value float64, tags map[string]string) error {
	m.calledGauge = true
	m.gaugeName = name
	m.gaugeValue = value
	m.gaugeTags = tags
	return nil
}
