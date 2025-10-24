package testing_test

import (
	"fmt"
	"time"

	telemetrytesting "github.com/fulmenhq/gofulmen/telemetry/testing"
)

func ExampleFakeCollector() {
	fc := telemetrytesting.NewFakeCollector()

	_ = fc.Counter("api_requests", 1, map[string]string{
		"endpoint": "/users",
		"status":   "200",
	})

	_ = fc.Gauge("memory_usage_bytes", 1024000, map[string]string{
		"service": "api",
	})

	_ = fc.Histogram("request_duration_ms", 150*time.Millisecond, map[string]string{
		"endpoint": "/users",
	})

	fmt.Printf("Total metrics collected: %d\n", fc.CountMetrics())
	fmt.Printf("Has api_requests: %v\n", fc.HasMetric("api_requests"))

	// Output:
	// Total metrics collected: 3
	// Has api_requests: true
}

func ExampleFakeCollector_GetMetricsByName() {
	fc := telemetrytesting.NewFakeCollector()

	_ = fc.Counter("requests", 1, map[string]string{"status": "200"})
	_ = fc.Counter("requests", 1, map[string]string{"status": "404"})
	_ = fc.Counter("requests", 1, map[string]string{"status": "500"})

	metrics := fc.GetMetricsByName("requests")
	fmt.Printf("Found %d request metrics\n", len(metrics))

	// Output:
	// Found 3 request metrics
}

func ExampleFakeCollector_GetMetricsByType() {
	fc := telemetrytesting.NewFakeCollector()

	_ = fc.Counter("counter1", 1, nil)
	_ = fc.Counter("counter2", 2, nil)
	_ = fc.Gauge("gauge1", 3, nil)
	_ = fc.Histogram("hist1", time.Second, nil)

	counters := fc.GetMetricsByType(telemetrytesting.MetricTypeCounter)
	gauges := fc.GetMetricsByType(telemetrytesting.MetricTypeGauge)
	histograms := fc.GetMetricsByType(telemetrytesting.MetricTypeHistogram)

	fmt.Printf("Counters: %d, Gauges: %d, Histograms: %d\n", len(counters), len(gauges), len(histograms))

	// Output:
	// Counters: 2, Gauges: 1, Histograms: 1
}
