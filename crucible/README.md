# Crucible Shim

Gofulmen's `crucible` package provides a unified facade for accessing Crucible schemas, standards, and observability artifacts. This shim allows downstream services to depend on a single Go module instead of importing both gofulmen and crucible separately.

> **Related documentation**: For a high-level explanation of available assets, start with `crucible/docs/guides/consuming-crucible-assets.md` in the Crucible repository, then use the APIs below to fetch the same artifacts directly from Go.

## Purpose

The crucible shim addresses the multi-import problem for Go applications:

- **Single Import**: Access crucible through gofulmen without separate crucible import
- **Schema Access**: Retrieve JSON schemas for validation and documentation
- **Standards Access**: Get coding standards, observability definitions
- **Version Tracking**: Track both gofulmen and crucible versions
- **Helper Functions**: Convenience methods for common schema operations

## Key Features

- **Complete Crucible API**: Full re-export of crucible.SchemaRegistry and crucible.StandardsRegistry
- **Convenience Helpers**: `GetLoggingEventSchema()`, `GetPathfinderFindQuerySchema()`, etc.
- **Version Diagnostics**: Expose both gofulmen and crucible versions
- **Schema Validation**: `ValidateAgainstSchema()` helper using gofulmen validator
- **Standards Access**: Direct access to Go and TypeScript coding standards

## Quick Lookup Recipes

Most use cases follow three steps:

1. **Locate the asset** – Navigate the registry (`crucible.SchemaRegistry`) or call a helper such as `GetLoggingConfigSchema()`.
2. **Consume the bytes** – Parse JSON/YAML with the standard library or feed schemas into `github.com/fulmenhq/gofulmen/schema`.
3. **Track provenance** – Use `crucible.GetVersion()` or `GetVersionString()` when filing tickets to show which Crucible snapshot you embedded.

Example:

```go
logging, _ := crucible.SchemaRegistry.Observability().Logging().V1_0_0()
schemaBytes, _ := logging.LogEvent()
validator, _ := schema.NewValidator(schemaBytes)
diags, err := validator.ValidateBytes(payload)
```

## Basic Usage

### Accessing Schema Registry

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Access logging schemas
    logging, err := crucible.SchemaRegistry.Observability().Logging().V1_0_0()
    if err != nil {
        log.Fatal(err)
    }

    // Get log event schema
    eventSchema, err := logging.LogEvent()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Log event schema size: %d bytes\n", len(eventSchema))
}
```

### Using Convenience Helpers

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get logging config schema directly
    configSchema, err := crucible.GetLoggingConfigSchema()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Logger config schema: %d bytes\n", len(configSchema))

    // Get pathfinder find query schema
    findQuerySchema, err := crucible.GetPathfinderFindQuerySchema()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Find query schema: %d bytes\n", len(findQuerySchema))
}
```

### Version Information

```go
package main

import (
    "fmt"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get version struct
    v := crucible.GetVersion()
    fmt.Printf("Gofulmen: %s\n", v.Gofulmen)
    fmt.Printf("Crucible: %s\n", v.Crucible)

    // Get version string
    fmt.Println(crucible.GetVersionString())
    // Output: gofulmen/0.1.0 crucible/2025.10.0
}
```

### Accessing Standards

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get Go coding standards
    goStandards, err := crucible.GetGoStandards()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Go Standards:\n%s\n", goStandards)

    // Get TypeScript standards
    tsStandards, err := crucible.GetTypeScriptStandards()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("TypeScript Standards:\n%s\n", tsStandards)
}
```

### Loading Schema Versions

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Load logging schemas
    logging, err := crucible.LoadLoggingSchemas()
    if err != nil {
        log.Fatal(err)
    }

    // Access all logging schemas
    definitions, _ := logging.Definitions()
    logEvent, _ := logging.LogEvent()
    loggerConfig, _ := logging.LoggerConfig()
    middleware, _ := logging.MiddlewareConfig()

    fmt.Printf("Loaded %d logging schemas\n", 4)

    // Load pathfinder schemas
    pathfinder, err := crucible.LoadPathfinderSchemas()
    if err != nil {
        log.Fatal(err)
    }

    findQuery, _ := pathfinder.FindQuery()
    finderConfig, _ := pathfinder.FinderConfig()
    pathResult, _ := pathfinder.PathResult()

    fmt.Printf("Loaded %d pathfinder schemas\n", 3)
}
```

### Schema Discovery

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // List schemas in a directory
    schemas, err := crucible.ListSchemas("observability/logging/v1.0.0")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d schemas:\n", len(schemas))
    for _, name := range schemas {
        fmt.Printf("  - %s\n", name)
    }

    // Get schema by path
    schema, err := crucible.GetSchema("terminal/v1.0.0/schema.json")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Terminal schema: %d bytes\n", len(schema))
}
```

### Terminal Catalogs

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get terminal schema
    terminalSchema, err := crucible.GetTerminalSchema()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Terminal schema: %d bytes\n", len(terminalSchema))

    // Get terminal catalog (all terminal profiles)
    catalog, err := crucible.GetTerminalCatalog()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Terminal catalog has %d profiles:\n", len(catalog))
    for name, data := range catalog {
        fmt.Printf("  - %s: %d bytes\n", name, len(data))
    }
}
```

### Parsing Schemas

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get schema as bytes
    schemaData, err := crucible.GetLoggingEventSchema()
    if err != nil {
        log.Fatal(err)
    }

    // Parse to map
    parsed, err := crucible.ParseJSONSchema(schemaData)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Schema title: %v\n", parsed["$id"])
    fmt.Printf("Schema type: %v\n", parsed["type"])
    fmt.Printf("Required fields: %v\n", parsed["required"])
}
```

## API Reference

### Version Information

#### crucible.GetVersion() Version

Returns version information for both gofulmen and crucible.

**Returns:**

- `Version`: Struct with Gofulmen and Crucible version strings

#### crucible.GetVersionString() string

Returns formatted version string.

**Returns:**

- `string`: Format "gofulmen/{version} crucible/{version}"

### Registry Access

#### crucible.SchemaRegistry

Global schema registry providing access to all crucible schemas.

**Methods:**

- `Terminal() *TerminalSchemas`
- `Pathfinder() *PathfinderSchemas`
- `ASCII() *ASCIISchemas`
- `SchemaValidation() *SchemaValidationSchemas`
- `Observability() *ObservabilitySchemas`

#### crucible.StandardsRegistry

Global standards registry providing access to coding standards.

**Methods:**

- `Coding() *CodingStandards`

### Schema Loaders

#### crucible.LoadLoggingSchemas() (\*LoggingSchemasV1, error)

Loads logging schema collection.

**Returns:**

- `*LoggingSchemasV1`: Logging schemas (Definitions, LogEvent, LoggerConfig, MiddlewareConfig)
- `error`: Load error if any

#### crucible.LoadPathfinderSchemas() (\*PathfinderSchemasV1, error)

Loads pathfinder schema collection.

**Returns:**

- `*PathfinderSchemasV1`: Pathfinder schemas (FindQuery, FinderConfig, PathResult, ErrorResponse, Metadata)
- `error`: Load error if any

#### crucible.LoadSchemaValidationSchemas() (\*SchemaValidationSchemasV1, error)

Loads schema validation schema collection.

**Returns:**

- `*SchemaValidationSchemasV1`: Schema validation schemas
- `error`: Load error if any

### Convenience Helpers

#### crucible.GetLoggingEventSchema() ([]byte, error)

Gets log event schema bytes.

#### crucible.GetLoggingConfigSchema() ([]byte, error)

Gets logger config schema bytes.

#### crucible.GetPathfinderFindQuerySchema() ([]byte, error)

Gets pathfinder find query schema bytes.

#### crucible.GetPathfinderConfigSchema() ([]byte, error)

Gets pathfinder config schema bytes.

#### crucible.GetGoStandards() (string, error)

Gets Go coding standards document.

#### crucible.GetTypeScriptStandards() (string, error)

Gets TypeScript coding standards document.

#### crucible.GetTerminalSchema() ([]byte, error)

Gets terminal schema bytes.

#### crucible.GetTerminalCatalog() (map[string][]byte, error)

Gets all terminal profile catalogs.

**Returns:**

- `map[string][]byte`: Map of filename to schema bytes
- `error`: Load error if any

#### crucible.GetASCIIStringAnalysisSchema() ([]byte, error)

Gets ASCII string analysis schema bytes.

#### crucible.GetASCIIBoxCharsSchema() ([]byte, error)

Gets ASCII box chars schema bytes.

### Generic Access

#### crucible.GetSchema(schemaPath string) ([]byte, error)

Gets schema by relative path.

**Parameters:**

- `schemaPath`: Path relative to schemas directory (e.g., "terminal/v1.0.0/schema.json")

**Returns:**

- `[]byte`: Schema bytes
- `error`: Load error if any

#### crucible.GetDoc(docPath string) (string, error)

Gets documentation by relative path.

**Parameters:**

- `docPath`: Path relative to docs directory

**Returns:**

- `string`: Document content
- `error`: Load error if any

#### crucible.ListSchemas(basePath string) ([]string, error)

Lists schemas in a directory.

**Parameters:**

- `basePath`: Base path relative to schemas directory

**Returns:**

- `[]string`: Slice of schema filenames
- `error`: List error if any

#### crucible.ParseJSONSchema(data []byte) (map[string]any, error)

Parses JSON schema bytes to map.

**Parameters:**

- `data`: Schema JSON bytes

**Returns:**

- `map[string]any`: Parsed schema
- `error`: Parse error if any

### Validation

#### crucible.ValidateAgainstSchema(schemaData []byte, jsonData []byte) error

Validates JSON data against a schema using gofulmen's validator.

**Parameters:**

- `schemaData`: JSON schema bytes
- `jsonData`: JSON data to validate

**Returns:**

- `error`: Validation error if invalid

**Note**: Requires proper schema resolution setup. For production use, see the logging package's `ValidateConfig()` for a complete example.

## Available Schemas

### Observability Schemas

**Logging (v1.0.0):**

- `definitions.schema.json` - Shared type definitions
- `log-event.schema.json` - Log event structure
- `logger-config.schema.json` - Logger configuration
- `middleware-config.schema.json` - Middleware configuration

### Pathfinder Schemas (v1.0.0)

- `find-query.schema.json` - File discovery query
- `finder-config.schema.json` - Finder configuration
- `path-result.schema.json` - Path discovery result
- `error-response.schema.json` - Error response
- `metadata.schema.json` - File metadata

### Terminal Schemas (v1.0.0)

- `schema.json` - Terminal configuration
- `catalog/*.json` - Terminal profile catalogs

### ASCII Schemas (v1.0.0)

- `string-analysis.schema.json` - String analysis result
- `box-chars.schema.json` - Box drawing characters

### Schema Validation (v1.0.0)

- `validator-config.schema.json` - Validator configuration
- `schema-registry.schema.json` - Schema registry

## Standards Documents

### Coding Standards

- `docs/standards/coding/go.md` - Go coding standards
- `docs/standards/coding/typescript.md` - TypeScript coding standards

## Usage Pattern

**For Application Code (e.g., Fulward):**

```go
import (
    "github.com/fulmenhq/gofulmen/logging"
    "github.com/fulmenhq/gofulmen/pathfinder"
    "github.com/fulmenhq/gofulmen/crucible"
)
```

**Single import path**, no need to separately import `github.com/fulmenhq/crucible`.

## Integration Example

```go
package main

import (
    "log"

    "github.com/fulmenhq/gofulmen/crucible"
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // Show versions
    log.Println(crucible.GetVersionString())

    // Create logger using crucible schemas (via logging package)
    config := logging.DefaultConfig("my-app")
    logger, err := logging.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    // Get schemas for documentation
    eventSchema, _ := crucible.GetLoggingEventSchema()
    log.Printf("Using log event schema: %d bytes", len(eventSchema))

    logger.Info("Application started with crucible schemas")
}
```

## Testing

```bash
go test ./crucible/...
```

All tests verify:

- Version information
- Schema registry access
- Schema loading functions
- Convenience helpers
- Standards access
- Schema discovery

## Architecture

**Design Pattern: Facade**

The crucible package is a thin facade that:

1. Re-exports entire crucible API via type aliases
2. Provides convenience helpers for common operations
3. Tracks both gofulmen and crucible versions
4. Avoids circular dependencies

**Why This Pattern:**

- Fulward can import `gofulmen` only (single dependency)
- Crucible remains SSOT for schemas/standards
- Gofulmen provides ergonomic Go API layer
- No code duplication, just re-export

## Version Compatibility

- **Gofulmen Version**: 0.1.0
- **Crucible Version**: 2025.10.0

Both versions are exposed via `GetVersion()` for diagnostics and compatibility tracking.

## Future Enhancements

- Schema caching layer
- Version negotiation helpers
- Schema diff utilities
- Automatic schema updates
- CLI tool for schema inspection
