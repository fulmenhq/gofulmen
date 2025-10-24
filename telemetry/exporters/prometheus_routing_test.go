package exporters

import (
	"bytes"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResponseWriter implements http.ResponseWriter for testing
type mockResponseWriter struct {
	buf        *bytes.Buffer
	header     http.Header
	statusCode int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		buf:    &bytes.Buffer{},
		header: make(http.Header),
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	return m.buf.Write(data)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func (m *mockResponseWriter) String() string {
	return m.buf.String()
}

// TestPrometheusMetricTypeRouting verifies that different metric types are routed correctly
func TestPrometheusMetricTypeRouting(t *testing.T) {
	exporter := NewPrometheusExporter("test", ":9091")

	// Test counter routing
	err := exporter.Counter("requests_total", 100, map[string]string{"status": "200"})
	require.NoError(t, err)

	// Test gauge routing
	err = exporter.Gauge("cpu_usage_percent", 75.5, map[string]string{"host": "server1"})
	require.NoError(t, err)

	// Test histogram routing with single value
	err = exporter.Histogram("request_duration_ms", 50*time.Millisecond, map[string]string{"endpoint": "/api"})
	require.NoError(t, err)

	// Test histogram routing with summary
	summary := telemetry.HistogramSummary{
		Count: 100,
		Sum:   5000,
		Buckets: []telemetry.HistogramBucket{
			{LE: 1, Count: 10},
			{LE: 5, Count: 50},
			{LE: 10, Count: 90},
			{LE: 50, Count: 100},
			{LE: math.Inf(1), Count: 100}, // +Inf bucket
		},
	}
	err = exporter.HistogramSummary("api_response_time_ms", summary, map[string]string{"service": "api"})
	require.NoError(t, err)

	// Get the metrics output using the Prometheus format
	mockWriter := newMockResponseWriter()
	exporter.metricsHandler(mockWriter, nil)

	output := mockWriter.String()
	t.Logf("Prometheus output:\n%s", output)
	t.Logf("Number of metrics in exporter: %d", len(exporter.GetMetrics()))

	// Debug: check the actual metrics
	metrics := exporter.GetMetrics()
	for i, metric := range metrics {
		t.Logf("Metric %d: Name=%s, Type=%s, Value=%v", i, metric.Name, metric.Type, metric.Value)
	}

	// Verify counter is formatted correctly (should end with _total)
	assert.Contains(t, output, "test_requests_total_total{status=\"200\"} 100")

	// Verify gauge is formatted correctly (should end with _gauge)
	assert.Contains(t, output, "test_cpu_usage_percent_gauge{host=\"server1\"} 75.5")

	// Verify histogram single value is handled
	assert.Contains(t, output, "test_request_duration_ms{endpoint=\"/api\"} 50")

	// Verify histogram summary is formatted correctly with buckets, sum, and count
	assert.Contains(t, output, "test_api_response_time_ms_bucket")
	assert.Contains(t, output, "le=\"1\"")
	assert.Contains(t, output, "le=\"5\"")
	assert.Contains(t, output, "le=\"10\"")
	assert.Contains(t, output, "le=\"50\"")
	assert.Contains(t, output, "le=\"+Inf\"")
	assert.Contains(t, output, "test_api_response_time_ms_sum{service=\"api\"} 5000")
	assert.Contains(t, output, "test_api_response_time_ms_count{service=\"api\"} 100")
}

// TestPrometheusMetricTypeRoutingInHandler tests the HTTP handler output
func TestPrometheusMetricTypeRoutingInHandler(t *testing.T) {
	exporter := NewPrometheusExporter("app", ":9092")

	// Add different types of metrics
	if err := exporter.Counter("http_requests_total", 1000, map[string]string{"method": "GET", "status": "200"}); err != nil {
		t.Logf("Error adding counter: %v", err)
	}
	if err := exporter.Gauge("memory_usage_bytes", 1073741824, map[string]string{"host": "app1"}); err != nil {
		t.Logf("Error adding gauge: %v", err)
	}

	summary := telemetry.HistogramSummary{
		Count: 50,
		Sum:   2500,
		Buckets: []telemetry.HistogramBucket{
			{LE: 10, Count: 5},
			{LE: 50, Count: 25},
			{LE: 100, Count: 45},
			{LE: math.Inf(1), Count: 50},
		},
	}
	if err := exporter.HistogramSummary("request_duration_ms", summary, map[string]string{"endpoint": "/users"}); err != nil {
		t.Logf("Error adding histogram summary: %v", err)
	}

	// Test the HTTP handler using the proper format
	mockWriter := newMockResponseWriter()
	exporter.metricsHandler(mockWriter, nil)

	output := mockWriter.String()
	t.Logf("HTTP handler output:\n%s", output)

	// Verify counter formatting
	assert.Contains(t, output, "app_http_requests_total_total{method=\"GET\",status=\"200\"} 1000")

	// Verify gauge formatting
	assert.Contains(t, output, "app_memory_usage_bytes_gauge{host=\"app1\"} 1073741824")

	// Verify histogram formatting with proper Prometheus conventions
	lines := strings.Split(output, "\n")

	// Count bucket lines
	bucketLines := 0
	sumFound := false
	countFound := false

	for _, line := range lines {
		if strings.Contains(line, "app_request_duration_ms_bucket") {
			bucketLines++
			assert.Contains(t, line, "endpoint=\"/users\"")
			assert.Contains(t, line, "le=")
		}
		if strings.Contains(line, "app_request_duration_ms_sum{endpoint=\"/users\"}") {
			sumFound = true
		}
		if strings.Contains(line, "app_request_duration_ms_count{endpoint=\"/users\"}") {
			countFound = true
		}
	}

	assert.Equal(t, 4, bucketLines, "Should have 4 bucket lines")
	assert.True(t, sumFound, "Should have sum line")
	assert.True(t, countFound, "Should have count line")
}
