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

- **Fixture Validation**: Tests validate fixtures against FulHash schema requirements (checksum patterns, required fields, structure)
- **Cross-Language Parity**: Fixtures ensure identical outputs across Go, Python, and TypeScript implementations
- **Block & Streaming**: Tests cover both hashing modes with real digest values

Run tests with schema validation:

```bash
go test ./fulhash/...
```

## Integration

Used by Pathfinder for checksum metadata and Docscribe for integrity verification.

See [Pathfinder Checksum Integration](../.plans/active/v0.1.4/pathfinder-fulhash-checksums.md) and [FulHash Fixture Retrofit](../.plans/active/v0.1.4/fulhash-fixture-schema-retrofit.md) for details.
