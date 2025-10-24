package testing

import (
	"sync"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
)

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

type RecordedMetric struct {
	Name      string
	Type      MetricType
	Value     interface{}
	Tags      map[string]string
	Unit      string
	Timestamp time.Time
}

type FakeCollector struct {
	mu      sync.RWMutex
	metrics []RecordedMetric
}

func NewFakeCollector() *FakeCollector {
	return &FakeCollector{
		metrics: make([]RecordedMetric, 0),
	}
}

func (fc *FakeCollector) Counter(name string, value float64, tags map[string]string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.metrics = append(fc.metrics, RecordedMetric{
		Name:      name,
		Type:      MetricTypeCounter,
		Value:     value,
		Tags:      copyTags(tags),
		Timestamp: time.Now(),
	})
	return nil
}

func (fc *FakeCollector) Gauge(name string, value float64, tags map[string]string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.metrics = append(fc.metrics, RecordedMetric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Tags:      copyTags(tags),
		Timestamp: time.Now(),
	})
	return nil
}

func (fc *FakeCollector) Histogram(name string, value time.Duration, tags map[string]string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.metrics = append(fc.metrics, RecordedMetric{
		Name:      name,
		Type:      MetricTypeHistogram,
		Value:     value,
		Tags:      copyTags(tags),
		Unit:      "ms",
		Timestamp: time.Now(),
	})
	return nil
}

func (fc *FakeCollector) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.metrics = append(fc.metrics, RecordedMetric{
		Name:      name,
		Type:      MetricTypeHistogram,
		Value:     summary,
		Tags:      copyTags(tags),
		Unit:      "ms",
		Timestamp: time.Now(),
	})
	return nil
}

func (fc *FakeCollector) GetMetrics() []RecordedMetric {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	result := make([]RecordedMetric, len(fc.metrics))
	copy(result, fc.metrics)
	return result
}

func (fc *FakeCollector) GetMetricsByName(name string) []RecordedMetric {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	var result []RecordedMetric
	for _, m := range fc.metrics {
		if m.Name == name {
			result = append(result, m)
		}
	}
	return result
}

func (fc *FakeCollector) GetMetricsByType(metricType MetricType) []RecordedMetric {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	var result []RecordedMetric
	for _, m := range fc.metrics {
		if m.Type == metricType {
			result = append(result, m)
		}
	}
	return result
}

func (fc *FakeCollector) CountMetrics() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.metrics)
}

func (fc *FakeCollector) CountMetricsByName(name string) int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	count := 0
	for _, m := range fc.metrics {
		if m.Name == name {
			count++
		}
	}
	return count
}

func (fc *FakeCollector) HasMetric(name string) bool {
	return fc.CountMetricsByName(name) > 0
}

func (fc *FakeCollector) Reset() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.metrics = make([]RecordedMetric, 0)
}

func copyTags(tags map[string]string) map[string]string {
	if tags == nil {
		return nil
	}
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}
