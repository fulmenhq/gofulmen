// Package telemetry provides structured metrics emission helpers for Fulmen libraries.
// It supports counters and histograms using the canonical taxonomy defined in
// config/crucible-go/taxonomy/metrics.yaml and validates emitted metrics against
// schemas/observability/metrics/v1.0.0/metrics-event.schema.json
package telemetry

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fulmenhq/gofulmen/schema"
)

// MetricType represents the type of metric being emitted
type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeHistogram MetricType = "histogram"
	TypeGauge     MetricType = "gauge"
)

// DefaultHistogramBucketsMS contains the default bucket boundaries for millisecond metrics
// per ADR-0007: [1, 5, 10, 50, 100, 500, 1000, 5000, 10000]
var DefaultHistogramBucketsMS = []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000}

// MetricsEmitter defines the interface for emitting structured metrics
type MetricsEmitter interface {
	// Counter emits a counter metric increment
	Counter(name string, value float64, tags map[string]string) error

	// Histogram emits a histogram metric with timing data
	Histogram(name string, duration time.Duration, tags map[string]string) error

	// HistogramSummary emits a pre-calculated histogram summary
	HistogramSummary(name string, summary HistogramSummary, tags map[string]string) error

	// Gauge emits a gauge metric with current value
	Gauge(name string, value float64, tags map[string]string) error
}

// HistogramSummary represents a pre-calculated histogram summary
type HistogramSummary struct {
	Count   int64             `json:"count"`
	Sum     float64           `json:"sum"`
	Buckets []HistogramBucket `json:"buckets"`
}

// HistogramBucket represents a single bucket in a histogram
type HistogramBucket struct {
	LE    float64 `json:"le"`    // Less than or equal boundary
	Count int64   `json:"count"` // Cumulative count up to this bucket
}

// MetricsEvent represents a structured metric event matching the schema
type MetricsEvent struct {
	Timestamp string            `json:"timestamp"`
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Unit      string            `json:"unit,omitempty"`
}

// calculateHistogramBuckets calculates histogram buckets for a given duration using ADR-0007 defaults
func calculateHistogramBuckets(duration time.Duration, buckets []float64) []HistogramBucket {
	if len(buckets) == 0 {
		buckets = DefaultHistogramBucketsMS
	}

	durationMs := float64(duration.Milliseconds())
	result := make([]HistogramBucket, len(buckets)+1) // +1 for +Inf bucket

	for i, boundary := range buckets {
		count := int64(0)
		if durationMs <= boundary {
			count = 1
		}
		result[i] = HistogramBucket{
			LE:    boundary,
			Count: count,
		}
	}

	// Add +Inf bucket to ensure all samples are counted
	result[len(buckets)] = HistogramBucket{
		LE:    math.Inf(1), // +Inf
		Count: 1,           // All samples should be <= +Inf
	}

	return result
}

// Config holds configuration for the telemetry system
type Config struct {
	Enabled       bool              `json:"enabled"`
	Emitter       MetricsEmitter    `json:"-"`
	Schema        *schema.Validator `json:"-"`
	BatchSize     int               `json:"batchSize,omitempty"`     // Maximum number of metrics in a batch (0 = no batching)
	BatchInterval time.Duration     `json:"batchInterval,omitempty"` // Maximum time to wait before emitting a batch (0 = immediate)
}

// DefaultConfig returns a default telemetry configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		BatchSize:     0, // No batching by default (immediate emission)
		BatchInterval: 0, // Immediate emission
	}
}

// System manages telemetry operations
type System struct {
	config *Config
	mu     sync.RWMutex

	// Batching support
	metricBuffer  []MetricsEvent
	lastFlushTime time.Time
	flushTimer    *time.Timer

	// Internal counters for tracking telemetry health
	validationErrors int64
	emissionErrors   int64
}

// NewSystem creates a new telemetry system
func NewSystem(config *Config) (*System, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Load metrics schema if not provided and enabled
	if config.Schema == nil && config.Enabled {
		catalog := schema.DefaultCatalog()
		validator, err := catalog.ValidatorByID("observability/metrics/v1.0.0/metrics-event")
		if err != nil {
			// Schema loading can fail due to reference resolution issues
			// We'll continue without schema validation in this case
			// This allows the system to work even if schema references aren't perfectly resolved
			config.Schema = nil
		} else {
			config.Schema = validator
		}
	}

	return &System{
		config: config,
	}, nil
}

// Counter emits a counter metric increment
func (s *System) Counter(name string, value float64, tags map[string]string) error {
	if !s.isEnabled() {
		return nil
	}

	event := MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      TypeCounter,
		Value:     value,
		Tags:      tags,
	}

	return s.emit(event)
}

// Gauge emits a gauge metric with current value
func (s *System) Gauge(name string, value float64, tags map[string]string) error {
	if !s.isEnabled() {
		return nil
	}

	event := MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      TypeGauge,
		Value:     value,
		Tags:      tags,
	}

	return s.emit(event)
}

// Histogram emits a histogram metric with timing data
// Automatically uses ADR-0007 default buckets for metrics ending with "_ms"
func (s *System) Histogram(name string, duration time.Duration, tags map[string]string) error {
	if !s.isEnabled() {
		return nil
	}

	// Check if this is a millisecond metric that should use ADR-0007 buckets
	if strings.HasSuffix(name, "_ms") {
		// Generate histogram summary with ADR-0007 default buckets
		summary := HistogramSummary{
			Count:   1,
			Sum:     float64(duration.Milliseconds()),
			Buckets: calculateHistogramBuckets(duration, DefaultHistogramBucketsMS),
		}
		return s.HistogramSummary(name, summary, tags)
	}

	// For non-ms metrics, emit as single value (backward compatibility)
	ms := float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds
	event := MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      TypeHistogram,
		Value:     ms,
		Tags:      tags,
		Unit:      "ms",
	}

	return s.emit(event)
}

// HistogramSummary emits a pre-calculated histogram summary
func (s *System) HistogramSummary(name string, summary HistogramSummary, tags map[string]string) error {
	if !s.isEnabled() {
		return nil
	}

	event := MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Type:      TypeHistogram,
		Value:     summary,
		Tags:      tags,
		Unit:      "ms",
	}

	return s.emit(event)
}

// emit handles the actual emission and validation
func (s *System) emit(event MetricsEvent) error {
	// Check if batching is enabled
	if s.config.BatchSize > 0 || s.config.BatchInterval > 0 {
		return s.bufferMetric(event)
	}

	// Immediate emission (no batching)
	return s.emitImmediate(event)
}

// bufferMetric adds a metric to the buffer and handles batching logic
func (s *System) bufferMetric(event MetricsEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metricBuffer = append(s.metricBuffer, event)

	// Check if we should flush based on batch size
	if s.config.BatchSize > 0 && len(s.metricBuffer) >= s.config.BatchSize {
		return s.flushBufferLocked()
	}

	// Check if we should flush based on time interval
	if s.config.BatchInterval > 0 && time.Since(s.lastFlushTime) >= s.config.BatchInterval {
		return s.flushBufferLocked()
	}

	// Schedule a flush timer if not already scheduled
	if s.config.BatchInterval > 0 && s.flushTimer == nil {
		s.flushTimer = time.AfterFunc(s.config.BatchInterval, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if len(s.metricBuffer) > 0 {
				if err := s.flushBufferLocked(); err != nil {
					// Log error but don't fail the timer - telemetry should be resilient
					fmt.Printf("telemetry: failed to flush buffer: %v\n", err)
				}
			}
		})
	}

	return nil
}

// flushBufferLocked flushes the current buffer (must be called with lock held)
func (s *System) flushBufferLocked() error {
	if len(s.metricBuffer) == 0 {
		return nil
	}

	// Cancel any pending timer
	if s.flushTimer != nil {
		s.flushTimer.Stop()
		s.flushTimer = nil
	}

	// Emit all buffered metrics
	for _, event := range s.metricBuffer {
		if err := s.emitImmediate(event); err != nil {
			return err
		}
	}

	// Clear buffer and update flush time
	s.metricBuffer = s.metricBuffer[:0]
	s.lastFlushTime = time.Now()

	return nil
}

// Flush manually flushes any buffered metrics
func (s *System) Flush() error {
	if !s.isEnabled() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.flushBufferLocked()
}

// emitImmediate handles immediate emission without batching
func (s *System) emitImmediate(event MetricsEvent) error {
	// Validate against schema if available
	if s.config.Schema != nil {
		// Convert struct to map for schema validation
		eventJSON, err := json.Marshal(event)
		if err != nil {
			s.incrementValidationErrors()
			return fmt.Errorf("failed to marshal event for validation: %w", err)
		}

		var eventMap map[string]interface{}
		if err := json.Unmarshal(eventJSON, &eventMap); err != nil {
			s.incrementValidationErrors()
			return fmt.Errorf("failed to unmarshal event for validation: %w", err)
		}

		diagnostics, err := s.config.Schema.ValidateData(eventMap)
		if err != nil {
			s.incrementValidationErrors()
			return fmt.Errorf("schema validation failed: %w", err)
		}
		if len(diagnostics) > 0 {
			s.incrementValidationErrors()
			return fmt.Errorf("schema validation failed: %v", diagnostics)
		}
	}

	// Use custom emitter if provided
	if s.config.Emitter != nil {
		switch event.Type {
		case TypeCounter:
			if v, ok := event.Value.(float64); ok {
				return s.config.Emitter.Counter(event.Name, v, event.Tags)
			}
			return fmt.Errorf("counter metric value must be float64, got %T", event.Value)
		case TypeGauge:
			if v, ok := event.Value.(float64); ok {
				return s.config.Emitter.Gauge(event.Name, v, event.Tags)
			}
			return fmt.Errorf("gauge metric value must be float64, got %T", event.Value)
		case TypeHistogram:
			switch v := event.Value.(type) {
			case float64:
				// Single histogram value - convert back to duration
				return s.config.Emitter.Histogram(event.Name, time.Duration(v*1e6)*time.Nanosecond, event.Tags)
			case HistogramSummary:
				return s.config.Emitter.HistogramSummary(event.Name, v, event.Tags)
			default:
				return fmt.Errorf("histogram metric value must be float64 or HistogramSummary, got %T", v)
			}
		default:
			return fmt.Errorf("unsupported metric type: %s", event.Type)
		}
	}

	// Default JSON emission
	jsonData, err := json.Marshal(event)
	if err != nil {
		s.incrementEmissionErrors()
		return fmt.Errorf("failed to marshal metric event: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// isEnabled checks if telemetry is enabled
func (s *System) isEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Enabled
}

// incrementValidationErrors increments the validation error counter
func (s *System) incrementValidationErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validationErrors++
}

// incrementEmissionErrors increments the emission error counter
// Currently unused but kept for future telemetry health monitoring
//
//nolint:unused // This function is part of the telemetry API for future use
func (s *System) incrementEmissionErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emissionErrors++
}

// Stats returns telemetry system statistics
func (s *System) Stats() (emissionErrors int64, validationErrors int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.emissionErrors, s.validationErrors
}

// MarshalJSON implements json.Marshaler for MetricsEvent
func (e MetricsEvent) MarshalJSON() ([]byte, error) {
	// Create a custom type to avoid infinite recursion
	type Alias MetricsEvent

	// Handle special cases for histogram summaries with +Inf values
	if e.Type == TypeHistogram {
		if summary, ok := e.Value.(HistogramSummary); ok {
			// Create a copy of the summary with +Inf values handled
			newSummary := HistogramSummary{
				Count:   summary.Count,
				Sum:     summary.Sum,
				Buckets: make([]HistogramBucket, len(summary.Buckets)),
			}
			copy(newSummary.Buckets, summary.Buckets)

			// Replace +Inf with a string representation for JSON
			for i := range newSummary.Buckets {
				if math.IsInf(newSummary.Buckets[i].LE, 1) {
					// For JSON serialization, we'll use a very large number instead of +Inf
					newSummary.Buckets[i].LE = 1e308 // Close to max float64
				}
			}

			// Create a new event with the modified summary
			newEvent := e
			newEvent.Value = newSummary
			return json.Marshal((*Alias)(&newEvent))
		}
	}

	return json.Marshal((*Alias)(&e))
}

// Global telemetry system for module instrumentation
var (
	globalSystem     *System
	globalSystemOnce sync.Once
	globalSystemMu   sync.RWMutex
)

// SetGlobalSystem sets the global telemetry system for module instrumentation.
// This should be called once during application initialization.
// If never called, modules will use a default no-op system.
func SetGlobalSystem(system *System) {
	globalSystemMu.Lock()
	defer globalSystemMu.Unlock()
	globalSystem = system
}

// GetGlobalSystem returns the global telemetry system.
// If no system has been set, it returns a disabled no-op system.
func GetGlobalSystem() *System {
	globalSystemOnce.Do(func() {
		globalSystemMu.RLock()
		if globalSystem == nil {
			globalSystemMu.RUnlock()
			// Create a disabled no-op system
			config := DefaultConfig()
			config.Enabled = false
			sys, _ := NewSystem(config)
			globalSystemMu.Lock()
			globalSystem = sys
			globalSystemMu.Unlock()
		} else {
			globalSystemMu.RUnlock()
		}
	})
	globalSystemMu.RLock()
	defer globalSystemMu.RUnlock()
	return globalSystem
}

// EmitCounter is a convenience function for modules to emit counter metrics.
// It uses the global telemetry system and gracefully handles nil system.
func EmitCounter(name string, value float64, tags map[string]string) {
	system := GetGlobalSystem()
	if system != nil {
		_ = system.Counter(name, value, tags)
	}
}

// EmitHistogram is a convenience function for modules to emit histogram metrics.
// It uses the global telemetry system and gracefully handles nil system.
func EmitHistogram(name string, duration time.Duration, tags map[string]string) {
	system := GetGlobalSystem()
	if system != nil {
		_ = system.Histogram(name, duration, tags)
	}
}

// EmitGauge is a convenience function for modules to emit gauge metrics.
// It uses the global telemetry system and gracefully handles nil system.
func EmitGauge(name string, value float64, tags map[string]string) {
	system := GetGlobalSystem()
	if system != nil {
		_ = system.Gauge(name, value, tags)
	}
}
