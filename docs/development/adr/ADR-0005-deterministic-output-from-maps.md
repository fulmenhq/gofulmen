---
id: "ADR-0005"
title: "Sort Map Keys for Deterministic Output in Tests and External Formats"
status: "accepted"
date: "2025-11-04"
deciders: ["@3leapsdave", "Foundation Forge"]
scope: "gofulmen"
tags: ["testing", "maps", "determinism", "prometheus", "observability"]
---

# ADR-0005: Sort Map Keys for Deterministic Output in Tests and External Formats

## Context

Go's map iteration order is **intentionally randomized** by the Go runtime (since Go 1.0) to prevent developers from relying on unspecified behavior. This design decision has important implications for testing and external data format generation.

### The Problem

During CI testing of Prometheus metrics exporter (`telemetry/exporters`), we encountered flaky test failures:

**Test expectation**:

```
app_http_requests_total{method="GET",status="200"} 1000
```

**Actual output (sometimes)**:

```
app_http_requests_total{status="200",method="GET"} 1000
```

The test failed intermittently:

- **Local (macOS, Go 1.25.1, ARM64)**: Failed ~20-30% of runs
- **CI (Ubuntu, Go 1.21-1.25, AMD64)**: Failed consistently

### Root Cause

The `formatPrometheusLabels()` function iterated over a map without sorting keys:

```go
func (e *PrometheusExporter) formatPrometheusLabels(tags map[string]string) string {
    labels := make([]string, 0, len(tags))
    for key, value := range tags {  // ⚠️ NON-DETERMINISTIC ORDER
        labels = append(labels, fmt.Sprintf(`%s="%s"`, key, value))
    }
    return strings.Join(labels, ",")
}
```

**Go Language Specification** ([For statements](https://go.dev/ref/spec#For_statements)):

> "The iteration order over maps is not specified and is not guaranteed to be the same from one iteration to the next."

**Go 1.0 Release Notes** ([Performance](https://go.dev/doc/go1#performance)):

> "In Go 1, the order in which elements are visited when iterating over a map using a for range statement is randomized. This was done to prevent programs from depending on an unspecified behavior."

### Why This Matters

1. **Testing**: String-based assertions become flaky and unreliable
2. **External Formats**: Consumers expect stable output (Prometheus, JSON, logs)
3. **Debugging**: Non-deterministic output makes comparisons difficult
4. **CI/CD**: Platform-specific failures create confusion and false negatives

### Prometheus Context

According to Prometheus documentation, labels are **unordered key-value pairs** - both `{method="GET",status="200"}` and `{status="200",method="GET"}` are semantically identical. However:

- Prometheus itself **sorts labels alphabetically** when ingesting metrics
- Consistent ordering improves debuggability and reduces confusion
- Many monitoring tools display labels in sorted order by convention

## Decision

**Always sort map keys alphabetically when generating external output or test assertions that involve map data.**

### Implementation Pattern

```go
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
        labels = append(labels, fmt.Sprintf(`%s="%s"`, key, value))
    }

    return strings.Join(labels, ",")
}
```

### When to Apply This Pattern

**✅ MUST sort keys when**:

- Generating external data formats (Prometheus metrics, JSON, YAML)
- Creating test assertions based on map data
- Producing logs or diagnostic output consumed by tools
- Generating checksums or hashes of map contents
- Producing output for human comparison or debugging

**❌ OPTIONAL (performance consideration)**:

- Internal processing where order doesn't matter
- Hot-loop operations with performance constraints (measure first!)
- Temporary debugging output for developers only

### Performance Considerations

**Cost of sorting**:

```go
// For small maps (< 10 keys), sorting overhead is negligible
keys := make([]string, 0, len(tags))  // ~16-80 bytes
for key := range tags { keys = append(keys, key) }  // ~5-50ns per key
sort.Strings(keys)  // ~O(n log n), typically <100ns for small n
```

**When sorting matters**:

- Maps with 100+ keys in hot loops (measure first!)
- Real-time metric emission with sub-microsecond budgets
- Pre-allocate key slice capacity when size is known

**Trade-off**: Deterministic output is almost always worth the negligible overhead (<1µs for typical maps).

## Rationale

### 1. Eliminates Test Flakiness

Before fix (100 runs):

- Failures: ~20-30 random failures
- CI failures: Consistent on Ubuntu

After fix (100 runs):

- Failures: 0
- CI failures: 0

### 2. Aligns with External Standards

Prometheus, JSON Schema validators, and most monitoring tools expect or produce sorted output by convention.

### 3. Improves Debuggability

Developers can reliably compare outputs, diff metrics, and spot changes without noise from random ordering.

### 4. Prevents Future Issues

Establishing this as a standard practice prevents similar issues in new code (logging, config export, metric systems).

### 5. Go Community Pattern

Widely adopted pattern in Go ecosystem:

- `encoding/json` sorts map keys in output
- Many Prometheus exporters sort labels
- Standard practice in test fixtures and golden file testing

## Alternatives Considered

### Alternative 1: Fix Tests to Accept Any Order

**Approach**: Update test assertions to accept labels in any order using regex or custom matchers.

**Rejected**:

- More complex test code
- Doesn't solve external format stability
- Still produces confusing diffs in debugging
- Doesn't prevent future issues

### Alternative 2: Document Non-Determinism

**Approach**: Document that output order is random and users should not rely on it.

**Rejected**:

- Prometheus ecosystem expects sorted labels by convention
- Creates confusion for users comparing outputs
- Makes debugging harder
- Doesn't align with ecosystem standards

### Alternative 3: Use Ordered Map Data Structure

**Approach**: Use `sync.Map` or third-party ordered map library.

**Rejected**:

- Adds complexity and dependencies
- Performance overhead for all operations
- `sync.Map` is for concurrency, not ordering
- Standard library doesn't provide ordered maps
- Over-engineered for the problem

### Alternative 4: Sort Only in Tests

**Approach**: Add sorting only in test helper functions.

**Rejected**:

- Production output still non-deterministic
- Users see different ordering than tests validate
- Doesn't help with external tool integration
- Tests should validate actual production behavior

## Consequences

### Positive

- ✅ Eliminates test flakiness (verified: 0 failures in 100 runs)
- ✅ Deterministic Prometheus metrics output
- ✅ Aligns with Prometheus ecosystem conventions
- ✅ Easier debugging and output comparison
- ✅ Consistent behavior across platforms (macOS, Linux, different Go versions)
- ✅ Prevents similar issues in future code

### Negative

- ⚠️ Negligible performance overhead (typically <1µs for small maps)
- ⚠️ Developers must remember to sort when adding new external output

### Neutral

- ℹ️ Adds 3-5 lines of code per formatting function
- ℹ️ Pattern documented for future reference
- ℹ️ Becomes standard practice for gofulmen codebase

## Implementation

### Files Modified

**`telemetry/exporters/prometheus.go`**:

1. Added `import "sort"`
2. Updated `formatPrometheusLabels()` to sort keys before iteration
3. Updated `formatPrometheusLabelsWithAdditional()` (uses sorted helper)

### Testing Validation

**Flakiness test** (100 iterations):

```bash
$ for i in {1..100}; do
    go test -count=1 ./telemetry/exporters -run TestPrometheusMetricTypeRoutingInHandler
  done
# Result: 0 failures
```

**Output verification**:

```
app_http_requests_total{method="GET",status="200"} 1000  ✅ Always alphabetical
app_memory_usage_bytes{host="app1"} 1073741824           ✅ Single label (no change)
app_request_duration_ms_bucket{endpoint="/users",le="10"} 5  ✅ Alphabetical
```

### Future Applications

Apply this pattern to:

- JSON/YAML config export functions
- Logging context field formatting
- Metric tag serialization
- Test fixture generation
- Diagnostic output formatting

## Related Ecosystem ADRs

None. This is a Go-specific decision driven by Go's map iteration randomization. Other languages (Python `dict`, TypeScript `Map`) have different ordering guarantees:

- **Python 3.7+**: Dict insertion order is guaranteed
- **TypeScript/JavaScript**: Map insertion order is guaranteed (ES2015+)
- **Go**: Map iteration order is explicitly randomized

## References

- [Go Spec: For statements](https://go.dev/ref/spec#For_statements) - "The iteration order over maps is not specified"
- [Go 1.0 Release Notes](https://go.dev/doc/go1#performance) - Explains randomization rationale
- [Effective Go: Maps](https://go.dev/doc/effective_go#maps) - Usage patterns
- [Prometheus Exposition Formats](https://prometheus.io/docs/instrumenting/exposition_formats/) - Label semantics
- [Analysis Document](.plans/20251104-prometheus-test-flakiness/analysis.md) - Detailed root cause analysis

## Code Review Checklist

When reviewing code that generates external output from maps, verify:

- [ ] Map keys are sorted before iteration if output is external
- [ ] Test assertions don't rely on random map iteration order
- [ ] External format generation is deterministic
- [ ] Performance impact is acceptable (measure if in hot loop)
- [ ] Documentation notes deterministic ordering when relevant

---

**Decision Outcome**: Always sort map keys alphabetically when generating external output or test assertions. This eliminates test flakiness, aligns with ecosystem conventions, improves debuggability, and prevents future issues. The negligible performance overhead (<1µs) is justified by the reliability and consistency benefits. Verified with 100-iteration test suite showing 0 failures post-fix.
