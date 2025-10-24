# Telemetry Package

The `telemetry` package provides structured metrics emission helpers for Fulmen libraries. It supports counters, gauges, and histograms using the canonical taxonomy defined in `config/crucible-go/taxonomy/metrics.yaml` and validates emitted metrics against `schemas/observability/metrics/v1.0.0/metrics-event.schema.json`.

## Features

- **Counter Metrics**: Simple incrementing counters for event counting
- **Gauge Metrics**: Real-time value metrics for system monitoring (CPU %, memory usage, temperature)
- **Histogram Metrics**: Timing and distribution metrics with automatic millisecond conversion
- **Custom Exporters**: Pluggable emitter interface with Prometheus exporter included
- **Schema Validation**: Automatic validation against the official metrics schema
- **Configurable**: Can be enabled/disabled and supports custom emitters
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Enterprise Ready**: Production-grade with <5% performance overhead

## Usage

### Basic Usage

```go
import "github.com/fulmenhq/gofulmen/telemetry"

// Create a telemetry system with default configuration
sys, err := telemetry.NewSystem(nil)
if err != nil {
    return err
}

// Emit a counter metric
err = sys.Counter("schema_validations", 1.0, map[string]string{
    "component": "validator",
    "status":    "success",
})

// Emit a gauge metric (real-time values like CPU usage, memory, temperature)
err = sys.Gauge("cpu_usage_percent", 75.5, map[string]string{
    "host": "server1",
    "cpu":  "cpu0",
})

// Emit a histogram metric (duration is automatically converted to milliseconds)
start := time.Now()
// ... do some work ...
err = sys.Histogram("config_load_ms", time.Since(start), map[string]string{
    "operation": "load",
})
```

### Prometheus Exporter (NEW)

```go
import "github.com/fulmenhq/gofulmen/telemetry/exporters"

// Create and start a Prometheus exporter
exporter := exporters.NewPrometheusExporter("myapp", ":9090")
if err := exporter.Start(); err != nil {
    log.Fatal(err)
}

// Configure telemetry to use the exporter
config := &telemetry.Config{
    Enabled: true,
    Emitter: exporter,
}
sys, err := telemetry.NewSystem(config)

// Metrics will be available at http://localhost:9090/metrics
// Example output:
// myapp_requests_total{status="200"} 42
// myapp_cpu_usage_percent_gauge{host="server1"} 75.5
// myapp_request_duration_ms_bucket{endpoint="/api",le="50"} 10
// myapp_request_duration_ms_sum{endpoint="/api"} 500
// myapp_request_duration_ms_count{endpoint="/api"} 10
```

### Advanced Usage with Custom Emitter

```go
type MyEmitter struct{}

func (m *MyEmitter) Counter(name string, value float64, tags map[string]string) error {
    // Custom counter implementation
    return nil
}

func (m *MyEmitter) Gauge(name string, value float64, tags map[string]string) error {
    // Custom gauge implementation
    return nil
}

func (m *MyEmitter) Histogram(name string, duration time.Duration, tags map[string]string) error {
    // Custom histogram implementation
    return nil
}

func (m *MyEmitter) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
    // Custom histogram summary implementation
    return nil
}

config := &telemetry.Config{
    Enabled: true,
    Emitter: &MyEmitter{},
}

sys, err := telemetry.NewSystem(config)
```

### Disabled Telemetry

```go
// Create a disabled telemetry system (no-op)
sys, err := telemetry.NewSystem(&telemetry.Config{Enabled: false})
// All operations will return nil without doing any work
```

## Metric Types

### Counter Metrics

- Used for counting events
- Accept any float64 value (including negative values for decrements)
- Automatically tagged with component and operation information
- In Prometheus: exported as `<name>_total` with proper counter semantics

### Gauge Metrics

- Used for real-time value measurements (CPU usage, memory, temperature)
- Can go up and down over time
- Perfect for system monitoring and resource utilization
- In Prometheus: exported as `<name>_gauge` with current value

### Histogram Metrics

- Used for timing and distribution measurements
- Duration is automatically converted to milliseconds
- Supports both single values and pre-calculated histogram summaries
- Default bucket boundaries are defined in the taxonomy configuration
- In Prometheus: exported as full bucket series with `_bucket`, `_sum`, and `_count`

## Schema Validation

The telemetry system automatically validates all emitted metrics against the official Fulmen metrics schema. This ensures:

- Metric names are from the approved taxonomy
- Units are valid (count, ms, bytes, percent)
- Value types are correct
- Required fields are present

## Error Handling

The system provides detailed error messages for:

- Schema validation failures
- Invalid metric names or units
- Malformed metric values
- Custom emitter errors

## Performance

The telemetry system is designed for minimal overhead:

- Lock-free operations when disabled
- Efficient schema validation
- Minimal memory allocations
- Thread-safe concurrent operations

## Integration with Logging

The telemetry package is designed to integrate seamlessly with the Fulmen logging system. Metrics can be emitted as structured log events or forwarded to external monitoring systems through custom emitters.
