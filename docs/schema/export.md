# Schema Export

Export Crucible schemas with provenance metadata for vendoring and distribution.

## Overview

The `schema/export` package provides functionality to export schemas from the embedded Crucible SSOT with:

- Provenance metadata tracking (Crucible version, gofulmen version, export timestamp, git revision)
- Multiple output formats (JSON, YAML)
- Validation before export
- Overwrite protection
- Flexible provenance styles

## Quick Start

### API Usage

```go
package main

import (
    "context"
    "github.com/fulmenhq/gofulmen/schema/export"
)

func main() {
    ctx := context.Background()

    opts := export.NewExportOptions(
        "observability/logging/v1.0.0/log-event.schema.json",
        "vendor/crucible/schemas/logging-event.schema.json",
    )

    if err := export.Export(ctx, opts); err != nil {
        panic(err)
    }
}
```

### CLI Usage

```bash
# Export logging schema as JSON
gofulmen-export-schema \
    --schema-id=observability/logging/v1.0.0/log-event.schema.json \
    --out=vendor/crucible/schemas/logging-event.schema.json

# Export as YAML with comment-style provenance
gofulmen-export-schema \
    --schema-id=terminal/v1.0.0/schema.json \
    --out=schema.yaml \
    --format=yaml \
    --provenance-style=comment

# Export without provenance
gofulmen-export-schema \
    --schema-id=terminal/v1.0.0/schema.json \
    --out=schema.json \
    --no-provenance
```

## Export Options

### API Options

```go
type ExportOptions struct {
    SchemaID         string              // REQUIRED: Schema identifier
    OutPath          string              // REQUIRED: Output file path
    Format           Format              // Output format (default: auto-detect)
    IncludeProvenance bool               // Include provenance (default: true)
    ProvenanceStyle  ProvenanceStyle     // Provenance style (default: object)
    ValidateSchema   bool                // Validate before export (default: true)
    Overwrite        bool                // Allow overwriting files (default: false)
    IdentityProvider IdentityProvider    // Optional identity provider
}
```

### CLI Flags

- `--schema-id` (required): Crucible schema identifier
- `--out` (required): Output file path
- `--format`: Output format (`json` or `yaml`, default: auto-detect from extension)
- `--provenance-style`: Provenance style (`object`, `comment`, or `none`)
- `--no-provenance`: Disable provenance metadata
- `--no-validate`: Skip schema validation
- `--force`: Overwrite existing files
- `--help`: Show help message

## Provenance Formats

### Object Style (Default - JSON)

```json
{
  "x-crucible-source": {
    "schema_id": "terminal/v1.0.0/schema.json",
    "crucible_version": "2025.10.5",
    "gofulmen_version": "0.1.8",
    "git_revision": "a4bc94f",
    "exported_at": "2025-11-03T15:20:00Z"
  },
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  ...
}
```

### Comment Style (JSON)

```json
{
  "$comment": "x-crucible-source: schema_id=terminal/v1.0.0/schema.json crucible=2025.10.5 gofulmen=0.1.8...",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  ...
}
```

### YAML Front-Matter

```yaml
# x-crucible-source:
#   schema_id: terminal/v1.0.0/schema.json
#   crucible_version: 2025.10.5
#   gofulmen_version: 0.1.8
#   git_revision: a4bc94f
#   exported_at: 2025-11-03T15:20:00Z
---
$schema: https://json-schema.org/draft/2020-12/schema
```

## Advanced Features

### Custom Identity Provider

```go
type myIdentityProvider struct{}

func (p *myIdentityProvider) GetIdentity(ctx context.Context) (*export.Identity, error) {
    return &export.Identity{
        Vendor: "mycompany",
        Binary: "myapp",
    }, nil
}

opts := export.NewExportOptions(schemaID, outPath)
opts.IdentityProvider = &myIdentityProvider{}
```

### Validation

By default, schemas are validated before export. Disable with:

```go
opts.ValidateSchema = false
```

Or via CLI:

```bash
gofulmen-export-schema --schema-id=... --out=... --no-validate
```

### Format Detection

The export automatically detects the output format from the file extension:

- `.json` → JSON format
- `.yaml`, `.yml` → YAML format
- Other → JSON format (default)

Override with:

```go
opts.Format = export.FormatYAML
```

## Exit Codes (CLI)

- `0` - Success
- `40` - Invalid arguments (ExitInvalidArgument)
- `54` - File write error (ExitFileWriteError)
- `60` - Schema validation error (ExitDataInvalid)

## Examples

### Export Multiple Schemas

```go
schemas := []string{
    "observability/logging/v1.0.0/log-event.schema.json",
    "observability/logging/v1.0.0/config.schema.json",
    "terminal/v1.0.0/schema.json",
}

for _, schemaID := range schemas {
    outPath := filepath.Join("vendor/crucible/schemas", filepath.Base(schemaID))
    opts := export.NewExportOptions(schemaID, outPath)

    if err := export.Export(ctx, opts); err != nil {
        log.Printf("Failed to export %s: %v", schemaID, err)
    }
}
```

### Export with Overwrite Protection

```go
opts := export.NewExportOptions(schemaID, outPath)
opts.Overwrite = false // default

err := export.Export(ctx, opts)
if errors.Is(err, export.ErrFileExists) {
    // Handle existing file
    fmt.Println("File already exists. Use Overwrite option to replace.")
}
```

## See Also

- [Schema Validation](../schema/validation.md)
- [Crucible Integration](../crucible/integration.md)
- [API Reference](https://pkg.go.dev/github.com/fulmenhq/gofulmen/schema/export)
