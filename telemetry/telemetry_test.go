package telemetry

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "explicit default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "disabled system",
			config: &Config{
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys, err := NewSystem(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sys)
			}
		})
	}
}

func TestCounter(t *testing.T) {
	sys, err := NewSystem(nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		value   float64
		tags    map[string]string
		wantErr bool
	}{
		{
			name:    "simple counter",
			value:   1.0,
			tags:    nil,
			wantErr: false,
		},
		{
			name:    "counter with tags",
			value:   42.0,
			tags:    map[string]string{"component": "test", "operation": "process"},
			wantErr: false,
		},
		{
			name:    "zero counter",
			value:   0.0,
			tags:    nil,
			wantErr: false,
		},
		{
			name:    "negative counter",
			value:   -1.0,
			tags:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.Counter("test_counter", tt.value, tt.tags)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHistogram(t *testing.T) {
	sys, err := NewSystem(nil)
	require.NoError(t, err)

	tests := []struct {
		name     string
		duration time.Duration
		tags     map[string]string
		wantErr  bool
	}{
		{
			name:     "1ms duration",
			duration: 1 * time.Millisecond,
			tags:     nil,
			wantErr:  false,
		},
		{
			name:     "100ms duration with tags",
			duration: 100 * time.Millisecond,
			tags:     map[string]string{"operation": "validation"},
			wantErr:  false,
		},
		{
			name:     "1 second duration",
			duration: 1 * time.Second,
			tags:     nil,
			wantErr:  false,
		},
		{
			name:     "microsecond duration",
			duration: 50 * time.Microsecond,
			tags:     nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.Histogram("test_histogram", tt.duration, tt.tags)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHistogramSummary(t *testing.T) {
	sys, err := NewSystem(nil)
	require.NoError(t, err)

	summary := HistogramSummary{
		Count: 100,
		Sum:   5000.0,
		Buckets: []HistogramBucket{
			{LE: 1.0, Count: 10},
			{LE: 5.0, Count: 50},
			{LE: 10.0, Count: 90},
			{LE: 50.0, Count: 100},
		},
	}

	err = sys.HistogramSummary("test_histogram_summary", summary, map[string]string{"test": "true"})
	assert.NoError(t, err)
}

func TestDisabledSystem(t *testing.T) {
	sys, err := NewSystem(&Config{Enabled: false})
	require.NoError(t, err)

	// All operations should be no-ops
	err = sys.Counter("test", 1.0, nil)
	assert.NoError(t, err)

	err = sys.Histogram("test", 1*time.Millisecond, nil)
	assert.NoError(t, err)

	err = sys.HistogramSummary("test", HistogramSummary{}, nil)
	assert.NoError(t, err)
}

func TestMetricsEventValidation(t *testing.T) {
	// Try to load the metrics schema, but skip if it fails due to reference resolution
	catalog := schema.DefaultCatalog()
	validator, err := catalog.ValidatorByID("observability/metrics/v1.0.0/metrics-event")
	if err != nil {
		t.Skipf("Skipping schema validation test due to schema loading issues: %v", err)
	}

	tests := []struct {
		name    string
		event   MetricsEvent
		wantErr bool
	}{
		{
			name: "valid counter event",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Name:      "schema_validations",
				Value:     1.0,
				Tags:      map[string]string{"component": "test"},
				Unit:      "count",
			},
			wantErr: false,
		},
		{
			name: "valid histogram event",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Name:      "config_load_ms",
				Value:     100.0,
				Tags:      map[string]string{"operation": "load"},
				Unit:      "ms",
			},
			wantErr: false,
		},
		{
			name: "valid histogram summary event",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Name:      "config_load_ms",
				Value: HistogramSummary{
					Count: 10,
					Sum:   500.0,
					Buckets: []HistogramBucket{
						{LE: 1.0, Count: 1},
						{LE: 10.0, Count: 8},
						{LE: 100.0, Count: 10},
					},
				},
				Unit: "ms",
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp",
			event: MetricsEvent{
				Timestamp: "invalid-timestamp",
				Name:      "schema_validations",
				Value:     1.0,
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				// Missing name and value
			},
			wantErr: true,
		},
		{
			name: "invalid metric name",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Name:      "invalid_metric_name",
				Value:     1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			event: MetricsEvent{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Name:      "schema_validations",
				Value:     1.0,
				Unit:      "invalid_unit",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics, err := validator.ValidateData(tt.event)
			if tt.wantErr {
				assert.True(t, err != nil || len(diagnostics) > 0, "Expected validation to fail")
			} else {
				assert.NoError(t, err)
				assert.Empty(t, diagnostics)
			}
		})
	}
}

func TestMetricsEventJSONSerialization(t *testing.T) {
	event := MetricsEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      "test_metric",
		Value:     42.0,
		Tags:      map[string]string{"tag1": "value1"},
		Unit:      "count",
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled MetricsEvent
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, event.Timestamp, unmarshaled.Timestamp)
	assert.Equal(t, event.Name, unmarshaled.Name)
	assert.Equal(t, event.Value, unmarshaled.Value)
	assert.Equal(t, event.Tags, unmarshaled.Tags)
	assert.Equal(t, event.Unit, unmarshaled.Unit)
}
