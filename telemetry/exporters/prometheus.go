// Package exporters provides custom metric exporters for various monitoring systems
package exporters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
)

// PrometheusExporter implements a basic Prometheus metrics exporter
type PrometheusExporter struct {
	mu       sync.RWMutex
	metrics  []telemetry.MetricsEvent
	prefix   string
	endpoint string
	server   *http.Server
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(prefix, endpoint string) *PrometheusExporter {
	return &PrometheusExporter{
		metrics:  make([]telemetry.MetricsEvent, 0),
		prefix:   prefix,
		endpoint: endpoint,
	}
}

// Counter implements telemetry.MetricsEmitter
func (e *PrometheusExporter) Counter(name string, value float64, tags map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := telemetry.MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      telemetry.TypeCounter,
		Value:     value,
		Tags:      tags,
	}
	e.metrics = append(e.metrics, event)
	return nil
}

// Histogram implements telemetry.MetricsEmitter
func (e *PrometheusExporter) Histogram(name string, duration time.Duration, tags map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := telemetry.MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      telemetry.TypeHistogram,
		Value:     float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
		Tags:      tags,
		Unit:      "ms",
	}
	e.metrics = append(e.metrics, event)
	return nil
}

// HistogramSummary implements telemetry.MetricsEmitter
func (e *PrometheusExporter) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := telemetry.MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      telemetry.TypeHistogram,
		Value:     summary,
		Tags:      tags,
		Unit:      "ms",
	}
	e.metrics = append(e.metrics, event)
	return nil
}

// Gauge implements telemetry.MetricsEmitter
func (e *PrometheusExporter) Gauge(name string, value float64, tags map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := telemetry.MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      telemetry.TypeGauge,
		Value:     value,
		Tags:      tags,
	}
	e.metrics = append(e.metrics, event)
	return nil
}

// Start starts the HTTP server for Prometheus metrics endpoint
func (e *PrometheusExporter) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", e.metricsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("<h1>Prometheus Metrics Exporter</h1><p><a href='/metrics'>Metrics</a></p>")); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
	})

	e.server = &http.Server{
		Addr:              e.endpoint,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Prometheus exporter server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Stop stops the HTTP server
func (e *PrometheusExporter) Stop() error {
	if e.server != nil {
		return e.server.Close()
	}
	return nil
}

// metricsHandler handles Prometheus metrics requests
func (e *PrometheusExporter) metricsHandler(w http.ResponseWriter, r *http.Request) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Group metrics by name and type for Prometheus format
	metricGroups := make(map[string][]telemetry.MetricsEvent)
	for _, metric := range e.metrics {
		key := fmt.Sprintf("%s_%s", metric.Name, e.getMetricType(metric))
		metricGroups[key] = append(metricGroups[key], metric)
	}

	// Write metrics in Prometheus format
	for _, metrics := range metricGroups {
		if len(metrics) == 0 {
			continue
		}

		// Get the first metric to determine type
		firstMetric := metrics[0]

		switch firstMetric.Type {
		case telemetry.TypeCounter:
			e.writeCounterMetrics(w, metrics)
		case telemetry.TypeGauge:
			e.writeGaugeMetrics(w, metrics)
		case telemetry.TypeHistogram:
			e.writeHistogramMetrics(w, metrics)
		}
	}
}

// formatPrometheusName converts metric name to Prometheus format
func (e *PrometheusExporter) formatPrometheusName(name string) string {
	// Add prefix if specified
	if e.prefix != "" {
		name = e.prefix + "_" + name
	}

	// Convert to Prometheus naming convention (snake_case)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return strings.ToLower(name)
}

// formatPrometheusLabels converts tags to Prometheus label format
func (e *PrometheusExporter) formatPrometheusLabels(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	labels := make([]string, 0, len(tags))
	for key, value := range tags {
		// Escape quotes in label values
		escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
		labels = append(labels, fmt.Sprintf(`%s="%s"`, key, escapedValue))
	}

	return strings.Join(labels, ",")
}

// formatPrometheusLabelsWithAdditional converts tags to Prometheus label format with additional label
func (e *PrometheusExporter) formatPrometheusLabelsWithAdditional(tags map[string]string, additionalKey, additionalValue string) string {
	// Create a copy of tags and add the additional label
	allTags := make(map[string]string)
	for k, v := range tags {
		allTags[k] = v
	}
	allTags[additionalKey] = additionalValue
	return e.formatPrometheusLabels(allTags)
}

// getMetricType determines the metric type from the event
func (e *PrometheusExporter) getMetricType(event telemetry.MetricsEvent) string {
	return string(event.Type)
}

// writeCounterMetrics writes counter metrics in Prometheus format
func (e *PrometheusExporter) writeCounterMetrics(w io.Writer, metrics []telemetry.MetricsEvent) {
	for _, metric := range metrics {
		name := e.formatPrometheusName(metric.Name + "_total")
		labels := e.formatPrometheusLabels(metric.Tags)
		value := e.extractMetricValue(metric.Value)

		if labels != "" {
			if _, err := fmt.Fprintf(w, "%s{%s} %f\n", name, labels, value); err != nil {
				fmt.Printf("Error writing counter metric: %v\n", err)
				return
			}
		} else {
			if _, err := fmt.Fprintf(w, "%s %f\n", name, value); err != nil {
				fmt.Printf("Error writing counter metric: %v\n", err)
				return
			}
		}
	}
}

// writeGaugeMetrics writes gauge metrics in Prometheus format
func (e *PrometheusExporter) writeGaugeMetrics(w io.Writer, metrics []telemetry.MetricsEvent) {
	for _, metric := range metrics {
		name := e.formatPrometheusName(metric.Name + "_gauge")
		labels := e.formatPrometheusLabels(metric.Tags)
		value := e.extractMetricValue(metric.Value)

		if labels != "" {
			if _, err := fmt.Fprintf(w, "%s{%s} %f\n", name, labels, value); err != nil {
				fmt.Printf("Error writing gauge metric: %v\n", err)
				return
			}
		} else {
			if _, err := fmt.Fprintf(w, "%s %f\n", name, value); err != nil {
				fmt.Printf("Error writing gauge metric: %v\n", err)
				return
			}
		}
	}
}

// writeHistogramMetrics writes histogram metrics in Prometheus format
func (e *PrometheusExporter) writeHistogramMetrics(w io.Writer, metrics []telemetry.MetricsEvent) {
	for _, metric := range metrics {
		switch v := metric.Value.(type) {
		case telemetry.HistogramSummary:
			// Write bucket series
			for _, bucket := range v.Buckets {
				bucketLabels := e.formatPrometheusLabelsWithAdditional(metric.Tags, "le", fmt.Sprintf("%g", bucket.LE))
				name := e.formatPrometheusName(metric.Name + "_bucket")
				if bucketLabels != "" {
					if _, err := fmt.Fprintf(w, "%s{%s} %d\n", name, bucketLabels, bucket.Count); err != nil {
						fmt.Printf("Error writing histogram bucket: %v\n", err)
						return
					}
				} else {
					if _, err := fmt.Fprintf(w, "%s %d\n", name, bucket.Count); err != nil {
						fmt.Printf("Error writing histogram bucket: %v\n", err)
						return
					}
				}
			}

			// Write sum and count
			sumName := e.formatPrometheusName(metric.Name + "_sum")
			countName := e.formatPrometheusName(metric.Name + "_count")
			labels := e.formatPrometheusLabels(metric.Tags)

			if labels != "" {
				if _, err := fmt.Fprintf(w, "%s{%s} %f\n", sumName, labels, v.Sum); err != nil {
					fmt.Printf("Error writing histogram sum: %v\n", err)
					return
				}
				if _, err := fmt.Fprintf(w, "%s{%s} %d\n", countName, labels, v.Count); err != nil {
					fmt.Printf("Error writing histogram count: %v\n", err)
					return
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s %f\n", sumName, v.Sum); err != nil {
					fmt.Printf("Error writing histogram sum: %v\n", err)
					return
				}
				if _, err := fmt.Fprintf(w, "%s %d\n", countName, v.Count); err != nil {
					fmt.Printf("Error writing histogram count: %v\n", err)
					return
				}
			}
		case float64:
			// Single histogram value - treat as a simple metric for now
			name := e.formatPrometheusName(metric.Name)
			labels := e.formatPrometheusLabels(metric.Tags)
			if labels != "" {
				if _, err := fmt.Fprintf(w, "%s{%s} %f\n", name, labels, v); err != nil {
					fmt.Printf("Error writing histogram value: %v\n", err)
					return
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s %f\n", name, v); err != nil {
					fmt.Printf("Error writing histogram value: %v\n", err)
					return
				}
			}
		}
	}
}

// extractMetricValue extracts the numeric value from a metric event
func (e *PrometheusExporter) extractMetricValue(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case telemetry.HistogramSummary:
		// For histograms, export the sum as the main metric
		return v.Sum
	default:
		return 0.0
	}
}

// WriteMetrics writes current metrics to a writer (for debugging)
func (e *PrometheusExporter) WriteMetrics(w io.Writer) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, metric := range e.metrics {
		jsonData, err := json.Marshal(metric)
		if err != nil {
			return err
		}
		if _, err := w.Write(jsonData); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

// Clear clears all stored metrics
func (e *PrometheusExporter) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.metrics = e.metrics[:0]
}

// GetMetrics returns a copy of current metrics (for testing)
func (e *PrometheusExporter) GetMetrics() []telemetry.MetricsEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]telemetry.MetricsEvent, len(e.metrics))
	copy(result, e.metrics)
	return result
}
