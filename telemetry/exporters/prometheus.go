// Package exporters provides custom metric exporters for various monitoring systems
package exporters

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// PrometheusExporter implements a Prometheus metrics exporter with health instrumentation
type PrometheusExporter struct {
	mu      sync.RWMutex
	metrics []telemetry.MetricsEvent
	config  *PrometheusConfig
	server  *http.Server

	// HTTP handler with middleware
	httpHandler *httpHandler

	// Refresh tracking
	refreshInflight atomic.Int64
	restartCount    atomic.Int64
}

// NewPrometheusExporter creates a new Prometheus exporter (legacy constructor for backward compatibility)
func NewPrometheusExporter(prefix, endpoint string) *PrometheusExporter {
	config := DefaultPrometheusConfig()
	config.Prefix = prefix
	config.Endpoint = endpoint
	return NewPrometheusExporterWithConfig(config)
}

// NewPrometheusExporterWithConfig creates a new Prometheus exporter with the given configuration
func NewPrometheusExporterWithConfig(config *PrometheusConfig) *PrometheusExporter {
	if config == nil {
		config = DefaultPrometheusConfig()
	}
	if err := config.Validate(); err != nil {
		// Fall back to defaults if validation fails
		config = DefaultPrometheusConfig()
	}

	return &PrometheusExporter{
		metrics: make([]telemetry.MetricsEvent, 0),
		config:  config,
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

// Start starts the HTTP server for Prometheus metrics endpoint with instrumentation
func (e *PrometheusExporter) Start() error {
	// Emit restart metric
	tags := map[string]string{metrics.TagReason: metrics.RestartReasonManual}
	telemetry.EmitCounter(metrics.PrometheusExporterRestartsTotal, 1, tags)
	e.restartCount.Add(1)

	// Create HTTP handler with middleware
	e.httpHandler = newHTTPHandler(e, e.config)

	mux := http.NewServeMux()
	mux.Handle("/metrics", e.httpHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("<h1>Prometheus Metrics Exporter</h1><p><a href='/metrics'>Metrics</a></p>")); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
	})

	// Use a listener to get the actual address when using port :0
	listener, err := net.Listen("tcp", e.config.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to start Prometheus exporter: %w", err)
	}

	// Store the actual address (important for :0 random port assignment)
	actualAddr := listener.Addr().String()

	e.server = &http.Server{
		Addr:              actualAddr,
		Handler:           mux,
		ReadHeaderTimeout: e.config.ReadHeaderTimeout,
	}

	go func() {
		if err := e.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Prometheus exporter server error: %v\n", err)
			// Emit restart on crash
			crashTags := map[string]string{metrics.TagReason: metrics.RestartReasonPanicRecover}
			telemetry.EmitCounter(metrics.PrometheusExporterRestartsTotal, 1, crashTags)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// GetAddr returns the actual address the server is listening on
// This is useful when the endpoint is configured as ":0" (random port)
func (e *PrometheusExporter) GetAddr() string {
	if e.server != nil {
		return e.server.Addr
	}
	return e.config.Endpoint
}

// Stop stops the HTTP server
func (e *PrometheusExporter) Stop() error {
	if e.server != nil {
		return e.server.Close()
	}
	return nil
}

// metricsHandler handles Prometheus metrics requests with refresh instrumentation
func (e *PrometheusExporter) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Instrument refresh pipeline
	overallStart := time.Now()
	e.refreshInflight.Add(1)
	defer e.refreshInflight.Add(-1)

	// Emit inflight gauge
	telemetry.EmitGauge(metrics.PrometheusExporterRefreshInflight, float64(e.refreshInflight.Load()), nil)

	// Phase 1: Collect - snapshot metrics
	collectStart := time.Now()
	e.mu.RLock()
	snapshot := make([]telemetry.MetricsEvent, len(e.metrics))
	copy(snapshot, e.metrics)
	e.mu.RUnlock()
	collectDuration := time.Since(collectStart)
	telemetry.EmitHistogram(metrics.PrometheusExporterRefreshDurationSeconds, collectDuration, map[string]string{metrics.TagPhase: metrics.PhaseCollect})

	// Phase 2: Convert - group and prepare Prometheus format
	convertStart := time.Now()
	metricGroups := make(map[string][]telemetry.MetricsEvent)
	for _, metric := range snapshot {
		key := fmt.Sprintf("%s_%s", metric.Name, e.getMetricType(metric))
		metricGroups[key] = append(metricGroups[key], metric)
	}
	convertDuration := time.Since(convertStart)
	telemetry.EmitHistogram(metrics.PrometheusExporterRefreshDurationSeconds, convertDuration, map[string]string{metrics.TagPhase: metrics.PhaseConvert})

	// Phase 3: Export - write to HTTP response
	exportStart := time.Now()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Write metrics in Prometheus format
	for _, metricsGroup := range metricGroups {
		if len(metricsGroup) == 0 {
			continue
		}

		// Get the first metric to determine type
		firstMetric := metricsGroup[0]

		switch firstMetric.Type {
		case telemetry.TypeCounter:
			e.writeCounterMetrics(w, metricsGroup)
		case telemetry.TypeGauge:
			e.writeGaugeMetrics(w, metricsGroup)
		case telemetry.TypeHistogram:
			e.writeHistogramMetrics(w, metricsGroup)
		}
	}
	exportDuration := time.Since(exportStart)
	telemetry.EmitHistogram(metrics.PrometheusExporterRefreshDurationSeconds, exportDuration, map[string]string{metrics.TagPhase: metrics.PhaseExport})

	// Emit overall refresh metrics
	overallDuration := time.Since(overallStart)
	telemetry.EmitHistogram(metrics.PrometheusExporterRefreshDurationSeconds, overallDuration, map[string]string{metrics.TagPhase: "overall"})
	telemetry.EmitCounter(metrics.PrometheusExporterRefreshTotal, 1, map[string]string{metrics.TagResult: metrics.ResultSuccess})
}

// formatPrometheusName converts metric name to Prometheus format
func (e *PrometheusExporter) formatPrometheusName(name string) string {
	// Add prefix if specified
	if e.config.Prefix != "" {
		name = e.config.Prefix + "_" + name
	}

	// Convert to Prometheus naming convention (snake_case)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return strings.ToLower(name)
}

// formatPrometheusLabels converts tags to Prometheus label format
// Labels are sorted alphabetically by key for deterministic output.
func (e *PrometheusExporter) formatPrometheusLabels(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	// Sort keys for deterministic output (Go map iteration order is randomized)
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	labels := make([]string, 0, len(tags))
	for _, key := range keys {
		value := tags[key]
		// Escape quotes in label values
		escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
		labels = append(labels, fmt.Sprintf(`%s="%s"`, key, escapedValue))
	}

	return strings.Join(labels, ",")
}

// formatPrometheusLabelsWithAdditional converts tags to Prometheus label format with additional label
// Labels are sorted alphabetically by key for deterministic output.
func (e *PrometheusExporter) formatPrometheusLabelsWithAdditional(tags map[string]string, additionalKey, additionalValue string) string {
	// Create a copy of tags and add the additional label
	allTags := make(map[string]string, len(tags)+1)
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
		name := e.formatPrometheusName(metric.Name)
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
		name := e.formatPrometheusName(metric.Name)
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

// writeHistogramMetrics writes histogram metrics in Prometheus format with bucket conversion
func (e *PrometheusExporter) writeHistogramMetrics(w io.Writer, metrics []telemetry.MetricsEvent) {
	for _, metric := range metrics {
		switch v := metric.Value.(type) {
		case telemetry.HistogramSummary:
			// Prometheus expects seconds for duration metrics, but ADR-0007 uses milliseconds
			// Convert if metric ends with _ms or _seconds
			convertToSeconds := strings.HasSuffix(metric.Name, "_ms") || strings.HasSuffix(metric.Name, "_seconds")

			// Write bucket series
			for _, bucket := range v.Buckets {
				bucketLE := bucket.LE
				if convertToSeconds {
					// Convert milliseconds to seconds
					bucketLE = bucketLE / 1000.0
				}
				bucketLabels := e.formatPrometheusLabelsWithAdditional(metric.Tags, "le", fmt.Sprintf("%g", bucketLE))
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

			// Write sum and count (also convert sum if needed)
			sum := v.Sum
			if convertToSeconds {
				sum = sum / 1000.0
			}

			sumName := e.formatPrometheusName(metric.Name + "_sum")
			countName := e.formatPrometheusName(metric.Name + "_count")
			labels := e.formatPrometheusLabels(metric.Tags)

			if labels != "" {
				if _, err := fmt.Fprintf(w, "%s{%s} %f\n", sumName, labels, sum); err != nil {
					fmt.Printf("Error writing histogram sum: %v\n", err)
					return
				}
				if _, err := fmt.Fprintf(w, "%s{%s} %d\n", countName, labels, v.Count); err != nil {
					fmt.Printf("Error writing histogram count: %v\n", err)
					return
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s %f\n", sumName, sum); err != nil {
					fmt.Printf("Error writing histogram sum: %v\n", err)
					return
				}
				if _, err := fmt.Fprintf(w, "%s %d\n", countName, v.Count); err != nil {
					fmt.Printf("Error writing histogram count: %v\n", err)
					return
				}
			}
		case float64:
			// Single histogram value - convert if needed
			value := v
			if strings.HasSuffix(metric.Name, "_ms") || strings.HasSuffix(metric.Name, "_seconds") {
				value = value / 1000.0
			}

			name := e.formatPrometheusName(metric.Name)
			labels := e.formatPrometheusLabels(metric.Tags)
			if labels != "" {
				if _, err := fmt.Fprintf(w, "%s{%s} %f\n", name, labels, value); err != nil {
					fmt.Printf("Error writing histogram value: %v\n", err)
					return
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s %f\n", name, value); err != nil {
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
