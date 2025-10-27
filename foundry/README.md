# Foundry Package

The **foundry** package provides immutable reference catalogs and utilities for common development tasks. It serves as a lightweight lookup library for patterns, MIME types, HTTP statuses, country codes, and correlation ID generation.

## Package Overview

Foundry is a **base-layer package** in the gofulmen architecture (per ADR-0001). It provides pure functional operations without dependencies on gofulmen's `errors` or `telemetry` packages, making it suitable for use in any context without introducing import cycles.

**Key Features**:

- Embedded reference data (patterns, MIME types, HTTP statuses, countries)
- Lazy-loaded catalogs with singleton pattern
- UUIDv7-based correlation ID generation
- Fast in-memory lookups with no I/O overhead
- Offline operation (all data embedded at compile time)

## Components

### Catalog

The `Catalog` provides access to reference datasets synced from Crucible SSOT:

```go
catalog := foundry.NewCatalog()

// Pattern lookups
pattern, err := catalog.GetPattern("ansi-email")
if err == nil && pattern.MustMatch("user@example.com") {
    // Valid email
}

// MIME type lookups
mimeType, err := catalog.GetMimeType("application/json")
ext := catalog.GetMimeTypeByExtension(".json")

// HTTP status helpers
group, err := catalog.GetHTTPStatusGroupForCode(404)
helper, err := catalog.GetHTTPStatusHelper()
reason := helper.GetReasonPhrase(404) // "Not Found"

// Country code lookups
country, err := catalog.GetCountry("US")
country, err := catalog.GetCountryByAlpha3("USA")
country, err := catalog.GetCountryByNumeric("840")
```

**Singleton Access**:

```go
catalog := foundry.GetDefaultCatalog()
```

### Correlation IDs

Generate time-sortable UUIDv7 correlation IDs for distributed tracing:

```go
correlationID := foundry.GenerateCorrelationID()
// Example: "018b2c5e-8f4a-7890-b123-456789abcdef"

// Validate correlation ID
if foundry.IsValidCorrelationID(correlationID) {
    // Valid UUIDv7
}

// Parse for timestamp extraction
parsed, err := foundry.ParseCorrelationID(correlationID)
```

**Benefits of UUIDv7**:

- Time-sortable (chronological ordering in logs)
- Globally unique across distributed systems
- Database-friendly (better index performance than UUIDv4)
- Consistent across all Fulmen libraries (Go/Python/TypeScript)

### Context Enrichment

Add correlation and trace context to log events:

```go
ctx := foundry.WithCorrelationID(context.Background(), correlationID)
ctx = foundry.WithTraceID(ctx, traceID)

// Extract later
correlationID := foundry.GetCorrelationID(ctx)
traceID := foundry.GetTraceID(ctx)
```

### Similarity (Subpackage)

Text similarity and suggestion utilities with v1 and v2 APIs (see `similarity/` subdirectory for complete documentation).

**NEW in v0.1.5: v2 API with 5 Algorithms**

```go
import "github.com/fulmenhq/gofulmen/foundry/similarity"

// v1 API (Levenshtein, still supported)
distance := similarity.Distance("kitten", "sitting") // 3
score := similarity.Score("kitten", "sitting")       // 0.5714...

// v2 API with algorithm selection
distance, _ := similarity.DistanceWithAlgorithm("kitten", "sitting", "osa")
score, _ := similarity.ScoreWithAlgorithm("kitten", "sitting", "jaro-winkler")

// Supported algorithms:
// - "levenshtein" - Edit distance (insertion, deletion, substitution)
// - "osa" - Optimal String Alignment (OSA variant of Damerau-Levenshtein)
// - "damerau" - Unrestricted Damerau-Levenshtein (allows multiple edits)
// - "jaro-winkler" - Phonetic similarity (0.0-1.0, higher is more similar)
// - "substring" - Longest common substring ratio

// Suggestion API with fuzzy matching
candidates := []string{"config", "configure", "conform"}
suggestions := similarity.Suggest("confg", candidates, similarity.DefaultSuggestOptions())
// Returns: [{"config", 0.8333}]

// Unicode normalization
normalized := similarity.Normalize("  Café  ", similarity.NormalizeOptions{
    StripAccents: true,
}) // "cafe"
```

**Telemetry Support (NEW in v0.1.5)**

Similarity includes opt-in counter-only telemetry per ADR-0008 Pattern 1:

```go
// Enable telemetry (disabled by default)
similarity.EnableTelemetry(telemetrySystem)

// Metrics emitted (when enabled):
// - foundry.similarity.distance.calls{algorithm}
// - foundry.similarity.score.calls{algorithm}
// - foundry.similarity.string_length{bucket,algorithm}
// - foundry.similarity.fast_path{reason}
// - foundry.similarity.edge_case{case}
// - foundry.similarity.error{type,algorithm,correct_api}

// Disable telemetry
similarity.DisableTelemetry()
```

**Performance**: Native OSA implementation provides 1.24-1.76x performance improvement over external libraries. Telemetry overhead is ~1μs per operation when enabled (negligible for typical use cases).

## Telemetry & Error Handling

### Architecture (ADR-0001)

Foundry is a **base-layer package** that:

- Returns standard Go `error` types
- Does NOT import `github.com/fulmenhq/gofulmen/errors`
- Does NOT import `github.com/fulmenhq/gofulmen/telemetry`
- Maintains zero dependencies on other gofulmen packages

This design prevents import cycles and allows foundry to be used by higher-level packages (config, logging, telemetry) without circular dependencies.

### Instrumentation Pattern (ADR-0008)

Foundry catalog operations follow **Pattern 2: Performance-Sensitive (Counter Only)**:

- **No histograms**: Lookups are in-memory operations called frequently (hot path)
- **Counter-only telemetry**: Emit at the caller boundary, not in foundry itself
- **Minimal overhead**: Avoid 50-100ns histogram overhead per lookup

**For Consumers Using Foundry Catalog**:

When your application uses foundry catalog lookups and needs observability, emit counters at your call site:

```go
import (
    "github.com/fulmenhq/gofulmen/foundry"
    "github.com/fulmenhq/gofulmen/telemetry"
    "github.com/fulmenhq/gofulmen/telemetry/metrics"
)

func LookupPattern(catalog *foundry.Catalog, id string, sys *telemetry.System) (*foundry.Pattern, error) {
    pattern, err := catalog.GetPattern(id)

    if sys != nil {
        status := metrics.StatusSuccess
        if err != nil {
            status = metrics.StatusFailure
        }
        _ = sys.Counter(metrics.FoundryLookupCount, 1, map[string]string{
            metrics.TagComponent: "foundry",
            metrics.TagOperation: "get_pattern",
            metrics.TagStatus:    status,
        })
    }

    return pattern, err
}
```

**Metrics**:

- `foundry_lookup_count` - Total catalog lookups (tagged by operation and status)
- No latency histograms (performance-sensitive pattern)

**Error Wrapping**:

If you need structured error envelopes, wrap foundry errors at your call site:

```go
import "github.com/fulmenhq/gofulmen/errors"

pattern, err := catalog.GetPattern(id)
if err != nil {
    envelope := errors.NewErrorEnvelope("FOUNDRY_LOOKUP_ERROR", "Pattern lookup failed")
    envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
    envelope = errors.SafeWithContext(envelope, map[string]interface{}{
        "component": "foundry",
        "operation": "get_pattern",
        "pattern_id": id,
    })
    envelope = envelope.WithOriginal(err)
    return nil, envelope
}
```

This approach maintains foundry's base-layer status while providing full observability at the consumer level.

## Data Sources

All reference data is synced from Crucible SSOT via `make sync`:

- **Patterns**: `assets/patterns.yaml` (regex patterns for validation)
- **MIME Types**: `assets/mime-types.yaml` (content type mappings)
- **HTTP Statuses**: `assets/http-statuses.yaml` (status codes and groups)
- **Countries**: `assets/country-codes.yaml` (ISO 3166-1 country codes)
- **Similarity Fixtures**: `assets/similarity-fixtures.yaml` (test data)

Data is embedded at compile time using Go's `embed` directive, ensuring offline operation and zero runtime I/O.

## Testing

```bash
# Run foundry tests
go test ./foundry/...

# Run with coverage
go test ./foundry/... -cover

# Run similarity benchmarks
go test ./foundry/similarity -bench=.
```

## API Stability

Foundry is part of gofulmen's stable API. Breaking changes follow semantic versioning and are communicated via release notes.

## See Also

- [ADR-0001](../docs/crucible-go/architecture/decisions/ADR-0001-import-cycle-resolution.md) - Import Cycle Resolution & Layered Architecture
- [ADR-0008](../docs/crucible-go/architecture/decisions/ADR-0008-helper-library-instrumentation-patterns.md) - Helper Library Instrumentation Patterns
- [Telemetry & Metrics](../docs/crucible-go/standards/library/modules/telemetry-metrics.md) - Metrics taxonomy and standards
- [Similarity Package](./similarity/README.md) - Text similarity utilities (if available)

## License

See repository root LICENSE file.
