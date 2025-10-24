# Error & Telemetry Dogfooding Guide

**Version**: 0.1.5  
**Status**: Active - Phase 0 Complete  
**Updated**: 2025-10-24

## Overview

This guide documents how gofulmen dogfoods its own `errors` and `telemetry` packages across all modules. The pattern established here serves as a reference for TypeScript and Python foundation teams.

## Core Principles

### 1. Return Structured Errors Directly

**Pattern**: Return `*errors.ErrorEnvelope` directly from all exported functions.

**Why**: We're pre-release, so no backward compatibility concerns. Structured errors provide:

- Direct access to severity, correlation IDs, and context
- Better error handling for internal module-to-module calls
- Clearer API contracts

```go
func DoSomething() (*Result, *errors.ErrorEnvelope) {
    if err := validate(); err != nil {
        envelope := errors.NewErrorEnvelope("VALIDATION_ERROR", "Validation failed")
        envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
        return nil, envelope
    }
    return result, nil
}
```

### 2. Preserve Sentinel Errors

**Pattern**: Use `WithOriginal()` to maintain `errors.Is()` compatibility.

```go
if !exists {
    envelope := errors.NewErrorEnvelope("FILE_NOT_FOUND", "Config file not found")
    envelope = envelope.WithOriginal(os.ErrNotExist)  // Preserves errors.Is(err, os.ErrNotExist)
    return nil, envelope
}
```

### 3. Use Metric Name Constants

**Pattern**: Always use constants from `telemetry/metrics` package.

```go
import "github.com/fulmenhq/gofulmen/telemetry/metrics"

sys.Histogram(metrics.ConfigLoadMs, duration, map[string]string{
    metrics.TagStatus: metrics.StatusSuccess,
    metrics.TagCategory: "terminal",
})
```

**Benefits**:

- Compile-time checking (typos caught at build time)
- Easier refactoring and grep/search
- Enforces taxonomy alignment

### 4. Emit Metrics Once Per Call

**Pattern**: Single histogram at function boundary, counters in specific branches.

```go
func ProcessData() error {
    start := time.Now()
    status := metrics.StatusSuccess

    defer func() {
        if sys != nil {
            _ = sys.Histogram(metrics.ProcessDurationMs, time.Since(start), map[string]string{
                metrics.TagStatus: status,
            })
        }
    }()

    if err := validate(); err != nil {
        status = metrics.StatusError
        _ = sys.Counter(metrics.ValidationErrors, 1, nil)
        return err
    }

    return nil
}
```

## Test Infrastructure

### FakeCollector Usage

The `telemetry/testing` package provides `FakeCollector` for asserting metrics in tests.

```go
import (
    "testing"
    "time"

    "github.com/fulmenhq/gofulmen/telemetry"
    telemetrytesting "github.com/fulmenhq/gofulmen/telemetry/testing"
    "github.com/fulmenhq/gofulmen/telemetry/metrics"
)

func TestFunctionEmitsMetrics(t *testing.T) {
    // Create fake collector
    fc := telemetrytesting.NewFakeCollector()

    // Create telemetry system with fake collector
    sys, _ := telemetry.NewSystem(&telemetry.Config{
        Enabled: true,
        Emitter: fc,  // Inject fake collector
    })

    // Run your function
    result, err := myFunction(sys)

    // Assert metrics were emitted
    if !fc.HasMetric(metrics.MyFunctionDurationMs) {
        t.Error("Expected duration metric to be emitted")
    }

    // Check metric count
    if fc.CountMetricsByName(metrics.MyFunctionDurationMs) != 1 {
        t.Errorf("Expected 1 metric, got %d", fc.CountMetricsByName(metrics.MyFunctionDurationMs))
    }

    // Inspect metric details
    metricRecords := fc.GetMetricsByName(metrics.MyFunctionDurationMs)
    if len(metricRecords) > 0 {
        m := metricRecords[0]
        if m.Tags[metrics.TagStatus] != metrics.StatusSuccess {
            t.Errorf("Expected status=success, got %s", m.Tags[metrics.TagStatus])
        }
    }
}
```

### FakeCollector API

```go
// Record metrics
fc.Counter(name string, value float64, tags map[string]string) error
fc.Gauge(name string, value float64, tags map[string]string) error
fc.Histogram(name string, value time.Duration, tags map[string]string) error

// Query metrics
fc.GetMetrics() []RecordedMetric
fc.GetMetricsByName(name string) []RecordedMetric
fc.GetMetricsByType(metricType MetricType) []RecordedMetric
fc.CountMetrics() int
fc.CountMetricsByName(name string) int
fc.HasMetric(name string) bool

// Reset between tests
fc.Reset()
```

## Metric Names Reference

From `telemetry/metrics/names.go`:

| Constant                     | Metric Name                    | Type      | Description                              |
| ---------------------------- | ------------------------------ | --------- | ---------------------------------------- |
| `SchemaValidations`          | `schema_validations`           | Counter   | Schema validation calls                  |
| `SchemaValidationErrors`     | `schema_validation_errors`     | Counter   | Schema validation failures               |
| `ConfigLoadMs`               | `config_load_ms`               | Histogram | Config load duration                     |
| `ConfigLoadErrors`           | `config_load_errors`           | Counter   | Config load failures                     |
| `PathfinderFindMs`           | `pathfinder_find_ms`           | Histogram | File find duration                       |
| `PathfinderValidationErrors` | `pathfinder_validation_errors` | Counter   | Pathfinder validation failures           |
| `PathfinderSecurityWarnings` | `pathfinder_security_warnings` | Counter   | Security warnings (path traversal, etc.) |
| `FoundryLookupCount`         | `foundry_lookup_count`         | Counter   | Foundry catalog lookups                  |
| `LoggingEmitCount`           | `logging_emit_count`           | Counter   | Log events emitted                       |
| `LoggingEmitLatencyMs`       | `logging_emit_latency_ms`      | Histogram | Log event write duration                 |
| `GoneatCommandDurationMs`    | `goneat_command_duration_ms`   | Histogram | Goneat command execution time            |

### Common Tags

| Constant       | Tag Name    | Common Values                       |
| -------------- | ----------- | ----------------------------------- |
| `TagStatus`    | `status`    | `success`, `failure`, `error`       |
| `TagComponent` | `component` | `pathfinder`, `config`, `schema`    |
| `TagOperation` | `operation` | `validate`, `load`, `find`          |
| `TagCategory`  | `category`  | `terminal`, `logging`, etc.         |
| `TagSeverity`  | `severity`  | `low`, `medium`, `high`, `critical` |

## Error Envelope Patterns

### Basic Error with Severity

```go
envelope := errors.NewErrorEnvelope("VALIDATION_FAILED", "Input validation failed")
envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
```

### Error with Context

```go
envelope := errors.NewErrorEnvelope("CONFIG_LOAD_ERROR", "Failed to load config")
envelope = errors.SafeWithContext(envelope, map[string]interface{}{
    "component": "config",
    "operation": "load",
    "path":      configPath,
    "layer":     "defaults",
})
```

### Error with Correlation ID

```go
envelope := errors.NewErrorEnvelope("SCHEMA_ERROR", "Schema validation failed")
envelope = envelope.WithCorrelationID(correlationID)
```

### Complete Pattern

```go
func LoadConfig(path string, correlationID string) (*Config, *errors.ErrorEnvelope) {
    data, err := os.ReadFile(path)
    if err != nil {
        envelope := errors.NewErrorEnvelope("CONFIG_READ_ERROR", "Failed to read config file")
        envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
        envelope = envelope.WithCorrelationID(correlationID)
        envelope = envelope.WithOriginal(err)  // Preserve os.ErrNotExist etc
        envelope = errors.SafeWithContext(envelope, map[string]interface{}{
            "component": "config",
            "operation": "read",
            "path":      path,
        })
        return nil, envelope
    }

    // ... parse config ...

    return config, nil
}
```

## Implementation Phases

See `.plans/active/v0.1.5/error-telemetry-retrofit-implementation.md` for full rollout plan.

- **Phase 0** ✅: Test infrastructure + metric constants (COMPLETE)
- **Phase 1**: Pathfinder retrofit (IN PROGRESS)
- **Phase 2**: Config retrofit
- **Phase 3**: Schema retrofit
- **Phase 4**: Logging self-instrumentation
- **Phase 5**: Foundry, Docscribe, ASCII
- **Phase 6**: Bootstrap & CLI tools

## Performance Considerations

- Target: <5% overhead from instrumentation
- Histogram operations are ~100ns
- Counter operations are ~50ns
- Batch metrics where appropriate
- Allow disabling telemetry via config for minimal builds

## Cross-Language Alignment

This pattern is designed for consistency with:

- **PyFulmen** (Python) - v0.1.6 telemetry rollout
- **TSFulmen** (TypeScript) - Future implementation
- **Fulmen** (Rust, C#) - Future languages

Key alignment points:

- Metric names from shared taxonomy
- Error envelope structure
- Tag naming conventions
- Test infrastructure patterns

## See Also

- `telemetry/README.md` - Telemetry package documentation
- `errors/ERROR_HANDLING.md` - Error envelope documentation
- `config/crucible-go/taxonomy/metrics.yaml` - Canonical metric taxonomy
- `.plans/active/v0.1.5/error-telemetry-retrofit-implementation.md` - Full implementation plan
