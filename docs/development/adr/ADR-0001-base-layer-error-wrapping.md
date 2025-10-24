---
id: "ADR-0001"
title: "Base Layer Packages Use Standard Go Errors, Callers Wrap with Envelopes"
status: "accepted"
date: "2025-10-24"
deciders: ["@3leapsdave", "Foundation Forge"]
scope: "gofulmen"
tags: ["error-handling", "architecture", "import-cycles", "telemetry"]
---

# ADR-0001: Base Layer Packages Use Standard Go Errors, Callers Wrap with Envelopes

## Context

During the v0.1.5 error and telemetry retrofit (Phases 1-3), we attempted to add structured error envelopes (`*errors.ErrorEnvelope`) and telemetry metrics emission to the `schema` package validation functions. This created an import cycle:

```
schema → errors (for ErrorEnvelope)
errors → schema (from errors/fixtures_test.go for testing)
```

Go's strict import cycle detection prevents this dependency structure. Similar constraints exist with telemetry:

```
schema → telemetry (for metrics emission)
telemetry → schema (for validator configuration)
config → telemetry (for metrics)
config → schema (for validation)
```

### Package Dependency Layers

**Base Layer** (no dependencies on errors/telemetry):

- `schema` - JSON Schema validation primitives
- `fulhash` - Hashing algorithms
- `foundry` - Static data catalogs

**Mid Layer** (imports base + errors):

- `errors` - Error envelope definitions
- `pathfinder` - File discovery (imports errors)

**Top Layer** (imports all):

- `config` - Configuration loading (imports schema, errors, telemetry)
- `telemetry` - Metrics system (imports schema for validation)
- `logging` - Structured logging

### User Perspective Analysis

**Schema Library External Users:**

- Expect clean, simple API with standard Go errors
- Want diagnostics separate from errors (`[]Diagnostic` + `error`)
- Don't need forced error envelope format
- May have their own error wrapping strategies

**Internal Library Users (config, pathfinder):**

- Already wrap schema errors with their own context
- Already emit domain-specific metrics (`config_load_errors`, `pathfinder_validation_errors`)
- Control correlation ID propagation at their layer
- Provide consistent user-facing error codes (`CONFIG_*`, `PATHFINDER_*`)

## Decision

**Base layer packages (schema, fulhash, foundry) will:**

1. Return standard Go errors (using `fmt.Errorf`, `errors.New`)
2. Return domain-specific diagnostics as separate values (e.g., `[]Diagnostic`)
3. NOT import `errors` package (to avoid cycles)
4. NOT import `telemetry` package (to avoid cycles)
5. NOT emit telemetry metrics directly

**Calling packages (config, pathfinder, logging) will:**

1. Wrap base layer errors with `*errors.ErrorEnvelope` at their boundaries
2. Add their own context (component, operation, error_type)
3. Emit their own domain-specific metrics
4. Control correlation ID propagation
5. Preserve original errors using `envelope.WithOriginal(err)`

**Error code naming convention:**

- Config errors: `CONFIG_*` (e.g., `CONFIG_VALIDATION_ERROR`)
- Pathfinder errors: `PATHFINDER_*` (e.g., `PATHFINDER_VALIDATION_ERROR`)
- Schema errors propagate as original cause in envelope.Original

**Metric naming convention:**

- Config metrics: `config_*` (e.g., `config_load_errors`)
- Pathfinder metrics: `pathfinder_*` (e.g., `pathfinder_validation_errors`)
- Schema operations are counted by the calling layer's metrics

## Rationale

### 1. Import Cycle Prevention

Go's import cycle detection is non-negotiable. Base layer packages cannot import higher-level concerns.

### 2. Maximum Reusability

Schema package stays pure and reusable by external consumers who may have different error handling strategies.

### 3. Clear Separation of Concerns

- Schema validates data (pure function)
- Calling layer adds context and telemetry (operational concern)
- User sees errors in caller's domain language

### 4. Consistent User Experience

```go
// User sees config-domain errors
merged, diags, err := config.LoadLayeredConfig(opts)
// err is *errors.ErrorEnvelope with Code="CONFIG_VALIDATION_ERROR"

// User sees pathfinder-domain errors
results, err := finder.FindFiles(ctx, query)
// err is *errors.ErrorEnvelope with Code="PATHFINDER_VALIDATION_ERROR"

// Both wrap underlying schema errors in their domain context
```

### 5. Already Implemented Correctly

Phase 1 (pathfinder) and Phase 2 (config) already follow this pattern:

```go
// pathfinder/finder.go
func validatePathResultWithTelemetry(result PathResult, correlationID string, telSys *telemetry.System) error {
    // Call schema validation
    diags, err := validator.ValidateData(result)
    if err != nil {
        // Emit pathfinder metric
        if telSys != nil {
            _ = telSys.Counter(metrics.PathfinderValidationErrors, 1, tags)
        }
        // Wrap in pathfinder envelope
        envelope := errors.NewErrorEnvelope("PATHFINDER_VALIDATION_ERROR", "...")
        envelope = envelope.WithOriginal(err)
        return envelope
    }
    // ...
}
```

## Alternatives Considered

### Alternative 1: Add \*WithEnvelope Variants to Schema

**Rejected**: Creates import cycle. Would require schema → errors dependency.

### Alternative 2: Create errors/base Package

**Rejected**: Splits error definitions across packages. Adds complexity for external users. Doesn't solve telemetry import cycle.

### Alternative 3: Refactor errors Package to Not Import Schema

**Rejected**: Breaking change. Schema validation in error tests is valuable. Doesn't solve telemetry cycles (telemetry needs schema for its own validation).

### Alternative 4: Use Interface Types to Break Cycles

**Rejected**: Over-engineered. Adds interface complexity for import cycle workaround. Standard Go pattern is to keep base packages simple.

## Consequences

### Positive

- ✅ No import cycles
- ✅ Schema package stays pure and reusable
- ✅ Clear architectural layers
- ✅ Each package emits domain-specific metrics
- ✅ Consistent error code prefixes per package
- ✅ Original errors preserved in envelope chain

### Negative

- ⚠️ Callers must remember to wrap schema errors (but they already do)
- ⚠️ No centralized schema validation metrics (but domain metrics are more useful)

### Neutral

- Schema package cannot provide \*WithEnvelope convenience variants
- Telemetry emission happens at caller layer, not validation layer
- Error wrapping is explicit rather than implicit

## Implementation

### Completed

- **Phase 1 (pathfinder)**: Added `ValidatePathResultWithEnvelope`, wraps schema errors, emits `pathfinder_validation_errors`
- **Phase 2 (config)**: `LoadLayeredConfigWithEnvelope` wraps schema errors, emits `config_load_errors`

### Phase 3 (schema)

**Status**: No changes needed. Schema correctly returns standard Go errors.

**Verified behavior:**

```go
// schema/validator.go - returns standard Go errors
func (v *Validator) ValidateData(data interface{}) ([]Diagnostic, error) {
    err := v.schema.Validate(data)
    if err == nil {
        return nil, nil
    }
    validationErr, ok := err.(*jsonschema.ValidationError)
    if !ok {
        return nil, err  // Standard Go error
    }
    return diagnosticsFromValidationError(validationErr, sourceGoFulmen), nil
}

// config/layered.go - wraps schema errors
diags, err := catalog.ValidateDataByID(opts.SchemaID, payload)
if err != nil {
    envelope := errors.NewErrorEnvelope("CONFIG_VALIDATION_ERROR", "...")
    envelope = envelope.WithOriginal(err)  // Preserves schema error
    _ = telSys.Counter(metrics.ConfigLoadErrors, 1, tags)
    return nil, diags, envelope
}
```

### Files Modified

- None (existing architecture already correct)

### Tests

- Existing tests in pathfinder and config verify wrapping behavior
- No schema-level envelope tests needed (would create import cycle)

## Related Ecosystem ADRs

None. This is a Go-specific decision driven by Go's import cycle constraints. Other languages (Python, TypeScript) don't have this limitation and may structure their error/telemetry differently.

## References

- [Phase 1-3 Implementation Plan](.plans/active/v0.1.5/error-telemetry-retrofit-implementation.md)
- [Error Telemetry Dogfood Guide](../error-telemetry-dogfood.md)
- [Go Import Cycles](https://golang.org/ref/spec#Import_declarations)
- [Effective Go: Errors](https://go.dev/doc/effective_go#errors)

---

**Decision Outcome**: Base layer packages (schema, fulhash, foundry) return standard Go errors. Calling packages (config, pathfinder, logging) wrap with error envelopes and emit domain-specific telemetry metrics. This preserves architectural layering, prevents import cycles, and provides consistent user-facing error semantics.
