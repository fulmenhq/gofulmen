# Gofulmen

**Curated Libraries for Scale**

Gofulmen is a collection of Go packages that provide consistent, high-quality implementations of common functionality across the FulmenHQ ecosystem. Built on top of Crucible's schemas and standards, these libraries ensure uniformity and reliability.

📖 **[Read the complete library overview](docs/gofulmen_overview.md)** for comprehensive documentation including module catalog, dependency map, and roadmap.

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

JSON Schema validation with support for draft 2020-12.

- Schema loading and caching
- Versioned schema registry
- Validation with detailed error reporting
- Support for YAML and JSON

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
- Local development overrides via `.goneat/tools.local.yaml`

### Foundry (`foundry/`)

Enterprise-grade foundation utilities providing consistent cross-language implementations from Crucible catalogs.

- **Time Utilities**: RFC3339Nano timestamps with nanosecond precision
- **Correlation IDs**: UUIDv7 time-sortable IDs for distributed tracing
- **Pattern Matching**: Regex, glob, and literal patterns from Crucible catalogs
- **MIME Type Detection**: Content-based detection and extension lookup
- **HTTP Status Helpers**: Status code grouping and validation
- **Country Code Validation**: ISO 3166-1 country codes (Alpha2, Alpha3, Numeric)
- **Exit Codes**: 54 standardized exit codes with metadata, platform detection, simplified mode mapping, BSD sysexits.h compatibility
- **Text Similarity** (`foundry/similarity/`): v1 API (Levenshtein) + v2 API (5 algorithms: Levenshtein, OSA, Damerau, Jaro-Winkler, Substring), normalized scoring, fuzzy matching, Unicode normalization, opt-in telemetry

All Foundry catalogs are embedded at compile time and work offline - no network dependencies required.

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

### Signals (`pkg/signals/`)

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

// Merge schemas at runtime (base + overlay)
mergedSchema, _ := schema.MergeJSONSchemas(
    []byte(`{"type":"object"}`),
    []byte(`{"properties":{"name":{"type":"string"}}}`),
)
fmt.Printf("Merged schema: %s\n", string(mergedSchema))
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

// v2 API with algorithm selection (NEW in v0.1.5)
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

// Enable opt-in telemetry (NEW in v0.1.5)
similarity.EnableTelemetry(telemetrySystem)
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

Install external tools using goneat bootstrap pattern:

```bash
# Install tools from .goneat/tools.yaml
make bootstrap

# Or run directly
go run ./cmd/bootstrap --install --verbose

# For local development, create override:
cp .goneat/tools.local.yaml.example .goneat/tools.local.yaml
# Then edit .goneat/tools.local.yaml to point to local binaries
```

## Development

### Running Tests

```bash
go test ./...
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
// Output: gofulmen/0.1.9 crucible/2025.10.0
```

## Contributing

Contributions are welcome! Please ensure:

- Code follows Go standards and conventions
- Tests are included for new functionality
- Documentation is updated
- Changes are consistent with Crucible standards

See [GONEAT.md](docs/GONEAT.md) for development setup and [ops/](ops/) for operational documentation.

## License

Licensed under the MIT License. See LICENSE file for details.

## Changelog

See CHANGELOG.md for version history.
