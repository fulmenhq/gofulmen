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

### Prometheus Exporter

The Prometheus exporter provides production-grade HTTP metrics exposition with enterprise features including authentication, rate limiting, and comprehensive health instrumentation.

#### Basic Usage

```go
import "github.com/fulmenhq/gofulmen/telemetry/exporters"

// Create and start a Prometheus exporter (simple)
exporter := exporters.NewPrometheusExporter("myapp", ":9090")
if err := exporter.Start(); err != nil {
    log.Fatal(err)
}
defer exporter.Stop()

// Configure telemetry to use the exporter
config := &telemetry.Config{
    Enabled: true,
    Emitter: exporter,
}
sys, err := telemetry.NewSystem(config)

// Metrics available at http://localhost:9090/metrics
```

#### Advanced Configuration

```go
import "github.com/fulmenhq/gofulmen/telemetry/exporters"

// Create exporter with full configuration
config := &exporters.PrometheusConfig{
    Prefix:             "myapp",
    Endpoint:           ":9090",
    BearerToken:        "secret-token-here",  // Enable authentication
    RateLimitPerMinute: 60,                   // 60 requests/minute
    RateLimitBurst:     10,                   // Allow bursts of 10
    RefreshInterval:    15 * time.Second,     // Auto-refresh every 15s
    QuietMode:          false,                // Log HTTP requests
    ReadHeaderTimeout:  10 * time.Second,     // Prevent Slowloris attacks
}

exporter := exporters.NewPrometheusExporterWithConfig(config)
if err := exporter.Start(); err != nil {
    log.Fatal(err)
}

// Access with authentication
// curl -H "Authorization: Bearer secret-token-here" http://localhost:9090/metrics
```

#### Output Format

The exporter converts telemetry metrics to Prometheus format:

**Counters** → `<prefix>_<name>_total{tags}`

```
myapp_requests_total{status="200",method="GET"} 42
```

**Gauges** → `<prefix>_<name>_gauge{tags}`

```
myapp_cpu_usage_percent_gauge{host="server1"} 75.5
```

**Histograms** → Prometheus histogram with buckets

```
myapp_request_duration_seconds_bucket{endpoint="/api",le="0.05"} 10
myapp_request_duration_seconds_bucket{endpoint="/api",le="0.1"} 25
myapp_request_duration_seconds_sum{endpoint="/api"} 2.5
myapp_request_duration_seconds_count{endpoint="/api"} 30
```

**Note**: Histogram durations are automatically converted from milliseconds to seconds for Prometheus compatibility.

#### Health Instrumentation

The exporter includes 7 built-in health metrics:

| Metric                                         | Type      | Description                                 |
| ---------------------------------------------- | --------- | ------------------------------------------- |
| `prometheus_exporter_refresh_duration_seconds` | Histogram | Refresh pipeline execution time             |
| `prometheus_exporter_refresh_total`            | Counter   | Total refresh operations (by phase/result)  |
| `prometheus_exporter_refresh_errors_total`     | Counter   | Refresh failures (by phase/reason)          |
| `prometheus_exporter_refresh_inflight`         | Gauge     | Currently running refresh operations        |
| `prometheus_exporter_http_requests_total`      | Counter   | HTTP endpoint requests (by endpoint/status) |
| `prometheus_exporter_http_errors_total`        | Counter   | HTTP errors (by endpoint/status)            |
| `prometheus_exporter_restarts_total`           | Counter   | Exporter restarts (by reason)               |

#### HTTP Endpoints

- **`GET /metrics`** - Prometheus metrics exposition (text/plain format)
- **`GET /`** - HTML landing page with exporter status and links
- **`GET /health`** - JSON health status endpoint

#### Rate Limiting

Rate limiting applies per-client IP when configured:

```go
config := &exporters.PrometheusConfig{
    RateLimitPerMinute: 60, // 60 requests per minute per IP
    RateLimitBurst:     10, // Allow bursts of 10 requests
}
```

Set `RateLimitPerMinute: 0` to disable rate limiting.

Requests exceeding the limit receive `429 Too Many Requests` with headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1699564800
```

#### Authentication

Bearer token authentication protects metrics endpoints:

```go
config := &exporters.PrometheusConfig{
    BearerToken: "your-secure-token",
}
```

Leave `BearerToken` empty to disable authentication.

Clients must include the token:

```bash
curl -H "Authorization: Bearer your-secure-token" http://localhost:9090/metrics
```

#### Examples

See comprehensive examples in:

- `examples/phase5-telemetry-demo.go` - Full Prometheus exporter demo
- `telemetry/exporters/prometheus_test.go` - Integration tests
- `cmd/phase5-demo/main.go` - Quick demo runner

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
