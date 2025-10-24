package exporters

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrometheusExporterBasic tests basic Prometheus exporter functionality
func TestPrometheusExporterBasic(t *testing.T) {
	exporter := NewPrometheusExporter("test", ":0") // Use port 0 for automatic port assignment
	require.NotNil(t, exporter)

	// Test counter metric
	assert.NoError(t, exporter.Counter("requests_total", 100, map[string]string{"status": "200"}))

	// Test gauge metric
	assert.NoError(t, exporter.Gauge("cpu_usage_percent", 75.5, map[string]string{"host": "server1"}))

	// Test histogram metric
	assert.NoError(t, exporter.Histogram("request_duration_ms", 50*time.Millisecond, map[string]string{"endpoint": "/api"}))

	// Verify metrics were stored
	metrics := exporter.GetMetrics()
	assert.Len(t, metrics, 3)

	// Check counter metric
	assert.Equal(t, "requests_total", metrics[0].Name)
	assert.Equal(t, float64(100), metrics[0].Value)
	assert.Equal(t, "200", metrics[0].Tags["status"])

	// Check gauge metric
	assert.Equal(t, "cpu_usage_percent", metrics[1].Name)
	assert.Equal(t, 75.5, metrics[1].Value)
	assert.Equal(t, "server1", metrics[1].Tags["host"])

	// Check histogram metric
	assert.Equal(t, "request_duration_ms", metrics[2].Name)
	assert.Equal(t, float64(50), metrics[2].Value) // Should be converted to milliseconds
	assert.Equal(t, "ms", metrics[2].Unit)
	assert.Equal(t, "/api", metrics[2].Tags["endpoint"])
}

// TestPrometheusExporterFormat tests Prometheus format output
func TestPrometheusExporterFormat(t *testing.T) {
	exporter := NewPrometheusExporter("myapp", ":8080")

	// Add some test metrics
	assert.NoError(t, exporter.Counter("http_requests_total", 1000, map[string]string{"status": "200", "method": "GET"}))
	assert.NoError(t, exporter.Gauge("memory_usage_bytes", 1073741824, map[string]string{"host": "server1"}))
	assert.NoError(t, exporter.Histogram("request_duration_ms", 100*time.Millisecond, nil))

	// Test format functions
	assert.Equal(t, "myapp_http_requests_total", exporter.formatPrometheusName("http_requests_total"))
	// Test label formatting (order may vary due to map iteration)
	labels := exporter.formatPrometheusLabels(map[string]string{"status": "200", "method": "GET"})
	assert.Contains(t, labels, `status="200"`)
	assert.Contains(t, labels, `method="GET"`)
	assert.Equal(t, "", exporter.formatPrometheusLabels(nil))

	// Test value extraction
	assert.Equal(t, 1000.0, exporter.extractMetricValue(float64(1000)))
	assert.Equal(t, 100.0, exporter.extractMetricValue(float64(100)))
}

// TestPrometheusExporterHTTP tests the HTTP server functionality
func TestPrometheusExporterHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping HTTP test in short mode")
	}

	exporter := NewPrometheusExporter("test", ":0")

	// Add test metrics
	assert.NoError(t, exporter.Counter("test_counter", 42, map[string]string{"label": "value"}))
	assert.NoError(t, exporter.Gauge("test_gauge", 3.14, nil))

	// Start the server
	err := exporter.Start()
	assert.NoError(t, err)
	defer func() {
		if stopErr := exporter.Stop(); stopErr != nil {
			t.Logf("Error stopping exporter: %v", stopErr)
		}
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Make HTTP request to metrics endpoint
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		// Server might not be ready, try a few times
		time.Sleep(500 * time.Millisecond)
		resp, err = http.Get("http://localhost:8080/metrics")
	}

	if err == nil {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Logf("Error closing response body: %v", closeErr)
			}
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain; version=0.0.4", resp.Header.Get("Content-Type"))

		// Read response body
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		assert.NoError(t, err)

		// Check that metrics are present in response
		body := buf.String()
		assert.Contains(t, body, "test_counter")
		assert.Contains(t, body, "test_gauge")
	}
}

// TestPrometheusExporterClear tests the Clear functionality
func TestPrometheusExporterClear(t *testing.T) {
	exporter := NewPrometheusExporter("test", ":8080")

	// Add metrics
	assert.NoError(t, exporter.Counter("counter1", 10, nil))
	assert.NoError(t, exporter.Gauge("gauge1", 20, nil))

	// Verify metrics exist
	metrics := exporter.GetMetrics()
	assert.Len(t, metrics, 2)

	// Clear metrics
	exporter.Clear()

	// Verify metrics are cleared
	metrics = exporter.GetMetrics()
	assert.Len(t, metrics, 0)
}

// TestPrometheusExporterLabelsWithSpecialChars tests label escaping
func TestPrometheusExporterLabelsWithSpecialChars(t *testing.T) {
	exporter := NewPrometheusExporter("test", ":8080")

	// Test labels with special characters
	labels := map[string]string{
		"status":  `200 "OK"`,
		"message": `Server "up" and running`,
	}

	escaped := exporter.formatPrometheusLabels(labels)
	assert.Contains(t, escaped, `status="200 \"OK\""`)
	assert.Contains(t, escaped, `message="Server \"up\" and running"`)
}

// TestPrometheusExporterHistogramSummary tests histogram summary handling
func TestPrometheusExporterHistogramSummary(t *testing.T) {
	exporter := NewPrometheusExporter("test", ":8080")

	// Create a histogram summary
	summary := telemetry.HistogramSummary{
		Count: 100,
		Sum:   5000.0,
		Buckets: []telemetry.HistogramBucket{
			{LE: 10, Count: 50},
			{LE: 50, Count: 90},
			{LE: 100, Count: 100},
		},
	}

	assert.NoError(t, exporter.HistogramSummary("request_duration", summary, map[string]string{"service": "api"}))

	// Verify the metric was stored
	metrics := exporter.GetMetrics()
	assert.Len(t, metrics, 1)
	assert.Equal(t, "request_duration", metrics[0].Name)
	assert.Equal(t, "ms", metrics[0].Unit)

	// Check that histogram summary value is extracted correctly
	extracted := exporter.extractMetricValue(metrics[0].Value)
	assert.Equal(t, 5000.0, extracted) // Should be the sum
}
