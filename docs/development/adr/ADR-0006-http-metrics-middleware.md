---
id: "ADR-0006"
title: "HTTP Server Metrics Middleware Implementation"
status: "accepted"
date: "2025-11-17"
deciders: ["@3leapsdave"]
scope: "gofulmen"
tags: ["http", "metrics", "middleware", "performance", "observability"]
---

## Context

The gofulmen telemetry package lacked HTTP server metrics collection, requiring users to implement custom metrics or go without observability. The Fulmen ecosystem needed a standardized, performant way to collect the 5 HTTP metrics defined in Crucible v0.2.18 taxonomy while maintaining low overhead and preventing cardinality explosion.

Key requirements:

- Support all 5 HTTP metrics from Crucible taxonomy
- Framework integration (net/http, Chi, Gin)
- Performance suitable for production (<50% overhead target)
- Cardinality control through route normalization
- Thread-safe concurrent operation
- Configurable service names and bucket options

## Decision

Implement a comprehensive HTTP server metrics middleware with the following architecture:

### Core Components

1. **HTTPMetricsMiddleware**: Factory function creating middleware handlers
2. **httpMetricsHandler**: Core middleware implementation with pooled resources
3. **DefaultRouteNormalizer**: Route normalization preventing cardinality explosion
4. **emitSizeHistogram**: Helper for size histogram bucket construction

### Performance Optimizations

Achieves **~26-35% per-request overhead** through:

1. **Tag Pooling**: Use `sync.Pool` for tag map reuse to reduce allocations
2. **Histogram Bucket Pooling**: Pool histogram slices for size metrics
3. **Pre-compiled UUID Regex**: Global pattern to avoid recompilation overhead
4. **Fast-path Route Handling**: Optimize simple routes that don't need normalization

### Route Normalization Strategy

- **UUID Segments**: Replace UUID patterns with `{uuid}` placeholder
- **Numeric Segments**: Replace pure numeric segments with `{id}` placeholder
- **Query Stripping**: Remove query parameters and fragments
- **Path Preservation**: Maintain leading slash for proper route matching

### API Design

```go
// Core middleware creation
middleware := telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("my-api"),
    telemetry.WithCustomSizeBuckets(sizeBuckets),
)

// Framework integration patterns documented
handler := middleware(http.HandlerFunc(handler))
```

## Rationale

### Why Middleware Pattern

- **Framework Agnostic**: Works with any HTTP router supporting http.Handler interface
- **Composable**: Can be stacked with other middleware (CORS, auth, logging)
- **Standard Go Pattern**: Familiar to Go developers using net/http ecosystem

### Why Route Normalization

- **Cardinality Control**: Prevent unlimited metric cardinality from dynamic routes
- **Taxonomy Compliance**: Aligns with Crucible v0.2.18 HTTP metrics standards
- **Performance**: Reduces memory usage and metric storage overhead

### Why Pooling Strategy

- **Allocation Reduction**: Tag maps and histogram slices are major allocation sources
- **GC Pressure**: Fewer allocations reduce garbage collection overhead
- **Concurrent Safety**: `sync.Pool` provides thread-safe resource reuse

### Why Size-Only Bucket Configuration

- **Emitter-Driven Duration**: Duration buckets are typically configured at emitter level
- **API Honesty**: Remove misleading options that don't affect behavior
- **Simplicity**: Clearer API with fewer configuration footguns

## Alternatives Considered

### Alternative 1: Per-Request Metric Objects

Create metric objects per request with method chaining.

- **Rejected**: Too much allocation overhead, complex API

### Alternative 2: Global Metric Registry

Use global registry for HTTP metrics similar to Prometheus client.

- **Rejected**: Not thread-safe for concurrent use, testing difficulties

### Alternative 3: Framework-Specific Implementations

Separate implementations for Chi, Gin, etc.

- **Rejected**: Code duplication, maintenance overhead

### Alternative 4: Duration Bucket Configuration

Allow custom duration buckets alongside size buckets.

- **Rejected**: Duration buckets are emitter-driven, misleading API

## Consequences

### Positive Impacts

- **Comprehensive Observability**: All 5 HTTP metrics with proper taxonomy alignment
- **Production Performance**: ~26-35% per-request overhead is acceptable for comprehensive instrumentation
- **Cardinality Safety**: Route normalization prevents metric explosion
- **Framework Integration**: Works with Chi, Gin, and standard net/http
- **Developer Experience**: Clear API with comprehensive documentation

### Negative Impacts

- **Memory Usage**: Pooled resources increase baseline memory usage
- **Complexity**: Route normalization logic adds implementation complexity
- **Test Surface**: More comprehensive testing required for all edge cases

### Neutral Impacts

- **Dependency Size**: No new external dependencies added
- **API Surface**: New functions but no breaking changes to existing code

## Implementation Notes

### Key Technical Decisions

**Performance Characteristics**: The implementation achieves **~26-35% per-request overhead** (measured via benchmarking at 316-410ns per request over a 1200-1300ns baseline) through aggressive pooling and pre-compilation strategies. This overhead is acceptable for comprehensive HTTP instrumentation collecting 5 distinct metrics per request. Overhead breakdown: ~72% from route normalization, ~2% from metric emission.

**Cardinality Management**: Route normalization replaces UUID patterns and numeric IDs with placeholders (`{uuid}`, `{id}`), preventing unbounded metric cardinality from dynamic routes. This is critical for production metrics systems where cardinality explosion can cause storage and query performance issues.

**Framework Integration**: The middleware uses Go's standard `http.Handler` interface, enabling zero-cost integration with any compliant router (Chi, Gin, net/http stdlib). No framework-specific code paths are required.

**Thread Safety**: Uses `sync.Pool` for tag map and histogram bucket reuse. Pool-based resource management provides thread-safe concurrent access without locks in the hot path.

## Related Ecosystem ADRs

None currently. This implementation follows existing Crucible v0.2.18 HTTP metrics taxonomy.

## References

- [Crucible v0.2.18 HTTP Metrics Taxonomy](../../../schemas/crucible-go/observability/metrics/v1.0.0/)
- [Go HTTP Handler Interface](https://golang.org/pkg/net/http/#Handler)
- [Chi Router Documentation](https://github.com/go-chi/chi)
- [Gin Router Documentation](https://github.com/gin-gonic/gin)
- [sync.Pool Documentation](https://golang.org/pkg/sync/#Pool)
- [Prometheus Histogram Best Practices](https://prometheus.io/docs/practices/histograms/)

---

**Decision Status**: Accepted  
**Deciders**: @3leapsdave  
**Date**: 2025-11-17
