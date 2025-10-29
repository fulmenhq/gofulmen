# ADR-0004: Crucible Runtime Dependency Pattern

**Status**: Accepted  
**Date**: 2025-10-28  
**Deciders**: @3leapsdave, Foundation Forge  
**Context**: forge-workhorse-groningen integration, gofulmen installation requirements

## Context and Problem Statement

During initial integration of gofulmen into forge-workhorse-groningen, we discovered that `go get github.com/fulmenhq/gofulmen` fails because gofulmen has a runtime dependency on `github.com/fulmenhq/crucible`, which was private at the time. This raised the question: Why does gofulmen have a hard runtime dependency on Crucible (an infoarch SSOT repository) rather than using a soft sync pattern via goneat like we do for documentation and schemas?

## Decision Drivers

- **Single Import Convenience**: Applications should import only `github.com/fulmenhq/gofulmen`, not both gofulmen and crucible
- **Runtime Schema Access**: Some gofulmen features require dynamic schema access at runtime
- **Version Tracking**: Need to track both gofulmen and crucible versions for compatibility
- **SSOT Principle**: Crucible remains the single source of truth for schemas and standards
- **Go Module Semantics**: Go's import system requires resolvable dependencies

## Considered Options

### Option 1: Runtime Dependency (Shim Pattern) - CHOSEN

Gofulmen includes a `crucible/` package that acts as a facade/shim, re-exporting the entire Crucible Go module.

**Pros**:
- Clean single-import API for applications (`import "github.com/fulmenhq/gofulmen/logging"`)
- Runtime schema access for validation and introspection
- Version diagnostics expose both gofulmen and crucible versions
- No code duplication - pure re-export
- Enables dynamic features (schema inspection, version negotiation)

**Cons**:
- Crucible must be publicly accessible for external users
- Dependency coupling between gofulmen and crucible versions
- Cannot use gofulmen if crucible is unavailable

### Option 2: Embedded Schemas (Soft Sync Pattern)

Embed Crucible schemas directly in gofulmen via `go:embed` and goneat sync, making crucible an optional dependency.

**Pros**:
- No runtime dependency on crucible repository
- Works in air-gapped environments
- Faster startup (no external schema loading)
- Simpler dependency graph

**Cons**:
- Schema duplication between crucible and gofulmen
- No dynamic schema inspection capabilities
- Version drift risk if sync not run regularly
- Larger gofulmen binary size
- Cannot access new schemas without gofulmen update

### Option 3: Conditional Dependency (Build Tags)

Support both patterns via build tags: `-tags embedded` for embedded mode, default uses runtime dependency.

**Pros**:
- Flexibility for different deployment scenarios
- Air-gap support when needed
- Runtime features available when desired

**Cons**:
- Increased maintenance burden (two code paths)
- Testing complexity (must test both modes)
- Documentation complexity
- Potential feature parity issues

## Decision Outcome

**Chosen option**: Option 1 - Runtime Dependency (Shim Pattern)

### Rationale

1. **Architectural Intent**: The `crucible/` shim package was intentionally designed to provide "single import" convenience for applications, as documented in `crucible/README.md`

2. **Runtime Features**: Several gofulmen features genuinely benefit from runtime schema access:
   - Schema validation with dynamic schema selection
   - Version compatibility checking
   - Documentation generation
   - Runtime introspection and diagnostics

3. **Crucible Publicity**: Crucible is transitioning to a public repository, making this a non-issue for external users

4. **Simplicity**: Pure re-export pattern avoids code duplication and synchronization issues

5. **Ecosystem Consistency**: This pattern allows all FulmenHQ Go libraries to share schema access patterns

### Implementation Details

**Current Structure**:

```go
// crucible/crucible.go
package crucible

import "github.com/fulmenhq/crucible"

// Re-export entire Crucible API
var (
    SchemaRegistry    = crucible.SchemaRegistry
    StandardsRegistry = crucible.StandardsRegistry
)

type Schemas = crucible.Schemas
// ... more type aliases

// Convenience helpers
func GetLoggingEventSchema() ([]byte, error) {
    logging, err := SchemaRegistry.Observability().Logging().V1_0_0()
    if err != nil {
        return nil, err
    }
    return logging.LogEvent()
}
```

**go.mod Dependency**:

```go
require (
    github.com/fulmenhq/crucible v2025.10.0
    // ... other dependencies
)

// Development only: use local crucible
replace github.com/fulmenhq/crucible => ../crucible/lang/go
```

**Usage in Application Code**:

```go
import (
    "github.com/fulmenhq/gofulmen/logging"
    "github.com/fulmenhq/gofulmen/crucible"
)

// Single import path - no separate crucible import needed
func main() {
    log.Println(crucible.GetVersionString())
    logger, _ := logging.New(logging.DefaultConfig("app"))
}
```

## Consequences

### Positive

- ✅ **Clean API**: Applications import only `github.com/fulmenhq/gofulmen`
- ✅ **Runtime Features**: Full schema inspection and validation capabilities
- ✅ **Version Tracking**: `crucible.GetVersion()` exposes both versions
- ✅ **SSOT Maintained**: Crucible remains authoritative source
- ✅ **No Duplication**: Pure re-export avoids schema drift

### Negative

- ⚠️ **Public Dependency**: Crucible must be publicly accessible (resolved when Crucible goes public)
- ⚠️ **Version Coupling**: Gofulmen releases may need coordination with crucible updates
- ⚠️ **Network Access**: Installation requires network access to fetch crucible

### Neutral

- 🔄 **Local Development**: `replace` directive works for local development
- 🔄 **Testing**: Requires crucible to be available in CI/CD

## Comparison with PyFulmen and TSFulmen

**PyFulmen**: 
- Likely uses `package_data` or `importlib.resources` to bundle schemas
- Python's dynamic nature makes runtime schema access easier without hard dependencies
- May use optional dependency pattern: `pip install pyfulmen[schemas]`

**TSFulmen**:
- TypeScript bundles schemas at build time via webpack/rollup
- Type definitions can be generated from schemas
- No runtime dependency needed since schemas are static assets

**Key Difference**: Go's strict compile-time module resolution requires explicit dependencies, unlike Python's runtime imports or TypeScript's build-time bundling.

## Installation Documentation

### For External Users (Crucible Public)

```bash
# Standard installation - works once Crucible is public
go get github.com/fulmenhq/gofulmen@latest
```

### For Internal Users (Crucible Private)

```bash
# 1. Clone crucible repository
git clone git@github.com:fulmenhq/crucible.git

# 2. Clone gofulmen repository 
git clone git@github.com:fulmenhq/gofulmen.git

# 3. Use local replace in go.mod (already configured)
# The replace directive in go.mod points to ../crucible/lang/go
```

### For CI/CD

```yaml
# Ensure both repositories are accessible
- name: Checkout crucible
  uses: actions/checkout@v4
  with:
    repository: fulmenhq/crucible
    path: crucible

- name: Checkout gofulmen  
  uses: actions/checkout@v4
  with:
    repository: fulmenhq/gofulmen
    path: gofulmen
```

## Future Considerations

### Potential Enhancements (Low Priority)

1. **Embedded Mode Option**: Add build tag for embedded schemas if air-gap deployment becomes a requirement
2. **Schema Caching**: Implement local schema cache to reduce repeated access
3. **Lazy Loading**: Only load schemas when actually needed
4. **Version Negotiation**: Helper functions to check crucible version compatibility

### Monitoring

Track in future reviews:
- Installation friction reports from external users
- Network access issues in restricted environments  
- Version compatibility problems between gofulmen and crucible
- Feature requests for embedded mode

## References

- [Crucible Shim README](../../crucible/README.md) - Documents the facade pattern
- [Crucible Repository](https://github.com/fulmenhq/crucible) - SSOT for schemas
- [Go Modules Documentation](https://go.dev/doc/modules/managing-dependencies)
- Issue: forge-workhorse-groningen installation failure (2025-10-28)

## Related ADRs

- None yet - this is the first ADR documenting dependency patterns

## Update: Monorepo Subdirectory Issue (2025-10-28)

**Discovery**: After Crucible was made public (v2025.10.4), we discovered that the Go module is located at `lang/go/` subdirectory in the Crucible monorepo, not at the repository root. This creates a Go modules limitation:

**Problem**: Go modules cannot directly reference subdirectories in a repository. The module path `github.com/fulmenhq/crucible` expects the go.mod to be at the repository root, but Crucible's structure is:

```
crucible/
├── lang/
│   ├── go/           # Go module here (go.mod declares github.com/fulmenhq/crucible)
│   ├── python/
│   └── typescript/
├── schemas/
├── docs/
└── ...
```

**Current Workaround**: Using `replace` directive in go.mod:

```go
// go.mod
require github.com/fulmenhq/crucible v0.0.0-00010101000000-000000000000

replace github.com/fulmenhq/crucible => ../crucible/lang/go
```

**Impact**:
- ❌ `go get github.com/fulmenhq/gofulmen` fails for external users
- ✅ Works fine for local development with sibling directories
- ✅ Works in CI/CD with both repos checked out

**Solutions Being Considered**:

1. **Separate Go Module Repository** (Recommended):
   - Create `github.com/fulmenhq/crucible-go` repository
   - Contains only the Go module code
   - Can be versioned independently
   - Standard `go get` works immediately

2. **Move Go Module to Root**:
   - Restructure Crucible to have go.mod at root
   - Place language-specific code in subdirectories
   - Breaking change for existing structure

3. **Go Workspace** (Complex):
   - Use Go 1.18+ workspace feature
   - Requires external users to set up workspace
   - Non-standard workflow

**Temporary Status**: forge-workhorse-groningen and other external projects must:
1. Clone both `crucible` and `gofulmen` repositories as siblings
2. Use the existing `replace` directive
3. Wait for Crucible team to implement Solution 1 or 2

**Action Required**: Crucible team to decide on repository restructuring approach.

## Changelog

- **2025-10-28**: Initial version documenting crucible runtime dependency pattern
- **2025-10-28**: Updated with monorepo subdirectory issue discovery and workarounds
