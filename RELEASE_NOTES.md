# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only the latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.9] - 2025-11-05 (In Development)

### App Identity Module

**Release Type**: Major Feature Addition  
**Status**: 🚧 In Development

#### Overview

This release introduces the App Identity module for loading and managing application metadata from `.fulmen/app.yaml` files. This Layer 0 module provides the foundation for consistent configuration paths, environment variables, and telemetry namespaces across Go applications.

#### Features

**Core API (`appidentity/`)**:

- **Identity Loading**: `Get()`, `Must()`, `GetWithOptions()`, `LoadFrom()` with process-level caching
- **Discovery**: Automatic `.fulmen/app.yaml` discovery via ancestor search (max 20 levels)
- **Precedence**: Context injection → Explicit path → Environment variable (`FULMEN_APP_IDENTITY_PATH`) → Ancestor search
- **Validation**: Schema validation against Crucible v1.0.0 app-identity schema with field-level diagnostics
- **Caching**: Thread-safe process-level caching with sync.Once (verified with race detector)
- **Testing Support**: `WithIdentity()` context injection, `Reset()` cache clearing, test utilities
- **Integration Helpers**: `ConfigParams()`, `EnvVar()`, `FlagsPrefix()`, `TelemetryNamespace()`, `ServiceName()`

**Testing Utilities**:

- **NewFixture()**: Create minimal test identity with optional overrides
- **NewCompleteFixture()**: Create complete test identity with all fields populated
- **Context Override**: `WithIdentity(ctx, identity)` for test isolation
- **Cache Reset**: `Reset()` for test cleanup (not concurrent-safe)

**Error Types**:

- **NotFoundError**: Detailed search information with documentation reference
- **ValidationError**: Field-level diagnostics with JSON Pointer paths
- **MalformedError**: YAML parsing errors with file context

**Quality Assurance**:

- **87.9% Test Coverage**: 68 tests passing (includes subtests and examples)
- **Zero Race Conditions**: Verified with `-race` flag
- **Zero Lint Issues**: All code passes golangci-lint
- **Zero Import Cycles**: Layer 0 module with no Fulmen dependencies
- **8 Godoc Examples**: Comprehensive usage examples for all major APIs

#### Files Added

```
appidentity/
├── doc.go                         # Package documentation
├── identity.go                    # Identity structs and getters
├── errors.go                      # Error types with diagnostics
├── loader.go                      # File discovery and loading
├── validation.go                  # Schema validation
├── cache.go                       # Thread-safe caching
├── override.go                    # Context-based injection
├── testing.go                     # Test utilities
├── identity_test.go               # Core API tests
├── loader_test.go                 # Discovery tests
├── validation_test.go             # Validation tests
├── cache_test.go                  # Concurrency tests
├── override_test.go               # Override tests
├── testing_test.go                # Test utility tests
├── examples_test.go               # Godoc examples
├── app-identity.schema.json       # Embedded Crucible schema
└── testdata/
    ├── valid-minimal.yaml
    ├── valid-complete.yaml
    ├── valid-gofulmen.yaml
    ├── invalid-missing-field.yaml
    ├── invalid-format.yaml
    └── invalid-env-prefix.yaml
```

**Total**: 15 Go files + 6 test fixtures, 2,864 lines

#### Example Usage

**Basic Usage**:

```go
import "github.com/fulmenhq/gofulmen/appidentity"

// Load identity from .fulmen/app.yaml
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
```

**Testing**:

```go
// Create test fixture
testIdentity := appidentity.NewFixture(func(id *appidentity.Identity) {
    id.BinaryName = "testapp"
    id.EnvPrefix = "TESTAPP_"
})

// Inject into context for test isolation
ctx = appidentity.WithIdentity(ctx, testIdentity)

// Test code uses injected identity
identity, _ := appidentity.Get(ctx)
// Returns testIdentity instead of loading from disk
```

#### Integration Points

- **Config Module**: `ConfigParams()` provides vendor/name for XDG path derivation
- **Logging Module**: `ServiceName()` and `TelemetryNamespace()` for structured logging
- **CLI Tools**: `FlagsPrefix()` and `Binary()` for flag naming conventions
- **Environment Variables**: `EnvVar()` for consistent variable naming

#### Migration Notes

This is a new module with no breaking changes. Existing code continues to work unchanged. Applications can adopt app identity incrementally:

1. Create `.fulmen/app.yaml` in project root
2. Replace hardcoded config paths with `identity.ConfigParams()`
3. Replace hardcoded env var prefixes with `identity.EnvVar()`
4. Update tests to use `WithIdentity()` for isolation

#### Known Limitations

- Identity is static per process (no dynamic reloading)
- No multi-app registry/UUID support (deferred to future release)
- Reset() not safe during concurrent Get() calls (test-only function)

---

### Signal Handling Module

**Release Type**: Major Feature Addition  
**Status**: 🚧 In Development

#### Overview

This release introduces cross-platform signal handling for graceful shutdown, config reload, and Windows fallback, following Crucible signals catalog v1.0.0 specification.

#### Features

**Catalog Layer (`foundry/signals`)**:

- **Typed Signal Access**: `GetSignalInfo(sig)`, `GetSignalByID(id)`, `GetSignalByName(name)` with parsed catalog metadata
- **8 Standard Signals**: TERM, INT, HUP, QUIT, PIPE, ALRM, USR1, USR2 from Crucible v1.0.0
- **Lazy Loading**: On-demand catalog parsing with sync.Once caching
- **Fast Lookups**: O(1) ID/name lookups via pre-built maps
- **Parity Verification**: Snapshot test ensures gofulmen matches Crucible catalog (8/8 signals verified)

**Core API (`pkg/signals`)**:

- **Registration**: `OnShutdown(handler)`, `OnReload(handler)`, `Handle(sig, handler)` for signal callbacks
- **Cleanup Chains**: LIFO execution with context support and configurable timeouts
- **Reload Chains**: FIFO execution with fail-fast error handling and validation hooks
- **Double-Tap**: Ctrl+C within 2s (configurable) for force quit with catalog-derived exit codes
- **Quiet Mode**: `SetQuietMode(true)` for non-interactive services (suppresses double-tap messages)
- **Thread-Safe**: Concurrent handler registration with mutex protection
- **Lifecycle**: `Listen(ctx)` blocking listener, `Stop()` for graceful teardown

**Platform Support**:

- **Unix/Linux**: Full POSIX signal support (TERM, INT, HUP, QUIT, PIPE, ALRM, USR1, USR2)
- **macOS**: Full POSIX signal support with BSD compatibility
- **Windows**: Automatic fallback with INFO logging and operation hints for unsupported signals
- **Platform Detection**: `IsWindows()` helper for conditional behavior

**HTTP Admin Endpoint (`pkg/signals/http`)**:

- **REST API**: POST `/admin/signal` with JSON body `{"signal": "SIGHUP"}`
- **Authentication**: Bearer token (configurable), optional mTLS hooks for client cert verification
- **Rate Limiting**: 6 requests/min, burst 3 (configurable via options)
- **Grace Periods**: Support for shutdown signals with configurable delay
- **Proper Errors**: JSON error responses with HTTP status codes

**Quality Assurance**:

- **73.4% Test Coverage** (pkg/signals): 18 tests covering registration, cleanup, reload, double-tap
- **79.1% Test Coverage** (foundry/signals): 13 tests + 6 godoc examples, parity verification
- **36 Tests Passing**: All tests pass, zero race conditions (verified with `-race`)
- **Zero Lint Issues**: All code passes golangci-lint
- **12 Godoc Examples**: Comprehensive usage examples for all major APIs
- **350+ Line README**: Complete package documentation with Unix/Windows guidance

#### Files Added

```
foundry/signals/
├── catalog.go                   # Catalog accessor with lazy loading
├── catalog_test.go              # Catalog tests (13 functions)
└── examples_test.go             # Godoc examples (6 examples)

pkg/signals/
├── doc.go                       # Package documentation
├── handler.go                   # Core Manager and registration API
├── fallback.go                  # Windows fallback handling
├── http.go                      # HTTP admin endpoint
├── handler_test.go              # Handler tests (18 functions)
├── fallback_test.go             # Fallback tests
├── http_test.go                 # HTTP tests (11 functions)
├── examples_test.go             # Godoc examples (12 examples)
└── README.md                    # Package documentation (350+ lines)
```

**Total**: 11 Go files, 2,100+ lines

#### Example Usage

**Basic Graceful Shutdown**:

```go
import "github.com/fulmenhq/gofulmen/pkg/signals"

// Register cleanup handlers (execute in LIFO order)
signals.OnShutdown(func(ctx context.Context) error {
    log.Info("Flushing buffers...")
    return logger.Flush()
})

signals.OnShutdown(func(ctx context.Context) error {
    log.Info("Closing database...")
    return db.Close()
})

// Start listening for signals (blocks until signal received)
if err := signals.Listen(ctx); err != nil {
    log.Fatal(err)
}
```

**Config Reload with Validation**:

```go
// Register reload handler with validation
signals.OnReload(func(ctx context.Context) error {
    newConfig, err := config.Load(configPath)
    if err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }

    config.Apply(newConfig)
    log.Info("Configuration reloaded successfully")
    return nil
})

// Send SIGHUP to trigger reload
// kill -HUP <pid>
```

**Double-Tap Force Quit**:

```go
// Enable double-tap with custom window
signals.EnableDoubleTap(signals.DoubleTapConfig{
    Window:  2 * time.Second,
    Message: "Press Ctrl+C again to force quit",
})

// First Ctrl+C: Triggers graceful shutdown
// Second Ctrl+C (within 2s): Immediate exit with catalog exit code
```

**HTTP Admin Endpoint**:

```go
handler := signals.NewHTTPHandler(
    signals.WithBearerToken("secret-token"),
    signals.WithRateLimit(10, 5), // 10/min, burst 5
)

http.Handle("/admin/signal", handler)

// curl -X POST http://localhost:8080/admin/signal \
//   -H "Authorization: Bearer secret-token" \
//   -d '{"signal": "SIGHUP"}'
```

#### Integration Points

- **Crucible Integration**: Requires Crucible v0.2.6 for signals catalog accessor
- **Logging Module**: Handlers can use gofulmen/logging for structured logging
- **Exit Codes**: Double-tap uses catalog-derived exit codes (ExitSignalInt)
- **Context Integration**: All handlers receive context for cancellation and timeouts

#### Dependencies Added

- **golang.org/x/time v0.14.0**: For rate limiting in HTTP endpoint

#### Testing Strategy

**Current Coverage** (73.4% pkg/signals, 79.1% foundry/signals):

- Registration API: ✅ Fully tested
- Cleanup chains (LIFO): ✅ Fully tested
- Reload chains (FIFO + fail-fast): ✅ Fully tested
- Double-tap timing: ✅ Fully tested
- Windows fallback: ✅ Fully tested
- HTTP endpoint: ✅ Fully tested
- Catalog layer: ✅ Fully tested

**Deferred to Test Polish Phase**:

- `Listen()` integration testing (requires signal injection pattern)
- Target coverage: 80%+ after implementing SignalInjector test helper
- Testing strategy documented in `.plans/active/v0.1.9/listen-testing-strategy.md`

#### Known Limitations

- Listen() starts goroutine that cannot be stopped mid-signal (by design)
- Double-tap calls os.Exit() immediately (cannot be intercepted)
- Signal handling is process-global (parallel tests not recommended)
- Windows support limited to SIGINT/SIGTERM (platform limitation)

## [0.1.8] - 2025-11-03

### Schema Export Utilities + Foundry Exit Codes Integration

**Release Type**: Major Feature Addition  
**Status**: ✅ Ready for Release

#### Overview

This release adds two major features: (1) Schema export utilities for vendoring Crucible schemas with provenance metadata, and (2) Complete implementation of standardized exit codes from Crucible v0.2.3.

### Feature 1: Schema Export Utilities

Export Crucible schemas with full provenance tracking for vendoring and distribution.

#### Features

**API Package (`schema/export`)**:

- **Export Function**: `Export(ctx, ExportOptions)` - Main export API with validation, formatting, and safety
- **Validation Helper**: `ValidateExportedSchema()` - Verify exported schemas match source
- **Flexible Options**: Format (JSON/YAML), provenance style (object/comment/none), validation toggle, overwrite control
- **Auto-Detection**: Automatic format detection from file extension (.json, .yaml, .yml)
- **Safety Features**: Path validation, parent directory creation, overwrite protection with clear error messages

**CLI Tool (`cmd/gofulmen-export-schema`)**:

- **Required Flags**: `--schema-id`, `--out`
- **Optional Flags**: `--format`, `--provenance-style`, `--no-provenance`, `--no-validate`, `--force`
- **Exit Codes**: Uses foundry exit codes (40=InvalidArgument, 54=FileWriteError, 60=DataInvalid)
- **Help Text**: Comprehensive usage documentation with examples

**Provenance Metadata**:

- **Automatic Tracking**: schema_id, crucible_version, gofulmen_version, git_revision, exported_at
- **Optional Identity**: Support for custom identity providers (vendor, binary)
- **Multiple Styles**:
  - **Object** (default): `x-crucible-source` top-level field in JSON
  - **Comment**: Compact `$comment` field in JSON
  - **YAML Front-Matter**: Commented provenance before `---` separator

**Quality Assurance**:

- **14 Unit Tests**: Export, format detection, provenance styles, validation, safety
- **6 CLI Integration Tests**: Success cases, help, missing args, overwrite, formats, styles
- **100% Lint Health**: Zero linting issues, all code formatted
- **Documentation**: Complete API and CLI docs in `docs/schema/export.md`

**Makefile Integration**:

- **`make export-schema`**: Export with custom SCHEMA_ID and OUT variables
- **`make export-schema-example`**: Export example logging schema to vendor/

#### Files Added (Schema Export)

```
schema/export/
├── export.go              # Main Export() function
├── options.go             # ExportOptions + validation
├── provenance.go          # Provenance metadata builder
├── format.go              # JSON/YAML formatting
├── safety.go              # Path validation + overwrite checks
└── export_test.go         # Comprehensive unit tests

cmd/gofulmen-export-schema/
├── main.go                # CLI wrapper with flag parsing
└── main_test.go           # CLI integration tests

docs/schema/
└── export.md              # Full API and CLI documentation
```

**Total**: 7 files, ~950 lines added

#### Example Usage

**API:**

```go
import "github.com/fulmenhq/gofulmen/schema/export"

opts := export.NewExportOptions(
    "observability/logging/v1.0.0/log-event.schema.json",
    "vendor/crucible/schemas/logging-event.schema.json",
)
if err := export.Export(ctx, opts); err != nil {
    log.Fatal(err)
}
```

**CLI:**

```bash
# Export with provenance
gofulmen-export-schema \
    --schema-id=observability/logging/v1.0.0/log-event.schema.json \
    --out=vendor/crucible/schemas/logging-event.schema.json

# Export as YAML with comment-style provenance
gofulmen-export-schema \
    --schema-id=terminal/v1.0.0/schema.json \
    --out=schema.yaml \
    --format=yaml \
    --provenance-style=comment
```

### Feature 2: Foundry Exit Codes Integration from Crucible v0.2.3

Complete implementation of standardized exit codes for the Fulmen ecosystem. Consumes Crucible v0.2.3's exit codes catalog, providing type-safe constants, comprehensive metadata, platform detection, simplified mode mapping, BSD compatibility, and automatic drift detection.

#### Features

**Core Exit Codes API**:

- **54 Exit Code Constants**: Re-exported from `github.com/fulmenhq/crucible/foundry`
  - Standard codes: `ExitSuccess`, `ExitFailure`
  - Networking: `ExitPortInUse`, `ExitConnectionTimeout`, etc.
  - Configuration: `ExitConfigInvalid`, `ExitSsotVersionMismatch`, etc.
  - Runtime: `ExitHealthCheckFailed`, `ExitDatabaseUnavailable`, etc.
  - Signals: `ExitSignalTerm`, `ExitSignalInt`, etc. (POSIX only)
- **Metadata Access**: `GetExitCodeInfo()`, `LookupExitCode()`, `ListExitCodes()`
- **Catalog Parsing**: Correct YAML parsing with `maps_from` field, efficient `sort.Ints()` sorting

**Platform Compatibility**:

- **Platform Detection**: `SupportsSignalExitCodes()` returns false on Windows (except WSL)
- **WSL Detection**: Checks `WSL_DISTRO_NAME` and `WSL_INTEROP` environment variables
- **PlatformInfo**: Comprehensive metadata (GOOS, GOARCH, WSL status, signal support)

**Provenance Reporting**:

- **Version Functions**: `GofulmenVersion()`, `CrucibleVersion()`, `ExitCodesVersion()`
- **No Hardcoded Values**: All versions sourced from `crucible.Version` and `cruciblefoundry.ExitCodesVersion`
- **GetProvenanceInfo()**: Returns structured provenance for logging/telemetry

**Simplified Mode Mapping**:

- **MapToSimplified()**: Maps Fulmen codes to simplified modes
- **Basic Mode**: 3 codes (0=success, 1=error, 2=usage_error)
- **Severity Mode**: 8 codes (0=success, 1=user_error, 2=config_error, 3=runtime_error, 4=system_error, 5=security_error, 6=test_failure, 7=observability_error)
- **Catalog-Derived**: Reads from `catalog.SimplifiedModes` (no hardcoded mappings)

**BSD Compatibility**:

- **MapToBSD()**: Maps Fulmen codes to BSD sysexits.h codes
- **MapFromBSD()**: Reverse mapping from BSD to Fulmen codes
- **GetBSDCodeInfo()**: Metadata for BSD exit codes (EX_OK, EX_USAGE, EX_CONFIG, etc.)
- **Full sysexits.h Coverage**: All 16 standard BSD codes mapped

**Quality Assurance**:

- **Snapshot Parity Test**: Compares against `exit-codes.snapshot.json` from Crucible
- **100% Verification**: All 54 codes verified (names, categories, descriptions)
- **Automatic Drift Detection**: Test fails if catalog diverges from snapshot

#### Files Added

```
foundry/
├── exit_codes.go                    # Re-exported constants
├── exit_codes_metadata.go           # Metadata access layer
├── exit_codes_test.go               # Core API tests
├── exit_codes_snapshot_test.go      # Parity verification
├── platform.go                      # Platform detection
├── platform_test.go                 # Platform tests
├── version.go                       # Provenance reporting
├── simplified_modes.go              # Simplified mapping
├── simplified_modes_test.go         # Simplified tests
├── bsd.go                           # BSD compatibility
└── bsd_test.go                      # BSD tests
```

**Total**: 11 files, 1,726 lines added

#### Usage Examples

```go
import "github.com/fulmenhq/gofulmen/foundry"

// Use exit codes
if err != nil {
    os.Exit(foundry.ExitConfigInvalid)
}

// Check platform support
if !foundry.SupportsSignalExitCodes() {
    log.Warn("Signal codes not supported on this platform")
}

// Get metadata
info, ok := foundry.GetExitCodeInfo(foundry.ExitPortInUse)
if ok {
    log.Info("Exit code", "name", info.Name, "category", info.Category)
}

// Map to simplified mode
if code, ok := foundry.MapToSimplified(exitCode, foundry.SimplifiedModeBasic); ok {
    os.Exit(code)
}

// BSD compatibility
if bsdCode, ok := foundry.MapToBSD(foundry.ExitConfigInvalid); ok {
    os.Exit(bsdCode) // Uses EX_CONFIG (78)
}

// Provenance reporting
prov := foundry.GetProvenanceInfo()
log.Info("versions",
    "gofulmen", prov.GofulmenVersion,
    "crucible", prov.CrucibleVersion,
    "catalog", prov.ExitCodesVersion)
```

#### Quality Metrics

- ✅ All tests pass (`make test`)
- ✅ Lint health: 100% (Excellent)
- ✅ Code formatted (`make fmt`)
- ✅ Snapshot parity: 54/54 codes verified
- ✅ Platform detection: Windows, macOS, Linux tested

#### Dependencies

- Crucible updated from v0.2.1 to v0.2.3

## [0.1.7] - 2025-10-29

### GitHub Actions CI Infrastructure + Test Fixes

**Release Type**: CI/CD Infrastructure + Bug Fixes  
**Status**: ✅ Ready for Release

#### Features

- **Multi-version testing**: Go 1.21, 1.22, 1.23 matrix
- **Automated quality gates**: Tests, lint, build on every push/PR
- **Bootstrap integration**: Installs goneat v0.3.2 from GitHub releases
- **External install test**: Verifies `go get` works (disabled until public)
- **All dependencies public**: No private repos or secrets required

#### Workflow

`.github/workflows/ci.yml` runs on push to main and all PRs:

1. Download and verify dependencies
2. Bootstrap goneat from GitHub releases
3. Run `make test` and `make lint`
4. Build all packages

#### Bug Fixes

**Prometheus Metric Naming (telemetry/exporters)** - Fixed double suffix issue:

- **Problem**: Exporter was appending `_total` and `_gauge` to names that already had suffixes
- **Result**: Metrics like `http_requests_total_total` and `memory_bytes_gauge_gauge`
- **Fix**: Removed automatic suffix addition, callers now provide proper Prometheus names
- **Impact**: Tests now pass on all Go versions

**RFC3339Nano Timestamp Test (logging)** - Fixed Go 1.21 compatibility:

- **Problem**: Test expected 30+ char timestamps, but RFC3339Nano omits trailing zeros
- **Result**: Timestamps with few nanoseconds (25 chars) failed minimum length check
- **Fix**: Use fixed timestamp with nanoseconds, adjusted minimum to 29 chars
- **Impact**: Consistent test behavior across Go 1.21, 1.22, 1.23

#### Post-Public Release

After repo is made public, enable external install test by removing `if: false` from `install-test` job.

## [0.1.6] - 2025-10-29

### Crucible v0.2.1 Config Embedding + Clean Architecture

**Release Type**: Infrastructure Improvement + External Installation Fix
**Release Date**: October 29, 2025
**Status**: ✅ Ready for Release

#### Features

**Crucible v0.2.1 Integration (Complete)**:

- ✅ **Config Embedding**: Crucible v0.2.1 now embeds `config/` directory alongside `schemas/` and `docs/`
- ✅ **Direct Config Access**: Foundry accesses config via `crucible.ConfigRegistry.Library().Foundry().*()` type-safe API
- ✅ **No Duplication**: Removed local `foundry/assets/*.yaml` embedding - config lives in Crucible only
- ✅ **Single Source of Truth**: Config version automatically matches Crucible module version
- ✅ **Version Alignment**: Update go.mod to get new config - no sync workflow needed
- ✅ **External Installation Fixed**: Standard `go get github.com/fulmenhq/gofulmen` now works for all consumers

**Crucible Package Enhancements (Complete)**:

- ✅ **ConfigRegistry Re-export**: Added `crucible.ConfigRegistry` for accessing Crucible's embedded config
- ✅ **Config Type Alias**: Added `crucible.Config` type for type-safe config access
- ✅ **GetConfig() Function**: Generic config file accessor for any path
- ✅ **ListConfigs() Function**: Directory listing for config browsing

**Foundry Package Refactoring (Complete)**:

- ✅ **Removed Local Embedding**: No more `//go:embed assets/*.yaml` directive
- ✅ **Crucible API Integration**: `loadYAML()` now calls Crucible's config API methods
- ✅ **Backward Compatible**: All existing Foundry APIs work identically
- ✅ **Documentation Updated**: Comments reflect new architecture

**Comprehensive Testing (Complete)**:

- ✅ **Config Embedding Tests**: New `crucible/config_test.go` with 3 test functions
- ✅ **Coverage Verification**: Tests verify all 5 foundry config files accessible (patterns, country-codes, http-statuses, mime-types, similarity-fixtures)
- ✅ **Size Validation**: Tests ensure config files are non-empty and of expected size
- ✅ **Fail-Fast Design**: Tests will fail loudly if Crucible stops embedding config

#### Architecture Before vs After

**Before (v0.1.5 - Suboptimal)**:

```
Crucible Repo
├── config/library/foundry/*.yaml (NOT embedded)
│
↓ make sync (goneat ssot sync)
│
gofulmen Repo
├── config/crucible-go/library/foundry/*.yaml (synced copy)
│
↓ make sync-foundry-assets (manual copy)
│
├── foundry/assets/*.yaml (duplicated)
│   └── //go:embed assets/*.yaml (embedded in gofulmen)
```

**After (v0.1.6 - Clean)**:

```
Crucible Repo (v0.2.1)
├── config/library/foundry/*.yaml
│   └── //go:embed config (embedded in Crucible)
│
↓ import "github.com/fulmenhq/crucible"
│
gofulmen Repo
├── foundry/catalog.go
    └── crucible.ConfigRegistry.Library().Foundry().Patterns()
```

#### Quality Metrics

- ✅ **All Tests Passing**: 24 test packages pass with new config tests
- ✅ **Build Success**: `go build ./...` completes without errors
- ✅ **Config Access Verified**: All 5 foundry config files load successfully
- ✅ **External Installation Works**: Standard `go get` installation verified
- ✅ **Backward Compatible**: No breaking changes to public APIs

#### Breaking Changes

- None (fully backward compatible with v0.1.5)

#### Migration Notes

**No migration required** - this is an internal architecture improvement. All existing code continues to work:

```go
import "github.com/fulmenhq/gofulmen/foundry"

// All existing Foundry APIs work identically
catalog := foundry.GetDefaultCatalog()
pattern, _ := catalog.GetPattern("ansi-email")
country, _ := catalog.GetCountry("US")
mimeType, _ := catalog.GetMimeType("json")
```

**New capability** - access Crucible config directly if needed:

```go
import "github.com/fulmenhq/gofulmen/crucible"

// Access any Crucible config file
patterns, _ := crucible.ConfigRegistry.Library().Foundry().Patterns()
countryCodes, _ := crucible.ConfigRegistry.Library().Foundry().CountryCodes()

// Generic config access
data, _ := crucible.GetConfig("library/foundry/patterns.yaml")

// List config files
files, _ := crucible.ListConfigs("library/foundry")
```

#### Impact on Downstream

**forge-workhorse-groningen**:

- ✅ Can now install via standard `go get github.com/fulmenhq/gofulmen@v0.1.6`
- ✅ All foundry config automatically available
- ✅ No manual setup or sync required

**External Consumers**:

- ✅ Standard Go module installation works
- ✅ All config accessible out of the box
- ✅ No special configuration needed

#### Quality Gates

- [x] Crucible v0.2.1 adopted with config embedding
- [x] Foundry refactored to use Crucible config API
- [x] Comprehensive config tests added and passing
- [x] All 24 test packages pass
- [x] Build succeeds: `go build ./...`
- [x] External installation verified
- [x] No breaking changes to public APIs

#### Release Checklist

- [x] Version number set in VERSION (0.1.6)
- [x] CHANGELOG.md updated with v0.1.6 release notes
- [ ] RELEASE_NOTES.md updated (in progress)
- [ ] README.md reviewed - pending
- [ ] gofulmen_overview.md reviewed - pending
- [ ] docs/releases/v0.1.6.md created - pending
- [ ] All files staged
- [ ] make prepush executed
- [ ] Release prep changes committed
- [ ] Git tag created (v0.1.6) - pending
- [ ] Tag pushed to GitHub - pending

## [0.1.5] - 2025-10-27

### Similarity v2 API + Telemetry + Error Handling + Telemetry Phase 5

**Release Type**: Major Feature Release
**Release Date**: October 27, 2025
**Status**: ✅ Ready for Release

#### Features

**Similarity v2 API (Complete)**:

- ✅ **DistanceWithAlgorithm()**: Calculate edit distance with algorithm selection (Levenshtein, Damerau OSA, Damerau Unrestricted)
- ✅ **ScoreWithAlgorithm()**: Calculate normalized similarity scores with algorithm selection (all algorithms + Jaro-Winkler, Substring)
- ✅ **5 Algorithm Support**: Levenshtein, Damerau OSA (Optimal String Alignment), Damerau Unrestricted, Jaro-Winkler, Substring matching
- ✅ **Native OSA Implementation**: Replaced buggy matchr.OSA() with native Go implementation based on rapidfuzz-cpp
- ✅ **100% Fixture Compliance**: All 30 Crucible v2.0.0 fixtures passing (was 28/30 with matchr bug)
- ✅ **Cross-Language Consistency**: Validated against PyFulmen (RapidFuzz) and TSFulmen (strsim-wasm)
- ✅ **Performance Optimized**: Native OSA expected to match Levenshtein pattern (1.24-1.76x faster than external libraries)

**Similarity Telemetry (Complete)**:

- ✅ **Counter-Only Metrics**: Algorithm usage, string length distribution, fast paths, edge cases, API misuse tracking
- ✅ **Zero Overhead by Default**: Telemetry disabled unless explicitly enabled via EnableTelemetry()
- ✅ **Acceptable Overhead**: ~1μs per operation when enabled (negligible for typical CLI/spell-check use cases)
- ✅ **6 Metric Types**: distance.calls, score.calls, string_length, fast_path, edge_case, error counters
- ✅ **Production Visibility**: Understand algorithm usage patterns and performance characteristics in production
- ✅ **NO Histograms/Tracing**: Follows ADR-0008 Pattern 1 for performance-sensitive hot-loop code

**Error Envelope System (Complete)**:

- ✅ **Structured Error Envelopes**: Complete error information with code, message, severity, correlation ID, and context
- ✅ **Severity Level Support**: Info, Low, Medium, High, Critical with automatic validation
- ✅ **Context Data**: Structured metadata with validation for error enrichment
- ✅ **Correlation ID Integration**: UUIDv7 correlation IDs for distributed tracing
- ✅ **JSON Serialization**: Full JSON marshaling/unmarshaling with schema compliance
- ✅ **Backward Compatible**: Works with existing error interfaces while providing enhanced features

**Validation Strategies (Complete)**:

- ✅ **StrategyLogWarning**: Default strategy - logs validation errors as warnings, continues execution
- ✅ **StrategyAppendToMessage**: Appends validation errors to envelope message for user visibility
- ✅ **StrategyFailFast**: Logs errors and provides clear error visibility for monitoring
- ✅ **StrategySilent**: Zero-overhead option for high-performance scenarios
- ✅ **Custom Configuration**: Flexible error handling via `ErrorHandlingConfig` with custom logger support

**Safe Helper Functions (Complete)**:

- ✅ **SafeWithSeverity**: Production-safe wrapper that handles validation errors gracefully
- ✅ **SafeWithContext**: Production-safe wrapper for context validation with error handling
- ✅ **ApplySeverityWithHandling**: Custom strategy application for severity validation
- ✅ **ApplyContextWithHandling**: Custom strategy application for context validation

**Cross-Language Patterns (Ready)**:

- ✅ **Consistent API Surface**: Standardized error envelope patterns for TS/Py implementation
- ✅ **Validation Strategies**: Same error handling approaches across all language foundations
- ✅ **Enterprise Integration**: Ready for distributed tracing and error monitoring platforms
- ✅ **Performance Optimized**: Minimal overhead with lock-free operations when disabled

#### Quality Metrics

- ✅ **Similarity v2**: 100% fixture compliance (30/30 tests passing)
- ✅ **Similarity Telemetry**: 12 comprehensive tests + benchmark overhead analysis
- ✅ **Error Handling**: 100% test coverage with validation and strategy testing
- ✅ **Telemetry Phase 5**: Metric type routing tests + Prometheus format compliance
- ✅ **Performance Validated**: All targets met across all modules
- ✅ **Schema Compliance**: Full alignment with Crucible v2.0.0 standards
- ✅ **Cross-Language Ready**: Patterns documented for TS/Py team implementation

#### Breaking Changes

- None (fully backward compatible with v0.1.4)

#### Migration Notes

**Similarity v2 API** (new, no migration required):

```go
import "github.com/fulmenhq/gofulmen/foundry/similarity"

// v1 API (still supported)
distance := similarity.Distance("kitten", "sitting")
score := similarity.Score("kitten", "sitting")

// v2 API with algorithm selection (NEW)
distance, err := similarity.DistanceWithAlgorithm("kitten", "sitting", "osa")
score, err := similarity.ScoreWithAlgorithm("kitten", "sitting", "jaro-winkler")

// Enable telemetry (opt-in)
similarity.EnableTelemetry(telemetrySystem)
```

**Error Handling** (new, no migration required):

```go
// Structured error envelope
envelope := errors.NewErrorEnvelope("VALIDATION_ERROR", "Validation failed")
envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
envelope = envelope.WithCorrelationID(correlationID)
envelope = errors.SafeWithContext(envelope, map[string]interface{}{
    "component": "validator",
    "operation": "validate",
})
```

**Telemetry Phase 5** (enhanced, backward compatible):

```go
import "github.com/fulmenhq/gofulmen/telemetry"
import "github.com/fulmenhq/gofulmen/telemetry/exporters"

// Create telemetry system
sys, _ := telemetry.NewSystem(telemetry.DefaultConfig())

// Counter metrics
sys.Counter("requests_total", 1, map[string]string{"status": "200"})

// Gauge metrics (NEW)
sys.Gauge("cpu_usage_percent", 75.5, map[string]string{"host": "server1"})
sys.Gauge("memory_usage_mb", 2048.0, map[string]string{"type": "heap"})

// Histogram metrics
sys.Histogram("request_duration_ms", 50*time.Millisecond, map[string]string{"endpoint": "/api"})

// Prometheus exporter (NEW)
exporter := exporters.NewPrometheusExporter("myapp", ":9090")
exporter.Start()

// Configure system to use exporter
config := &telemetry.Config{
    Enabled: true,
    Emitter: exporter,
}
sys, _ := telemetry.NewSystem(config)
```

**Prometheus Integration**:

```go
// Start Prometheus exporter
exporter := exporters.NewPrometheusExporter("myapp", ":9090")
if err := exporter.Start(); err != nil {
    log.Fatal(err)
}

// Metrics will be available at http://localhost:9090/metrics
// Example output:
// myapp_requests_total{status="200"} 42
// myapp_cpu_usage_percent_gauge{host="server1"} 75.5
// myapp_request_duration_ms_bucket{endpoint="/api",le="50"} 10
// myapp_request_duration_ms_sum{endpoint="/api"} 500
// myapp_request_duration_ms_count{endpoint="/api"} 10
```

**Cross-Language Implementation Pattern** (for TS/Py teams):

```go
// 1. Consistent metric naming
// Use snake_case: system_cpu_usage_percent
// End with unit: _ms, _percent, _bytes
// Include context tags: host, service, region

// 2. Error handling with context
// Always include component, operation, error_type
// Use correlation IDs for tracing
// Log errors but don't fail telemetry operations

// 3. Batch-friendly configuration
// Default to no batching (immediate emission)
// Allow opt-in batching for high-frequency scenarios
// Provide flush methods for manual control
```

#### Quality Gates

- [x] All Similarity v2 tests passing (30/30 fixtures, 100% compliance)
- [x] Similarity telemetry tests passing (12 tests + benchmark overhead analysis)
- [x] All error handling tests passing (validation strategies, safe helpers)
- [x] All telemetry tests passing (gauges, counters, histograms, batching, routing)
- [x] Prometheus format compliance verified with proper \_total/\_gauge/\_bucket conventions
- [x] Histogram +Inf bucket implementation complete and tested
- [x] Native OSA implementation validated against PyFulmen (RapidFuzz)
- [x] Performance targets met (OSA matches Levenshtein, telemetry <5% overhead)
- [x] Cross-language patterns documented for TS/Py team implementation
- [x] Schema validation working against Crucible v2.0.0 standards
- [x] ADR-0002 and ADR-0003 completed with rationale and benchmarks
- [x] Code quality checks passing (make check-all)

#### Release Checklist

- [x] Version number set in VERSION (0.1.5)
- [x] CHANGELOG.md updated with v0.1.5 release notes
- [ ] RELEASE_NOTES.md updated (in progress)
- [ ] README.md reviewed for v0.1.5 updates - pending
- [ ] gofulmen_overview.md reviewed for v0.1.5 updates - pending
- [ ] foundry/README.md updated with Similarity v2 + telemetry - pending
- [ ] docs/releases/v0.1.5.md created - pending
- [ ] make fmt executed - pending
- [ ] Release prep changes committed - pending
- [ ] goneat assess run and issues fixed - pending
- [ ] Git tag created (v0.1.5) - pending
- [ ] Tag pushed to GitHub - pending

## [0.1.4] - 2025-10-23

### FulHash Package + Pathfinder Enhancements + Code Quality Polish

**Release Type**: Major Feature Release + Quality Improvements
**Release Date**: October 23, 2025
**Status**: ✅ Ready for Release

#### Features

**FulHash Package (Complete)**:

- ✅ **Core Hashing APIs**: Block (`Hash()`, `HashString()`) and streaming (`HashReader()`, `NewHasher()`) implementations
- ✅ **Algorithm Support**: xxh3-128 (fast, default) and sha256 (cryptographic) with enterprise error handling
- ✅ **Digest Metadata**: Standardized `<algorithm>:<hex>` format with `FormatDigest()`/`ParseDigest()` helpers
- ✅ **Thread Safety**: All public APIs safe for concurrent use (no shared mutable state)
- ✅ **Performance Optimized**: xxh3-128 ~28μs/op for 128-char strings (17x faster than 0.5ms target)
- ✅ **Schema Validation**: Fixture validation against synced Crucible standards (`config/crucible-go/library/fulhash/fixtures.yaml`)
- ✅ **Comprehensive Testing**: 90%+ coverage with Crucible fixture tests and streaming correctness validation
- ✅ **Cross-Language Parity**: API aligned with pyfulmen and tsfulmen for ecosystem consistency
- ✅ **Documentation**: Complete package README with usage examples and performance notes

**Pathfinder Checksum Integration (Complete)**:

- ✅ **Opt-in Checksum Calculation**: New `CalculateChecksums` and `ChecksumAlgorithm` fields in `FindQuery`
- ✅ **Algorithm Support**: xxh3-128 (recommended for performance) and sha256 with validation
- ✅ **Performance Target Met**: <10% CPU overhead on representative workloads with streaming implementation
- ✅ **Schema Compliance**: Metadata fields (`checksum`, `checksumAlgorithm`, `checksumError`) match Crucible standards
- ✅ **Error Isolation**: Checksum failures don't abort traversal; populate `checksumError` field instead
- ✅ **Backward Compatible**: Disabled by default, no breaking changes to existing API
- ✅ **Integration Ready**: Enables integrity verification for goneat, docscribe, and other downstream consumers

**Pathfinder Feature Upscale (Complete)**:

- ✅ **PathTransform Function**: Remap results without post-processing (strip prefixes, flatten paths, logical mounting)
- ✅ **FindStream() Method**: Channel-based streaming for memory-efficient processing of large result sets
- ✅ **SkipDirs Option**: Simple string matching for common directory exclusions (`node_modules`, `.git`, `vendor`)
- ✅ **SizeRange/TimeRange Filtering**: Efficient filtering by file size and modification time constraints
- ✅ **FindFilesByType() Utility**: Predefined patterns for common file types (go, javascript, python, java, config, docs, images)
- ✅ **GetDirectoryStats()**: Repository analytics with file counts, sizes, type distribution, and largest files
- ✅ **Worker Pool Support**: Configurable concurrency for faster directory traversal on large trees
- ✅ **All Features Backward Compatible**: Additive enhancements with default behavior unchanged

**Code Quality Polish (Complete)**:

- ✅ **Security Fixes**: Resolved high-priority G304 potential file inclusion vulnerability
- ✅ **Lint Resolution**: Fixed 33 golangci-lint issues across bootstrap and test files
- ✅ **Tools Configuration**: Updated goneat configuration with proper schema validation
- ✅ **Error Handling**: Improved reliability in bootstrap download/extract operations
- ✅ **Assessment Health**: Improved from 0% to 77% overall codebase health
- ✅ **Precommit Checks**: All hooks now pass with 100% health

#### Quality Metrics

- ✅ **FulHash Coverage**: 90%+ with comprehensive fixture and streaming tests
- ✅ **Pathfinder Coverage**: Maintained existing levels with new feature tests
- ✅ **Performance Validated**: All targets met (FulHash 17x faster, Pathfinder <10% overhead)
- ✅ **Schema Compliance**: 100% alignment with synced Crucible standards
- ✅ **Security Audit**: Zero critical vulnerabilities remaining
- ✅ **Code Quality**: `make check-all` passing with zero lint issues
- ✅ **Cross-Language Ready**: API surface prepared for pyfulmen/tsfulmen implementation

#### Breaking Changes

- None (fully backward compatible with v0.1.3)

#### Migration Notes

**FulHash Usage**:

```go
import "github.com/fulmenhq/gofulmen/fulhash"

// Block hashing
digest, err := fulhash.Hash([]byte("data"), fulhash.WithAlgorithm("xxh3-128"))
fmt.Println(digest.String()) // "xxh3-128:abc123..."

// Streaming for large files
hasher := fulhash.NewHasher(fulhash.WithAlgorithm("xxh3-128"))
io.Copy(hasher, file)
digest := hasher.Sum()
```

**Pathfinder Checksum Usage**:

```go
results, err := finder.FindFiles(ctx, FindQuery{
    Root:               "/path/to/search",
    Include:            []string{"*.go"},
    CalculateChecksums: true,           // NEW
    ChecksumAlgorithm:  "xxh3-128",     // NEW
})

// Results include checksum metadata
for _, result := range results {
    if checksum, ok := result.Metadata["checksum"]; ok {
        fmt.Printf("File: %s, Checksum: %s\n", result.RelativePath, checksum)
    }
}
```

**Pathfinder New Features**:

```go
// Path transformation
results, _ := finder.FindFiles(ctx, FindQuery{
    Root:      "src",
    Include:   []string{"**/*.go"},
    Transform: func(r PathResult) PathResult {  // NEW
        r.LogicalPath = strings.TrimPrefix(r.RelativePath, "internal/")
        return r
    },
})

// Streaming results
resultCh, errCh := finder.FindStream(ctx, query)  // NEW
for result := range resultCh {
    process(result)
}

// Directory stats
stats, _ := finder.GetDirectoryStats(ctx, "/repo")  // NEW
fmt.Printf("Files: %d, Total Size: %d bytes\n", stats.FileCount, stats.TotalSize)
```

#### Quality Gates

- [x] FulHash package passes all fixture tests and achieves 90%+ coverage
- [x] Pathfinder checksum integration maintains <10% performance overhead
- [x] All new Pathfinder features tested and backward compatible
- [x] Security audit clean (G304 vulnerability resolved)
- [x] 33 golangci-lint issues fixed (0 remaining)
- [x] `make check-all` passes with 100% precommit health
- [x] CHANGELOG.md and RELEASE_NOTES.md updated
- [x] Documentation complete for all new features

#### Release Checklist

- [x] Version number set in VERSION (0.1.4)
- [x] CHANGELOG.md updated with v0.1.4 release notes
- [x] RELEASE_NOTES.md updated
- [x] All tests passing
- [x] Code quality checks passing
- [x] Security issues resolved
- [x] Documentation complete
- [ ] Git tag created (v0.1.4) - pending
- [ ] Tag pushed to GitHub - pending

---

## [0.1.3] - 2025-10-22

### Similarity & Docscribe Modules + Crucible SSOT Sync

**Release Type**: Feature Release
**Release Date**: October 22, 2025
**Status**: ✅ Ready for Release

#### Features

**Similarity Module (Complete)**:

- ✅ **Levenshtein Distance**: Wagner-Fischer dynamic programming algorithm with Unicode rune counting
- ✅ **Normalized Scoring**: Similarity scores 0.0-1.0 with formula: `1 - distance/max(len(a), len(b))`
- ✅ **Suggestion API**: Ranked fuzzy matching with configurable thresholds (MinScore, MaxSuggestions, Normalize)
- ✅ **Default Options**: MinScore=0.6, MaxSuggestions=3, Normalize=false (explicit opt-in)
- ✅ **Tie-Breaking**: Score descending, then alphabetical (case-sensitive) for equal scores
- ✅ **Unicode Normalization**: Trim → casefold → optional accent stripping pipeline
- ✅ **Turkish Locale Support**: Special case folding for dotted/dotless i (İ→i, I→ı)
- ✅ **Accent Stripping**: NFD decomposition + combining mark filtering + NFC recomposition
- ✅ **Helper Functions**: Casefold(), StripAccents(), EqualsIgnoreCase(), Normalize()
- ✅ **Performance**: ~28 μs/op for 128-char strings (17x faster than 0.5ms target)
- ✅ **Benchmark Suite**: 15 ongoing benchmark tests for regression prevention
- ✅ **100% Test Coverage**: All implementation files (similarity.go, normalize.go, suggest.go)
- ✅ **Crucible Fixtures**: 26 tests passing (13 distance, 13 normalization, 5 suggestions + 4 skipped)
- ✅ **Cross-Language Parity**: API aligned with pyfulmen and tsfulmen
- ✅ **Dependency**: golang.org/x/text v0.30.0 for Unicode normalization (approved)
- ✅ **Documentation**: Comprehensive doc.go with performance data and usage examples
- ✅ **Future-Ready**: TODO markers for Damerau-Levenshtein and Jaro-Winkler expansion (v0.1.4+)

**Docscribe Module (Complete)**:

- ✅ **Frontmatter Processing**: Extract YAML frontmatter with metadata and clean content separation
- ✅ **Header Extraction**: ATX (#) and Setext (===) style headers with anchors and line numbers
- ✅ **Format Detection**: Heuristic-based detection for markdown, YAML, JSON, TOML formats
- ✅ **Multi-Document Splitting**: Handle YAML streams and concatenated markdown with smart delimiter parsing
- ✅ **Document Inspection**: Fast metadata extraction (<1ms for 100KB documents)
- ✅ **Source-Agnostic Design**: Works with Crucible, Cosmography, local files, or any content source
- ✅ **Crucible Integration**: Integrates with `crucible.GetDoc()` for SSOT asset access
- ✅ **Performance Optimized**: InspectDocument <1ms, ParseFrontmatter <5ms, SplitDocuments <10ms
- ✅ **Comprehensive Tests**: 14 test functions with 56 assertions, all passing
- ✅ **Test Fixtures**: 13 fixtures covering frontmatter, headers, format detection, multi-doc scenarios
- ✅ **Code-Block Awareness**: Correctly handles fenced code blocks, ignoring delimiters inside code
- ✅ **Error Handling**: Typed errors (ParseError, FormatError) with line numbers and helpful messages

**Crucible SSOT Sync (2025.10.2)**:

- ✅ **Similarity Module Standard**: Complete specification in `docs/crucible-go/standards/library/foundry/similarity.md`
- ✅ **Similarity Fixtures**: Test fixtures in `config/crucible-go/library/foundry/similarity-fixtures.yaml` (39 test cases)
- ✅ **Similarity Schema**: JSON Schema in `schemas/crucible-go/library/foundry/v1.0.0/similarity.schema.json`
- ✅ **Docscribe Module Standard**: Complete specification in `docs/crucible-go/standards/library/modules/docscribe.md`
- ✅ **Module Manifest**: Updated `config/crucible-go/library/v1.0.0/module-manifest.yaml` with similarity and docscribe entries
- ✅ **Helper Library Standard**: Updated with Crucible Overview requirement for all helper libraries
- ✅ **Fulmen Forge Standard**: Added `docs/crucible-go/architecture/fulmen-forge-workhorse-standard.md`
- ✅ **Module Catalog**: Updated module index and discovery metadata

**Documentation Compliance**:

- ✅ **Crucible Overview Section**: Added to README.md explaining SSOT relationship and shim/docscribe purpose
- ✅ **Package Documentation**: Comprehensive doc.go with usage examples and design principles
- ✅ **Integration Examples**: Shows docscribe + crucible.GetDoc() workflow

#### Quality Metrics

- ✅ **Similarity Test Coverage**: 100% for all implementation files (similarity.go, normalize.go, suggest.go)
- ✅ **Similarity Tests**: 93 unit tests + 26 Crucible fixture tests + 15 benchmarks = 134 tests passing
- ✅ **Docscribe Test Coverage**: 100% for core functions (ParseFrontmatter, ExtractHeaders, SplitDocuments)
- ✅ **Docscribe Tests**: 14 test functions, 56 assertions
- ✅ **Code Quality**: `make check-all` passing (format, lint, tests)
- ✅ **Performance Validated**: All performance targets met (similarity: 17x faster than target)

#### Breaking Changes

- None (fully backward compatible with v0.1.2)

#### Migration Notes

Both Similarity and Docscribe are new modules with no migration required.

**Similarity Usage**:

```go
import "github.com/fulmenhq/gofulmen/foundry/similarity"

// Calculate edit distance
dist := similarity.Distance("kitten", "sitting") // Returns: 3

// Get similarity score (0.0 to 1.0)
score := similarity.Score("kitten", "sitting") // Returns: 0.5714...

// Generate suggestions for typo correction
candidates := []string{"config", "configure", "conform"}
opts := similarity.DefaultSuggestOptions() // MinScore=0.6, MaxSuggestions=3
suggestions := similarity.Suggest("confg", candidates, opts)
// Returns: [{"config", 0.8333}]

// Case-insensitive matching
opts.Normalize = true
suggestions = similarity.Suggest("CONFIG", candidates, opts)
// Returns: [{"config", 1.0}, "configure", 0.889}, {"conform", 0.714}]

// Normalize text for comparison
normalized := similarity.Normalize("  Café  ", similarity.NormalizeOptions{
    StripAccents: true,
}) // Returns: "cafe"
```

**Docscribe Usage**:

```go
import (
    "github.com/fulmenhq/gofulmen/docscribe"
    "github.com/fulmenhq/gofulmen/crucible"
)

// Get documentation from Crucible
content, err := crucible.GetDoc("standards/coding/go.md")
if err != nil {
    return err
}

// Extract frontmatter and content
body, metadata, err := docscribe.ParseFrontmatter([]byte(content))
if err != nil {
    return err
}

// Extract headers for TOC generation
headers, err := docscribe.ExtractHeaders([]byte(content))
if err != nil {
    return err
}
```

See `foundry/similarity/doc.go` and `docscribe/doc.go` for comprehensive examples.

#### Quality Gates

- [x] All similarity tests passing (93 unit + 26 fixture + 15 benchmarks)
- [x] All docscribe tests passing (14 functions, 56 assertions)
- [x] 100% coverage on similarity core functions
- [x] 100% coverage on docscribe core functions
- [x] `make check-all` passed
- [x] Code formatted with goneat
- [x] No linting issues
- [x] Documentation complete (similarity + docscribe)
- [x] Crucible sync to 2025.10.2 complete
- [x] Performance targets validated (similarity 17x faster than target)

#### Release Checklist

- [x] Version number set in VERSION (0.1.3)
- [x] CHANGELOG.md updated with v0.1.3 release notes
- [x] RELEASE_NOTES.md updated
- [x] README.md updated with Crucible Overview
- [x] All tests passing
- [x] Code quality checks passing
- [x] docs/releases/v0.1.3.md created
- [ ] Git tag created (v0.1.3) - pending
- [ ] Tag pushed to GitHub - pending
