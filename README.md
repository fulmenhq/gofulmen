# Gofulmen

**Curated Libraries for Scale**

Gofulmen is a collection of Go packages that provide consistent, high-quality implementations of common functionality across the FulmenHQ ecosystem. Built on top of Crucible's schemas and standards, these libraries ensure uniformity and reliability.

📖 **[Read the complete library overview](docs/gofulmen_overview.md)** for comprehensive documentation including module catalog, dependency map, and roadmap.

## Packages

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

All Foundry catalogs are embedded at compile time and work offline - no network dependencies required.

## Installation

```bash
go get github.com/fulmenhq/gofulmen
```

## Usage

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
err = validator.Validate("hello")
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

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
```

## CLI Tools

### Terminal Calibration

Calibrate your terminal for proper Unicode display:

```bash
go run ./cmd/terminal-calibrate
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
// Output: gofulmen/0.1.0 crucible/2025.10.0
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
