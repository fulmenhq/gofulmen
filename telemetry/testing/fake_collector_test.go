package testing

import (
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
)

func TestFakeCollector_Counter(t *testing.T) {
	fc := NewFakeCollector()

	err := fc.Counter("test_counter", 5.0, map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("Counter() failed: %v", err)
	}

	if fc.CountMetrics() != 1 {
		t.Errorf("Expected 1 metric, got %d", fc.CountMetrics())
	}

	metrics := fc.GetMetricsByName("test_counter")
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric named 'test_counter', got %d", len(metrics))
	}

	m := metrics[0]
	if m.Type != MetricTypeCounter {
		t.Errorf("Expected type counter, got %v", m.Type)
	}
	if m.Value != 5.0 {
		t.Errorf("Expected value 5.0, got %v", m.Value)
	}
	if m.Tags["env"] != "test" {
		t.Errorf("Expected tag env=test, got %v", m.Tags["env"])
	}
}

func TestFakeCollector_Gauge(t *testing.T) {
	fc := NewFakeCollector()

	err := fc.Gauge("cpu_usage", 75.5, map[string]string{"host": "server1"})
	if err != nil {
		t.Fatalf("Gauge() failed: %v", err)
	}

	metrics := fc.GetMetricsByType(MetricTypeGauge)
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 gauge metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Name != "cpu_usage" {
		t.Errorf("Expected name cpu_usage, got %s", m.Name)
	}
	if m.Value != 75.5 {
		t.Errorf("Expected value 75.5, got %v", m.Value)
	}
}

func TestFakeCollector_Histogram(t *testing.T) {
	fc := NewFakeCollector()

	duration := 100 * time.Millisecond
	err := fc.Histogram("request_duration_ms", duration, map[string]string{"endpoint": "/api"})
	if err != nil {
		t.Fatalf("Histogram() failed: %v", err)
	}

	metrics := fc.GetMetricsByType(MetricTypeHistogram)
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 histogram metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Name != "request_duration_ms" {
		t.Errorf("Expected name request_duration_ms, got %s", m.Name)
	}
	if m.Value != duration {
		t.Errorf("Expected value %v, got %v", duration, m.Value)
	}
	if m.Unit != "ms" {
		t.Errorf("Expected unit ms, got %s", m.Unit)
	}
}

func TestFakeCollector_MultipleMetrics(t *testing.T) {
	fc := NewFakeCollector()

	_ = fc.Counter("counter1", 1.0, nil)
	_ = fc.Counter("counter2", 2.0, nil)
	_ = fc.Gauge("gauge1", 3.0, nil)
	_ = fc.Histogram("hist1", time.Second, nil)

	if fc.CountMetrics() != 4 {
		t.Errorf("Expected 4 metrics, got %d", fc.CountMetrics())
	}

	counters := fc.GetMetricsByType(MetricTypeCounter)
	if len(counters) != 2 {
		t.Errorf("Expected 2 counters, got %d", len(counters))
	}

	gauges := fc.GetMetricsByType(MetricTypeGauge)
	if len(gauges) != 1 {
		t.Errorf("Expected 1 gauge, got %d", len(gauges))
	}

	histograms := fc.GetMetricsByType(MetricTypeHistogram)
	if len(histograms) != 1 {
		t.Errorf("Expected 1 histogram, got %d", len(histograms))
	}
}

func TestFakeCollector_HasMetric(t *testing.T) {
	fc := NewFakeCollector()

	if fc.HasMetric("nonexistent") {
		t.Error("HasMetric returned true for nonexistent metric")
	}

	_ = fc.Counter("exists", 1.0, nil)

	if !fc.HasMetric("exists") {
		t.Error("HasMetric returned false for existing metric")
	}
}

func TestFakeCollector_Reset(t *testing.T) {
	fc := NewFakeCollector()

	_ = fc.Counter("test", 1.0, nil)
	_ = fc.Gauge("test2", 2.0, nil)

	if fc.CountMetrics() != 2 {
		t.Errorf("Expected 2 metrics before reset, got %d", fc.CountMetrics())
	}

	fc.Reset()

	if fc.CountMetrics() != 0 {
		t.Errorf("Expected 0 metrics after reset, got %d", fc.CountMetrics())
	}
}

func TestFakeCollector_TagIsolation(t *testing.T) {
	fc := NewFakeCollector()

	tags := map[string]string{"key": "value"}
	_ = fc.Counter("test", 1.0, tags)

	tags["key"] = "modified"

	metrics := fc.GetMetricsByName("test")
	if metrics[0].Tags["key"] != "value" {
		t.Error("Tag modification affected recorded metric")
	}
}

func TestFakeCollector_ConcurrentAccess(t *testing.T) {
	fc := NewFakeCollector()

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			_ = fc.Counter("test", float64(i), nil)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = fc.CountMetrics()
			_ = fc.HasMetric("test")
		}
		done <- true
	}()

	<-done
	<-done

	if fc.CountMetrics() != 100 {
		t.Errorf("Expected 100 metrics after concurrent operations, got %d", fc.CountMetrics())
	}
}

func TestFakeCollector_HistogramSummary(t *testing.T) {
	fc := NewFakeCollector()

	summary := telemetry.HistogramSummary{
		Count: 100,
		Sum:   5000.0,
		Buckets: []telemetry.HistogramBucket{
			{LE: 1, Count: 10},
			{LE: 5, Count: 50},
			{LE: 10, Count: 90},
			{LE: 50, Count: 100},
		},
	}

	err := fc.HistogramSummary("api_response_time_ms", summary, map[string]string{"service": "api"})
	if err != nil {
		t.Fatalf("HistogramSummary() failed: %v", err)
	}

	metrics := fc.GetMetricsByType(MetricTypeHistogram)
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 histogram metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Name != "api_response_time_ms" {
		t.Errorf("Expected name api_response_time_ms, got %s", m.Name)
	}

	recordedSummary, ok := m.Value.(telemetry.HistogramSummary)
	if !ok {
		t.Fatalf("Expected value to be HistogramSummary, got %T", m.Value)
	}

	if recordedSummary.Count != 100 {
		t.Errorf("Expected count 100, got %d", recordedSummary.Count)
	}
	if recordedSummary.Sum != 5000.0 {
		t.Errorf("Expected sum 5000.0, got %f", recordedSummary.Sum)
	}
	if len(recordedSummary.Buckets) != 4 {
		t.Errorf("Expected 4 buckets, got %d", len(recordedSummary.Buckets))
	}
}
