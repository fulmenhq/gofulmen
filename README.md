# gofulmen

**Stop reinventing catalogs. Start shipping.**

Every team writes their own HTTP status helpers, exit code enums, and country code lookups. gofulmen provides production-grade Go implementations derived from a single source of truth—so your Go services use the same codes as your Rust, Python, and TypeScript services.

- **Zero runtime dependencies**: All catalogs embedded at compile time
- **Cross-language parity**: Same exit codes, signals, and schemas as rsfulmen, pyfulmen, tsfulmen
- **Reference implementation**: gofulmen is the canonical implementation other languages follow

📖 **[Read the complete library overview](docs/gofulmen_overview.md)** for comprehensive documentation including module catalog, dependency map, and roadmap.

## Who Should Use This

**Platform Engineers & SREs**: Standardize exit codes across all services so alerting thresholds and runbooks work consistently—whether the service is written in Go, Rust, Python, or TypeScript.

**Security & Compliance Teams**: Embedded catalogs eliminate network calls for reference data. Audit the dependency tree once with `go mod graph`.

**Polyglot Teams**: When your organization runs multiple languages, gofulmen ensures your Go services speak the same language as the rest of your stack. Same HTTP status groupings. Same signal handling semantics. Same error codes.

**Library Authors**: Build on gofulmen's catalogs instead of maintaining your own. Context-aware APIs with proper timeout handling built in.

## Crucible Overview

**What is Crucible?**

Crucible is the FulmenHQ single source of truth (SSOT) for schemas, standards, and configuration templates. It ensures consistent APIs, documentation structures, and behavioral contracts across all language foundations (gofulmen, pyfulmen, tsfulmen, etc.).

**Why the Shim & Docscribe Module?**

Rather than copying Crucible assets into every project, helper libraries provide idiomatic access through shim APIs. This keeps your application lightweight, versioned correctly, and aligned with ecosystem-wide standards. The docscribe module lets you discover, parse, and validate Crucible content programmatically without manual file management.

**Where to Learn More:**

- [Crucible Repository](https://github.com/fulmenhq/crucible) - SSOT schemas, docs, and configs
- [Fulmen Technical Manifesto](docs/crucible-go/architecture/fulmen-technical-manifesto.md) - Philosophy and design principles

## Packages

### App Identity (`appidentity/`)

Application identity metadata from `.fulmen/app.yaml` for consistent configuration, logging, and telemetry.

- Automatic discovery with ancestor search
- Schema validation with detailed diagnostics
- Thread-safe caching with sync.Once
- Context-based testing overrides
- Config/CLI/telemetry integration helpers
- Zero Fulmen dependencies (Layer 0)

### ASCII (`ascii/`)

Terminal and Unicode utilities for Go applications.

- Unicode-aware string width calculation
- Terminal-specific character width overrides
- Box drawing utilities
- Interactive calibration tools
- String analysis utilities

### Schema (`schema/`)

JSON Schema validation with support for Draft-04 through Draft-2020-12.

- Schema loading and caching
- Versioned schema registry
- Validation with detailed error reporting
- Support for YAML and JSON
- Multi-draft metaschema resolution (04, 06, 07, 2019-09, 2020-12)

### Config (`config/`)

Configuration management with XDG Base Directory support.

- XDG Base Directory compliance
- Config file discovery patterns
- Environment variable handling

### Bootstrap (`bootstrap/`)

Simple, dependency-free tool installation for Go repositories using the goneat bootstrap pattern.

- Install tools from GitHub releases
- SHA-256 checksum verification
- Support for tar.gz and zip archives
- Cross-platform (macOS, Linux, Windows)
- Local development overrides via `.goneat/tools.local.yaml` (goneat doctor pattern; never commit)

### Logging (`logging/`)

Progressive logging system with profiles, middleware pipeline, and policy enforcement.

- **Progressive Profiles**: SIMPLE (minimal), STRUCTURED (JSON + middleware), ENTERPRISE (full observability)
- **Middleware Pipeline**: Pluggable event processing (correlation, redaction, throttling)
- **Redaction Middleware**: Pattern and field-based PII/secrets filtering (API keys, passwords, SSNs, credit cards)
- **Correlation Middleware**: UUIDv7 correlation ID injection for distributed tracing
- **Throttling Middleware**: Token bucket rate limiting with configurable drop policies
- **Full Crucible Envelope**: 20+ fields including traceId, spanId, contextId, requestId
- **Schema Validation**: Automatic validation against Crucible logging schema
- **Backward Compatible**: 100% compatibility with existing configurations

### Pathfinder (`pathfinder/`)

Safe filesystem discovery with path traversal protection and repository root detection.

- **File Discovery**: Glob-based pattern matching with include/exclude support
- **Security**: Path traversal prevention, boundary enforcement, hidden file filtering
- **Repository Root Discovery**: Safe upward traversal with multiple marker sets (Git, Go, Node, Python, Monorepo)
- **Safety Boundaries**: Home directory ceiling, filesystem root detection, max depth limits
- **Symlink Protection**: Loop detection, boundary enforcement when following symlinks
- **Checksum Support**: Optional integrity verification using FulHash (xxh3-128, sha256)
- **Performance**: <30µs for repository root discovery, <10% overhead for checksums
- **Cross-Platform**: Works on macOS, Linux, Windows with platform-specific path handling

### Foundry (`foundry/`)

Enterprise-grade foundation utilities providing consistent cross-language implementations from Crucible catalogs.

- **Time Utilities**: RFC3339Nano timestamps with nanosecond precision
- **Correlation IDs**: UUIDv7 time-sortable IDs for distributed tracing
- **Pattern Matching**: Regex, glob, and literal patterns from Crucible catalogs
- **MIME Type Detection**: Content-based detection and extension lookup
- **HTTP Status Helpers**: Status code grouping and validation
- **Country Code Validation**: ISO 3166-1 country codes (Alpha2, Alpha3, Numeric)
- **Exit Codes**: 54 standardized exit codes with metadata, platform detection, simplified mode mapping, BSD sysexits.h compatibility
- **Signal Resolution** (`foundry/signals/`): Ergonomic signal name resolution (`ResolveSignal`, `ListSignalNames`, `MatchSignalNames`) for CLI applications with cross-language parity
- **Text Similarity** (`foundry/similarity/`): v1 API (Levenshtein) + v2 API (5 algorithms: Levenshtein, OSA, Damerau, Jaro-Winkler, Substring), normalized scoring, fuzzy matching, Unicode normalization, opt-in telemetry

All Foundry catalogs are embedded at compile time and work offline - no network dependencies required.

### Fulencode (`fulencode/`)

Canonical encoding/decoding library with built-in security protections for cross-language consistency.

- **Encode**: Base64/Base64URL/Base64-raw, Base32/Base32hex, Hex + character encodings (UTF-8/16, ISO-8859-1, CP1252, ASCII)
- **Decode**: All formats with expansion ratio limits, max size protection, checksum verification
- **Detect**: BOM detection, UTF-16 null-pattern heuristic, UTF-8 validation, confidence scoring
- **Normalize**: NFC/NFD/NFKC/NFKD + security-focused `text_safe` profile
- **BOM Helpers**: `DetectBOM`, `RemoveBOM`, `AddBOM` for byte order mark handling
- **Security by Default**: Expansion ratio limits (10x), max size (100MB decoded/500MB encoded), control char rejection
- **SSOT Integration**: Uses Crucible-generated enums for cross-language parity

### Telemetry (`telemetry/`)

Structured metrics emission with support for counters, gauges, and histograms. Includes production-grade Prometheus exporter.

- **Core Metrics**: Counters, gauges, histograms with automatic unit conversion
- **Custom Exporters**: Pluggable emitter interface
- **Prometheus Exporter** (`telemetry/exporters/`): HTTP metrics exposition with enterprise features
  - Bearer token authentication
  - Per-IP rate limiting (configurable requests/minute and burst)
  - 7 built-in health metrics tracking exporter performance
  - Automatic format conversion (ms→seconds for histograms)
  - Three-phase refresh pipeline (collect, convert, export)
- **Thread-Safe**: Concurrent metric emission across goroutines
- **Schema Validation**: Automatic validation against Crucible metrics schema

### Signals (`signals/`)

Cross-platform signal handling with graceful shutdown, config reload, and Windows fallback support.

- **Graceful Shutdown**: LIFO cleanup chains with context support
- **Config Reload**: SIGHUP with validation hooks and restart semantics
- **Ctrl+C Double-Tap**: 2-second window for force quit (configurable)
- **Windows Fallback**: HTTP admin endpoint for unsupported signals
- **Rate Limiting**: Built-in request throttling for HTTP endpoint
- **Thread-Safe**: Concurrent handler registration and execution

Signal definitions and behaviors come from Crucible catalog (v1.0.0) ensuring cross-language parity.

## Installation

```bash
go get github.com/fulmenhq/gofulmen
```

## Usage

### App Identity Package

```go
import "github.com/fulmenhq/gofulmen/appidentity"

// Load application identity from .fulmen/app.yaml
identity, err := appidentity.Get(ctx)
if err != nil {
    log.Fatal(err)
}

// Use identity for configuration
vendor, name := identity.ConfigParams()
configPath := configpaths.GetAppConfigDir(vendor, name)

// Construct environment variables
logLevelVar := identity.EnvVar("LOG_LEVEL")
os.Getenv(logLevelVar) // MYAPP_LOG_LEVEL

// Get telemetry namespace
namespace := identity.TelemetryNamespace()

// For testing, use context override
testIdentity := appidentity.NewFixture()
ctx = appidentity.WithIdentity(ctx, testIdentity)
```

### ASCII Package

```go
import "github.com/fulmenhq/gofulmen/ascii"

// Draw a box around content
box := ascii.DrawBox("Hello, World!", 20)
fmt.Print(box)

// Calculate string width
width := ascii.StringWidth("Café 🚀")
fmt.Printf("Width: %d\n", width)

// Analyze string properties
analysis := ascii.Analyze("Hello\nWorld")
fmt.Printf("Lines: %d, Unicode: %v\n", analysis.LineCount, analysis.HasUnicode)
```

### Schema Package

```go
import "github.com/fulmenhq/gofulmen/schema"

// Create a validator
schemaData := []byte(`{"type": "string"}`)
validator, err := schema.NewValidator(schemaData)
if err != nil {
    log.Fatal(err)
}

// Validate data
diagnostics, err := validator.ValidateData("hello")
if err != nil {
    log.Fatal(err)
}
if len(diagnostics) > 0 {
    for _, d := range diagnostics {
        fmt.Printf("%s: %s\n", d.Pointer, d.Message)
    }
}

// Export schemas with provenance metadata
import "github.com/fulmenhq/gofulmen/schema/export"

opts := export.NewExportOptions(
    "observability/logging/v1.0.0/log-event.schema.json",
    "vendor/crucible/schemas/logging-event.schema.json",
)
if err := export.Export(context.Background(), opts); err != nil {
    log.Fatal(err)
}
```

**CLI Export:**

```bash
# Export schema with provenance
gofulmen-export-schema \
    --schema-id=observability/logging/v1.0.0/log-event.schema.json \
    --out=vendor/crucible/schemas/logging-event.schema.json

# Export as YAML
gofulmen-export-schema \
    --schema-id=terminal/v1.0.0/schema.json \
    --out=schema.yaml \
    --format=yaml
```

See [docs/schema/export.md](docs/schema/export.md) for detailed export documentation.

### Config Package

```go
import "github.com/fulmenhq/gofulmen/config"

// Load configuration
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Get XDG directories
xdg := config.GetXDGBaseDirs()
fmt.Printf("Config dir: %s\n", xdg.ConfigHome)

// Three-layer configuration (defaults + user + runtime)
opts := config.LayeredConfigOptions{
    Category:     "sample",
    Version:      "v1.0.0",
    DefaultsFile: "sample-defaults.yaml",
    SchemaID:     "sample/v1.0.0/schema",
}

merged, diagnostics, err := config.LoadLayeredConfig(opts,
    map[string]any{"settings": map[string]any{"retries": 5}},
)
if err != nil {
    log.Fatal(err)
}
if len(diagnostics) > 0 {
    log.Fatalf("validation issues: %v", diagnostics)
}

fmt.Printf("Retries => %v\n", merged["settings"].(map[string]any)["retries"])

// Environment variable overrides
envOverrides, err := config.LoadEnvOverrides([]config.EnvVarSpec{
    {Name: "APP_RETRIES", Path: []string{"settings", "retries"}, Type: config.EnvInt},
})
if err != nil {
    log.Fatal(err)
}

merged, _, _ = config.LoadLayeredConfig(opts, envOverrides)
fmt.Printf("Retries (env) => %v\n", merged["settings"].(map[string]any)["retries"])

// Environment variable aliases + conflict diagnostics (optional)
report, err := config.LoadEnvOverridesWithReport([]config.EnvVarSpecWithAliases{
    {
        Name:    "APP_SERVER_PORT",
        Aliases: []string{"APP_PORT"},
        Path:    []string{"server", "port"},
        Type:    config.EnvInt,
    },
})
if err != nil {
    log.Fatal(err)
}
if len(report.Conflicts) > 0 {
    // Conflicts include masked values by default for sensitive env vars.
    log.Printf("env conflicts: %+v", report.Conflicts)
}

merged, _, _ = config.LoadLayeredConfig(opts, report.Overrides)
fmt.Printf("Port (env) => %v\n", merged["server"].(map[string]any)["port"])

// Merge schemas at runtime (base + overlay)
mergedSchema, _ := schema.MergeJSONSchemas(
    []byte(`{"type":"object"}`),
    []byte(`{"properties":{"name":{"type":"string"}}}`),
)
fmt.Printf("Merged schema: %s\n", string(mergedSchema))
```

### Logging Package

```go
import "github.com/fulmenhq/gofulmen/logging"

// Create logger with SIMPLE profile (minimal output)
logger, err := logging.BundleSimple()
if err != nil {
    log.Fatal(err)
}

logger.Info("Application started", map[string]any{
    "version": "1.0.0",
    "environment": "production",
})

// Create logger with STRUCTURED profile + redaction middleware
logger, err = logging.BundleStructuredWithRedaction()
if err != nil {
    log.Fatal(err)
}

// Automatically redacts sensitive data
logger.Info("User login", map[string]any{
    "username": "alice",
    "password": "secret123",  // Will be redacted as [REDACTED]
    "apiKey": "sk_live_abc123", // Will be redacted
})

// Diagnostics helpers (for envinfo/doctor):
_ = logging.IsSensitiveKey("GITHUB_TOKEN")

// Custom redaction configuration
redactionConfig := logging.RedactionConfig{
    Patterns: []string{
        `\b[A-Z0-9]{20,}\b`,  // API tokens
        `credit_card=\d+`,    // Credit card patterns
    },
    Fields: []string{"ssn", "tax_id"},
    ReplacementMode: "hash",  // Use hash prefix instead of [REDACTED]
}

logger, err = logging.New(logging.LoggingConfig{
    Profile: "STRUCTURED",
    MinSeverity: "INFO",
    Middleware: []logging.MiddlewareConfig{
        logging.WithRedaction(redactionConfig),
    },
})

// Progressive profiles for different environments
// SIMPLE: Development (minimal output)
// STRUCTURED: Staging (JSON with middleware)
// ENTERPRISE: Production (full observability with correlation, throttling, policies)
```

### Pathfinder Package

```go
import "github.com/fulmenhq/gofulmen/pathfinder"

// Find all Go files in a directory
query := pathfinder.FindQuery{
    Root: "/path/to/project",
    Include: []string{"**/*.go"},
    Exclude: []string{"vendor/**", "**/testdata/**"},
}

results, err := pathfinder.Find(query)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Found: %s (%d bytes)\n", result.Path, result.Size)
}

// Find repository root (safe upward traversal)
rootPath, err := pathfinder.FindRepositoryRoot(".")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Repository root: %s\n", rootPath)

// Find repository root with custom markers and boundaries
opts := []pathfinder.RootOption{
    pathfinder.WithMarkers([]string{".git", "go.mod", "package.json"}),
    pathfinder.WithMaxDepth(5),
    pathfinder.WithBoundary("/home/user/projects"),  // Don't search above this
}

rootPath, err = pathfinder.FindRepositoryRoot("/home/user/projects/myapp/src", opts...)
if err != nil {
    log.Fatal(err)
}

// Find files with checksums (integrity verification)
query = pathfinder.FindQuery{
    Root: "/path/to/files",
    Include: []string{"**/*.tar.gz"},
    CalculateChecksums: true,
    ChecksumAlgorithm: "sha256",
}

results, err = pathfinder.Find(query)
for _, result := range results {
    fmt.Printf("%s: %s\n", result.Path, result.Checksum)
}
```

### Foundry Package

```go
import "github.com/fulmenhq/gofulmen/foundry"

// RFC3339Nano timestamps (cross-language compatible)
timestamp := foundry.UTCNowRFC3339Nano()
fmt.Println(timestamp) // "2025-10-13T14:32:15.123456789Z"

// UUIDv7 correlation IDs (time-sortable, globally unique)
correlationID := foundry.GenerateCorrelationID()
fmt.Println(correlationID) // "018b2c5e-8f4a-7890-b123-456789abcdef"

// Pattern matching from Crucible catalogs
catalog := foundry.GetDefaultCatalog()
emailPattern, _ := catalog.GetPattern("ansi-email")
if emailPattern.MustMatch("user@example.com") {
    fmt.Println("Valid email address")
}

// MIME type detection from content or extension
mimeType, _ := foundry.GetMimeTypeByExtension("json")
fmt.Printf("MIME: %s\n", mimeType.Mime) // "application/json"

data := []byte(`{"key": "value"}`)
detected, _ := foundry.DetectMimeType(data)
fmt.Printf("Detected: %s\n", detected.Name) // "JSON"

// HTTP status helpers
helper, _ := catalog.GetHTTPStatusHelper()
if helper.IsSuccess(200) {
    fmt.Println("Success response")
}
reason := helper.GetReasonPhrase(404) // "Not Found"

// Country code validation (Alpha2, Alpha3, Numeric)
country, _ := foundry.GetCountry("US")
fmt.Printf("%s (%s)\n", country.Name, country.Alpha3)
// "United States of America (USA)"

// Validate any ISO 3166-1 format (case-insensitive)
if foundry.ValidateCountryCode("usa") {
    fmt.Println("Valid country code")
}

// Text similarity and fuzzy matching
import "github.com/fulmenhq/gofulmen/foundry/similarity"

// v1 API (Levenshtein, still supported)
distance := similarity.Distance("kitten", "sitting") // 3
score := similarity.Score("kitten", "sitting")       // 0.5714...

// v2 API with algorithm selection
distance, _ := similarity.DistanceWithAlgorithm("kitten", "sitting", "osa")
score, _ := similarity.ScoreWithAlgorithm("kitten", "sitting", "jaro-winkler")
// Algorithms: "levenshtein", "osa", "damerau", "jaro-winkler", "substring"

// Suggest corrections for typos
candidates := []string{"config", "configure", "conform"}
opts := similarity.DefaultSuggestOptions()
suggestions := similarity.Suggest("confg", candidates, opts)
for _, s := range suggestions {
    fmt.Printf("%s (%.0f%% match)\n", s.Value, s.Score*100)
}
// Output: config (83% match)

// Unicode-aware normalization
normalized := similarity.Normalize("  Café  ", similarity.NormalizeOptions{
    StripAccents: true,
}) // "cafe"

// Enable opt-in telemetry
similarity.EnableTelemetry(telemetrySystem)
```

### Fulencode Package

```go
import "github.com/fulmenhq/gofulmen/fulencode"

// Encode bytes to base64
data := []byte("Hello, World!")
result, err := fulencode.Encode(data, fulencode.BASE64, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Data) // "SGVsbG8sIFdvcmxkIQ=="

// Decode with security limits
decoded, err := fulencode.DecodeString("SGVsbG8sIFdvcmxkIQ==", fulencode.BASE64, &fulencode.DecodeOptions{
    MaxDecodedSize:    100 * 1024 * 1024,  // 100MB limit
    MaxExpansionRatio: 10.0,                // Encoding bomb protection
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(decoded.Data)) // "Hello, World!"

// Detect encoding with confidence
unknownBytes := []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'}
detection, _ := fulencode.Detect(unknownBytes, nil)
fmt.Printf("Encoding: %s (%.0f%% confidence)\n", *detection.Encoding, detection.Confidence*100)
// Output: Encoding: utf-8 (100% confidence)

// BOM handling
clean, _ := fulencode.RemoveBOM(unknownBytes, nil)
fmt.Println(string(clean)) // "Hello"

// Normalize text safely (reject dangerous characters)
userInput := "Hello\x00World"  // Contains null byte
_, err = fulencode.Normalize(userInput, fulencode.TEXT_SAFE, nil)
if err != nil {
    fmt.Println("Rejected: control character")
}

// Unicode normalization (NFC/NFD/NFKC/NFKD)
result, _ := fulencode.Normalize("café", fulencode.NFC, nil)
fmt.Println(result.Text) // "café" (composed form)
```

### Signals Package

```go
import "github.com/fulmenhq/gofulmen/signals"

// Register graceful shutdown handlers (execute in LIFO order)
signals.OnShutdown(func(ctx context.Context) error {
    log.Println("Closing database...")
    return db.Close()
})

signals.OnShutdown(func(ctx context.Context) error {
    log.Println("Stopping workers...")
    return workers.Stop(ctx)
})

// Enable Ctrl+C double-tap (2-second window for force quit)
signals.EnableDoubleTap(signals.DoubleTapConfig{
    Window:  2 * time.Second,
    Message: "Press Ctrl+C again to force quit",
})

// Register config reload handler (SIGHUP)
signals.OnReload(func(ctx context.Context) error {
    // Validate new config
    if err := config.Validate(); err != nil {
        return err // Abort reload on validation failure
    }
    // Reload and restart
    return config.ReloadAndRestart(ctx)
})

// Start listening for signals
ctx := context.Background()
if err := signals.Listen(ctx); err != nil {
    log.Fatal(err)
}

// HTTP admin endpoint for Windows (SIGHUP fallback)
config := signals.HTTPConfig{
    TokenAuth: os.Getenv("SIGNAL_ADMIN_TOKEN"),
    RateLimit: 6,  // requests per minute
    RateBurst: 3,
}
handler := signals.NewHTTPHandler(config)
http.Handle("/admin/signal", handler)

// Example: Trigger reload on Windows via HTTP
// curl -X POST http://localhost:8080/admin/signal \
//   -H "Authorization: Bearer <token>" \
//   -d '{"signal": "SIGHUP", "reason": "config reload"}'
```

## CLI Tools

### Terminal Calibration

Calibrate your terminal for proper Unicode display:

```bash
go run ./cmd/terminal-calibrate
```

### Schema Validation Shim

Demonstrate the schema validation APIs without installing goneat:

```bash
go run ./cmd/gofulmen-schema -- schema validate \
  --schema-id pathfinder/v1.0.0/path-result ./path-result.json

go run ./cmd/gofulmen-schema -- schema validate-schema ./schema.json

# Optional goneat integration
go run ./cmd/gofulmen-schema -- schema validate \
  --use-goneat --schema-id pathfinder/v1.0.0/path-result ./path-result.json
```

### Bootstrap

Install external tools (trust anchor `sfetch`, then `goneat`) and foundation toolchain:

```bash
# Installs sfetch (if missing), then installs goneat (minisign required),
# then installs foundation tools via goneat doctor.
make bootstrap
```

## Development

### Running Tests

```bash
make test
```

### Running Quality Gates

```bash
make check-all
```

### Building CLI Tools

```bash
go build ./cmd/terminal-calibrate
go build ./cmd/bootstrap
```

### Developer Experience with Goneat

Gofulmen uses [Goneat](https://github.com/fulmenhq/goneat) for standardized DX operations:

```bash
# Bootstrap tools
make bootstrap

# Version management
make version-bump TYPE=patch

# Sync assets from Crucible
make sync
```

See [GONEAT.md](docs/GONEAT.md) for development tooling guide.

### Releases

gofulmen releases are **GPG-signed annotated git tags** (`vX.Y.Z`). See `RELEASE_CHECKLIST.md`.

## Documentation

### For Library Consumers

- **[Integration Guide](docs/INTEGRATION.md)** - Start here to integrate gofulmen into your application
- **Package Documentation**:
  - [Logging](logging/README.md) - Structured logging
  - [Pathfinder](pathfinder/README.md) - Safe filesystem discovery
  - [Config](config/README.md) - Configuration management
  - [ASCII](ascii/README.md) - Terminal utilities
  - [Crucible](crucible/README.md) - Schema and doc access

### For Contributors

- **[Goneat Guide](docs/GONEAT.md)** - Development tooling and workflows
- **[Bootstrap Strategy](ops/bootstrap-strategy.md)** - Bootstrap architecture
- **[Operations Docs](ops/)** - ADRs, decisions, runbooks

## Integration with Crucible

Gofulmen provides unified access to Crucible schemas and standards through the `crucible/` package. All schemas and documentation are embedded in the library - no external dependencies required.

```go
import "github.com/fulmenhq/gofulmen/crucible"

// Access version info
fmt.Println(crucible.GetVersionString())
// Output: gofulmen/0.3.5 crucible/0.4.12
```

## Supply Chain & Security

gofulmen is designed for environments where dependency hygiene matters.

**Dependency Transparency:**

- **Minimal by default**: Core packages have minimal external dependencies
- **Auditable**: Run `go mod graph` to inspect the full dependency graph
- **SBOM-ready**: Compatible with `cyclonedx-gomod` and standard Go tooling
- **License-clean**: All dependencies use MIT, Apache-2.0, or compatible licenses

**Embedded Data:**

- All Crucible catalogs (country codes, exit codes, HTTP statuses) are embedded at compile time
- No runtime network calls for reference data
- Version and provenance tracked in `.crucible/metadata/metadata.yaml`

**Security Practices:**

- Context-aware APIs with proper timeout handling
- Pattern matching uses bounded execution (no ReDoS vulnerabilities)
- Vulnerability scanning via `govulncheck`

**Audit Commands:**

```bash
# View dependency graph
go mod graph

# Check for known vulnerabilities
govulncheck ./...

# Generate SBOM
cyclonedx-gomod mod -output sbom.json
```

See [SECURITY.md](SECURITY.md) for vulnerability reporting and our full security policy.

## Contributing

Contributions are welcome! Please ensure:

- Code follows Go standards and conventions
- Tests are included for new functionality
- Documentation is updated
- Changes are consistent with Crucible standards

See [GONEAT.md](docs/GONEAT.md) for development setup, [MAINTAINERS.md](MAINTAINERS.md) for governance, and [SECURITY.md](SECURITY.md) for vulnerability reporting.

## License

Licensed under the MIT License. See LICENSE file for details.

**Trademarks**: "Fulmen" and "3 Leaps" are trademarks of 3 Leaps, LLC. While code is open source, please use distinct names for derivative works to prevent confusion.

## Changelog

See CHANGELOG.md for version history.

---

<div align="center">

**Built by the [3 Leaps](https://3leaps.net) team**

Part of the [Fulmen Ecosystem](https://github.com/fulmenhq) — Enterprise-grade libraries that thrive on scale

</div>
