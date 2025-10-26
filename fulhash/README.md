# FulHash Package

The `fulhash` package provides canonical hashing utilities for the Fulmen ecosystem, implementing the [FulHash module standard](https://github.com/fulmenhq/crucible/blob/main/docs/standards/library/modules/fulhash.md).

## Features

- **Block Hashing**: One-shot hashing for in-memory data
- **Streaming Hashing**: Incremental hashing for large files/streams
- **Algorithm Support**: xxh3-128 (default, fast) and sha256 (cryptographic)
- **Metadata Formatting**: Standardized `<algorithm>:<hex>` format
- **Enterprise-Ready**: Thread-safe, performant, comprehensive error handling

## Quick Start

```go
import "github.com/fulmenhq/gofulmen/fulhash"

// Block hashing
digest, err := fulhash.Hash([]byte("Hello, World!"))
fmt.Println(digest.String()) // "xxh3-128:531df2844447dd5077db03842cd75395"

// Streaming hashing
hasher, err := fulhash.NewHasher()
if err != nil {
    // handle error
}
io.Copy(hasher, file)
digest := hasher.Sum()

// Parse formatted digest
parsed, err := fulhash.ParseDigest("xxh3-128:abc123...")
```

## Algorithms

| Algorithm | Use Case               | Performance        |
| --------- | ---------------------- | ------------------ |
| xxh3-128  | General integrity      | ~50-100 GB/s       |
| sha256    | Cryptographic security | ~500 MB/s - 2 GB/s |

## API Reference

### Block Hashing

- `Hash(data []byte, opts ...Option) (Digest, error)`: Hash byte slice
- `HashString(s string, opts ...Option) (Digest, error)`: Hash string
- `HashReader(r io.Reader, opts ...Option) (Digest, error)`: Hash from reader

### Streaming Hashing

- `NewHasher(opts ...Option) (Hasher, error)`: Create streaming hasher
- `Hasher.Write(p []byte) (n int, err error)`: Add data
- `Hasher.Sum() Digest`: Finalize and get digest
- `Hasher.Reset()`: Reset for reuse

### Metadata

- `Digest.Algorithm() Algorithm`: Get algorithm
- `Digest.Hex() string`: Get hex string
- `Digest.Bytes() []byte`: Get raw bytes
- `Digest.String() string`: Get formatted string
- `FormatDigest(d Digest) string`: Format digest
- `ParseDigest(s string) (Digest, error)`: Parse formatted string

### Options

- `WithAlgorithm(alg Algorithm)`: Set algorithm
- `WithBufferSize(size int)`: Set buffer size for readers (default 32KiB)

## Performance

Benchmarks show xxh3-128 suitable for high-throughput scenarios:

```
BenchmarkHash_Small_XXH3-10      10000000    120 ns/op
BenchmarkHash_Large_XXH3-10         1000  1.2 ms/op (10MB)
```

For large files, streaming avoids memory allocation.

## Error Handling

All functions return typed errors for unsupported algorithms, invalid formats, and I/O issues.

## Testing & Validation

The package includes comprehensive tests using canonical fixtures synced from Crucible:

- **Schema Validation**: Fixtures are validated against the JSON schema at `schemas/crucible-go/library/fulhash/v1.0.0/fixtures.schema.json` to ensure compliance with required fields and formats
- **Fixture Coverage**: Tests exercise all block, streaming, error, and format fixtures with real digest values
- **Cross-Language Parity**: Fixtures ensure identical outputs across Go, Python, and TypeScript implementations

Run tests with schema validation:

```bash
go test ./fulhash/...
```

Schema requirements are documented in [`docs/crucible-go/standards/library/modules/fulhash.md`](docs/crucible-go/standards/library/modules/fulhash.md). When adding new fixtures, ensure they validate against the schema and include required fields like `xxh3_128`, `sha256` for block fixtures, and `expected_xxh3_128`, `expected_sha256` for streaming fixtures.

## Telemetry & Error Handling

FulHash supports optional telemetry metrics for all hash operations using ADR-0008 Pattern 2 (Performance-Sensitive - Counter Only). **Telemetry is disabled by default** and must be explicitly enabled via `SetTelemetrySystem()`.

### Emitted Metrics (when enabled)

| Metric                 | Type    | Tags                          | Description                                    |
| ---------------------- | ------- | ----------------------------- | ---------------------------------------------- |
| `fulhash_hash_count`   | Counter | algorithm, status             | Hash operations (Hash, HashString, HashReader) |
| `fulhash_errors_count` | Counter | algorithm, status, error_type | Hash failures (I/O, unsupported algorithm)     |

**Note**: Counters only, no histograms (ADR-0008 Pattern 2: performance-sensitive operations).

### Example: Default Behavior (No Telemetry)

```go
import "github.com/fulmenhq/gofulmen/fulhash"

// Telemetry disabled by default - no metrics emitted
digest, err := fulhash.Hash([]byte("data"))
```

### Example: Enable Telemetry

```go
import (
    "github.com/fulmenhq/gofulmen/fulhash"
    "github.com/fulmenhq/gofulmen/telemetry"
)

// Explicitly enable telemetry
telSys, _ := telemetry.NewSystem(&telemetry.Config{
    Enabled: true,
    Emitter: myCustomEmitter, // Your Prometheus/OTLP/etc. emitter
})

fulhash.SetTelemetrySystem(telSys)

// Now emits fulhash_hash_count{algorithm="xxh3-128",status="success"}
digest, err := fulhash.Hash([]byte("data"))
```

### Example: Disable Telemetry After Enabling

```go
fulhash.SetTelemetrySystem(nil)
```

### Advanced Usage: Wrapping with Error Envelopes

For application-level error handling with structured error envelopes and additional telemetry:

```go
import (
    "github.com/fulmenhq/gofulmen/errors"
    "github.com/fulmenhq/gofulmen/fulhash"
    "github.com/fulmenhq/gofulmen/telemetry"
    "github.com/fulmenhq/gofulmen/telemetry/metrics"
)

func hashFileWithEnvelope(path string) (fulhash.Digest, error) {
    telSys := getTelemetrySystem() // Your app's telemetry system

    file, err := os.Open(path)
    if err != nil {
        envelope := errors.NewErrorEnvelope(
            "FILE_HASH_ERROR",
            "failed to open file for hashing",
        )
        envelope = envelope.
            WithSeverity(errors.SeverityHigh).
            WithContext("path", path).
            WithOriginal(err)

        if telSys != nil {
            _ = telSys.Counter("app_file_hash_errors", 1, map[string]string{
                metrics.TagErrorType: "file_open",
            })
        }
        return fulhash.Digest{}, envelope
    }
    defer file.Close()

    // FulHash automatically emits its own metrics
    digest, err := fulhash.HashReader(file)
    if err != nil {
        envelope := errors.NewErrorEnvelope(
            "FILE_HASH_ERROR",
            "hash computation failed",
        )
        envelope = envelope.
            WithSeverity(errors.SeverityMedium).
            WithContext("path", path).
            WithOriginal(err)

        if telSys != nil {
            _ = telSys.Counter("app_file_hash_errors", 1, map[string]string{
                metrics.TagErrorType: "hash_computation",
            })
        }
        return fulhash.Digest{}, envelope
    }

    return digest, nil
}
```

This pattern demonstrates:

- FulHash emits low-level counters (`fulhash_hash_count`, `fulhash_errors_count`)
- Application layer adds error envelopes for structured error handling
- Application layer emits domain-specific metrics (`app_file_hash_errors`)
- Original errors preserved via `WithOriginal()` for debuggability

### Cross-Language Parity

FulHash telemetry matches PyFulmen's implementation for consistency across Fulmen helper libraries. See ADR-0008 (Helper Library Instrumentation Patterns) for ecosystem-wide standards.

## Integration

Used by Pathfinder for checksum metadata and Docscribe for integrity verification.

See [Pathfinder Checksum Integration](../.plans/active/v0.1.4/pathfinder-fulhash-checksums.md) and [FulHash Fixture Retrofit](../.plans/active/v0.1.4/fulhash-fixture-schema-retrofit.md) for details.
