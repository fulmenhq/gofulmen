# Schema Package

Gofulmen's `schema` package provides catalog-backed JSON Schema helpers aligned with
Crucible standards.

## Features

- Offline schema catalog discovery (`ListSchemas`, `GetSchema`, `CompareSchema`).
- Validation helpers for data and schema definitions with structured diagnostics.
- Composition utilities (`MergeJSONSchemas`) and drift diffing (`DiffSchemas`).
- Minimal CLI shim (`cmd/gofulmen-schema`) for demonstration/testing.

## Quick Start

```go
catalog := schema.DefaultCatalog()
diags, err := catalog.ValidateDataByID("pathfinder/v1.0.0/path-result", payload)
if err != nil {
    log.Fatal(err)
}
for _, d := range diags {
    fmt.Printf("%s (%s): %s\n", d.Pointer, d.Keyword, d.Message)
}
```

## CLI Shim

```
go run ./cmd/gofulmen-schema -- schema validate \
  --schema-id pathfinder/v1.0.0/path-result sample.json

# Optional goneat integration
go run ./cmd/gofulmen-schema -- schema validate \
  --use-goneat --schema-id pathfinder/v1.0.0/path-result sample.json
```

The CLI defaults to the library-backed validator. Pass `--use-goneat` (or set
`GOFULMEN_GONEAT_PATH`) to shell out to `goneat` when installed.

## Composition & Drift

```go
merged, _ := schema.MergeJSONSchemas(baseBytes, overlayBytes)
diffs, _ := schema.DiffSchemas(baseBytes, merged)
```

Merged schemas and diffs are emitted as canonical JSON bytes for downstream use.
