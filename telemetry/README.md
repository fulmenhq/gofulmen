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
- **Enterprise Ready**: Production-grade with ~35% HTTP middleware overhead (optimized from 55-84%)

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

## HTTP Server Metrics

The telemetry package includes comprehensive HTTP server metrics middleware that supports all five HTTP metrics from the Crucible v0.2.18 taxonomy with proper bucket mathematics and minimal label cardinality.

### Features

- **Complete HTTP Metrics**: All five metrics (requests, duration, request size, response size, active requests)
- **Proper Bucket Mathematics**: Mathematically correct cumulative histogram construction
- **Cardinality Control**: Minimal label sets to prevent metric explosion
- **Route Normalization**: Template-based routes (/users/{id} vs /users/123)
- **Framework Integration**: Works with net/http, Chi, Gin, and other frameworks
- **Performance Optimized**: ~26-35% per-request overhead with tag pooling and optimized allocations

### Basic Usage

```go
import "github.com/fulmenhq/gofulmen/telemetry"

// Create middleware with default configuration
middleware := telemetry.HTTPMetricsMiddleware(
    emitter, // Your MetricsEmitter implementation
    telemetry.WithServiceName("my-api"),
)

// Apply to net/http handler
handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello, World!"))
}))
```

### Configuration Options

```go
// Custom route normalization
middleware := telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("my-api"),
    telemetry.WithRouteNormalizer(func(method, path string) string {
        // Custom normalization logic
        if strings.Contains(path, ":id") {
            return strings.Replace(path, ":id", "{id}", 1)
        }
        return telemetry.DefaultRouteNormalizer(method, path)
    }),
)

// Custom size bucket configuration (duration buckets are emitter-driven)
middleware := telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("my-api"),
    telemetry.WithCustomSizeBuckets(
        []float64{1024, 10240, 102400, 1048576, 10485760}, // size buckets (bytes)
    ),
)
```

### Framework Integration

#### Chi Router

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()

// Apply HTTP metrics middleware
r.Use(telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("chi-api"),
))

r.Get("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"id": "` + userID + `"}`))
})
```

#### Gin Router

```go
import "github.com/gin-gonic/gin"

r := gin.Default()

// Apply HTTP metrics middleware
r.Use(telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("gin-api"),
    telemetry.WithRouteNormalizer(func(method, path string) string {
        // Convert Gin's :param to {param} for consistency
        if strings.Contains(path, ":id") {
            return strings.Replace(path, ":id", "{id}", 1)
        }
        return telemetry.DefaultRouteNormalizer(method, path)
    }),
))

r.GET("/api/users/:id", func(c *gin.Context) {
    userID := c.Param("id")
    c.JSON(http.StatusOK, gin.H{"id": userID})
})
```

### Emitted Metrics

The middleware emits the following HTTP metrics:

| Metric                          | Type      | Labels                         | Description                    |
| ------------------------------- | --------- | ------------------------------ | ------------------------------ |
| `http_requests_total`           | Counter   | method, route, status, service | Total HTTP requests            |
| `http_request_duration_seconds` | Histogram | method, route, status, service | Request duration in seconds    |
| `http_request_size_bytes`       | Histogram | method, route, status, service | Request body size in bytes     |
| `http_response_size_bytes`      | Histogram | method, route, status, service | Response body size in bytes    |
| `http_active_requests`          | Gauge     | service                        | Currently active HTTP requests |

### Route Normalization

Route normalization prevents cardinality explosion by converting dynamic paths to templates:

- `/users/123` → `/users/{id}`
- `/api/v1/users/550e8400-e29b-41d4-a716-446655440000` → `/api/v1/users/{uuid}`
- `/static/css/main.css` → `/static/css/main.css` (unchanged)

### Performance

Benchmark results on Apple M2 Max:

```
BenchmarkHTTPMetricsMiddleware-12                 660091    1823 ns/op   6388 B/op   27 allocs/op
BenchmarkHTTPMetricsMiddlewareWithoutMetrics-12  1000000    1504 ns/op   5317 B/op   14 allocs/op
BenchmarkHTTPMetricsMiddlewareConcurrent-12       575192    2183 ns/op   6388 B/op   27 allocs/op
```

Overhead: ~84% for comprehensive metrics collection (includes histogram construction and tag creation).

### Migration Guide

To add HTTP metrics to existing servers:

1. **Import the package**

   ```go
   import "github.com/fulmenhq/gofulmen/telemetry"
   ```

2. **Create the middleware**

   ```go
   middleware := telemetry.HTTPMetricsMiddleware(
       yourEmitter,
       telemetry.WithServiceName("your-service"),
   )
   ```

3. **Apply to your router**

   ```go
   // For net/http
   handler := middleware(yourHandler)

   // For Chi
   r.Use(middleware)

   // For Gin
   r.Use(middleware)
   ```

4. **Configure custom size buckets** (optional)
   ```go
   telemetry.WithCustomSizeBuckets(
       []float64{1024, 10240, 102400, 1048576, 10485760},
   )
   ```

### Testing

See comprehensive tests in:

- `telemetry/http_metrics_external_test.go` - Basic functionality tests
- `telemetry/http_metrics_chi_test.go` - Chi integration patterns
- `telemetry/http_metrics_gin_test.go` - Gin integration patterns
- `telemetry/http_metrics_bench_test.go` - Performance benchmarks

## Integration with Logging

The telemetry package is designed to integrate seamlessly with the Fulmen logging system. Metrics can be emitted as structured log events or forwarded to external monitoring systems through custom emitters.
