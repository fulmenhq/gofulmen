# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.10] - 2025-11-09

### Changed

- **Signals Package Migration** - Moved `pkg/signals/` to `signals/` for consistency with top-level module structure
  - Import path changed from `github.com/fulmenhq/gofulmen/pkg/signals` to `github.com/fulmenhq/gofulmen/signals`
  - All documentation updated to reflect new import path
  - Eliminates confusion for users expecting consistent top-level module structure

### Fixed

- **Template Support** - Fixed import paths referenced in documentation and templates to support downstream microtool development
  - Corrected references from non-existent `pkg/` paths to actual top-level module paths
  - Ensures template examples compile and work correctly with current gofulmen structure

## [Unreleased]

### Added

### Changed

### Fixed

## [0.1.9] - 2025-11-08

### Added

- **Prometheus Exporter** (`telemetry/exporters`) - Production-grade HTTP metrics exposition with enterprise features
  - **Core Exporter**: PrometheusExporter implementing telemetry.MetricsEmitter interface
  - **Three-Phase Refresh Pipeline**: Collect → Convert → Export with health instrumentation at each stage
  - **Format Conversion**: Automatic conversion to Prometheus text exposition format
    - Counters: `<prefix>_<name>_total{labels}`
    - Gauges: `<prefix>_<name>_gauge{labels}`
    - Histograms: Full bucket series with automatic ms→seconds conversion, `_bucket`, `_sum`, `_count` suffixes
  - **HTTP Server**: Built-in HTTP server with configurable endpoint (default `:9090`)
  - **Authentication**: Bearer token authentication with configurable token
  - **Rate Limiting**: Per-IP rate limiting with configurable requests/minute and burst size
  - **Health Instrumentation**: 7 built-in metrics tracking exporter health
    - `prometheus_exporter_refresh_duration_seconds` (histogram)
    - `prometheus_exporter_refresh_total` (counter by phase/result)
    - `prometheus_exporter_refresh_errors_total` (counter by phase/reason)
    - `prometheus_exporter_refresh_inflight` (gauge)
    - `prometheus_exporter_http_requests_total` (counter by endpoint/status)
    - `prometheus_exporter_http_errors_total` (counter by endpoint/status)
    - `prometheus_exporter_restarts_total` (counter by reason)
  - **HTTP Endpoints**: `/metrics` (Prometheus format), `/` (landing page), `/health` (JSON status)
  - **Configuration**: PrometheusConfig with defaults and validation
    - Prefix, endpoint, bearer token, rate limits, refresh interval, quiet mode, read header timeout
  - **Backward Compatibility**: Legacy `NewPrometheusExporter(prefix, endpoint)` constructor preserved
  - **Thread-Safe**: Concurrent metric emission with mutex protection
  - **Test Coverage**: Comprehensive unit tests and integration tests
  - **Documentation**: Enhanced telemetry/README.md with 137+ lines of Prometheus documentation
  - **Examples**: Working examples in examples/phase5-telemetry-demo.go
  - **Module Instrumentation**: 19 module-specific metrics across Foundry, Error Handling, and FulHash
  - **Files Added**: 3 core files (prometheus.go, config.go, http.go) + tests
  - **Crucible Integration**: Updated to v0.2.7 for Prometheus metrics taxonomy

- **Signal Handling Module** (`pkg/signals`) - Cross-platform signal management with graceful shutdown
  - **Catalog Layer** (`foundry/signals`): Typed access to Crucible signals catalog v1.0.0 with 8 standard signals (TERM, INT, HUP, QUIT, PIPE, ALRM, USR1, USR2)
  - **Core API**: `OnShutdown()`, `OnReload()`, `Handle()`, `EnableDoubleTap()` for signal registration and management
  - **Graceful Shutdown**: LIFO cleanup chains with context support and configurable timeouts
  - **Ctrl+C Double-Tap**: 2-second window (configurable) for force quit with catalog-derived defaults
  - **Config Reload**: SIGHUP with validation hooks, restart semantics, and fail-fast execution
  - **Windows Fallback**: Automatic platform detection with INFO-level logging and operation hints for unsupported signals
  - **HTTP Admin Endpoint**: `/admin/signal` helper with bearer token auth, optional mTLS, and rate limiting (6/min, burst 3)
  - **Thread-Safe**: Concurrent handler registration with mutex protection and sync.Once caching
  - **Testing Support**: Manager API for unit testing without OS signals (Listen() integration deferred to test polish phase)
  - **Comprehensive Documentation**: Package README (350+ lines), 12 godoc examples, Unix/Windows operational guidance
  - **Test Coverage**: 73.4% for pkg/signals, 79.1% for foundry/signals (36 tests passing, Listen() testing documented in adjunct)
  - **Files Added**: 11 new files (foundry/signals/\*, pkg/signals/\*, plus tests, examples, README)
  - **Dependencies**: Added golang.org/x/time v0.14.0 for rate limiting
  - **Crucible Integration**: Updated from v0.2.5 to v0.2.6 for signals catalog accessor

- **App Identity Module** - Application identity metadata from `.fulmen/app.yaml`
  - **Core API**: `Get()`, `Must()`, `GetWithOptions()`, `LoadFrom()` for loading identity with caching
  - **Discovery**: Automatic `.fulmen/app.yaml` discovery via ancestor search (max 20 levels)
  - **Precedence**: Context injection → Explicit path → Environment variable (`FULMEN_APP_IDENTITY_PATH`) → Ancestor search
  - **Validation**: Schema validation against Crucible v1.0.0 app-identity schema with field-level diagnostics
  - **Caching**: Thread-safe process-level caching with sync.Once (verified with race detector)
  - **Testing Support**: `WithIdentity()` context injection, `Reset()` cache clearing, `NewFixture()`/`NewCompleteFixture()` test utilities
  - **Integration Helpers**: `ConfigParams()`, `EnvVar()`, `FlagsPrefix()`, `TelemetryNamespace()`, `ServiceName()`
  - **Error Types**: `NotFoundError`, `ValidationError`, `MalformedError` with detailed diagnostics
  - **Zero Dependencies**: Layer 0 module with no Fulmen dependencies (stdlib + gopkg.in/yaml.v3 only)
  - **Test Coverage**: 88.4% coverage with 68 tests passing (includes subtests and examples)
  - **Documentation**: Comprehensive godoc with 8 runnable examples, README integration guide
  - **Files Added**: 12 new files (doc.go, identity.go, loader.go, validation.go, cache.go, override.go, testing.go + 5 test files)
  - **Test Fixtures**: 6 YAML fixtures covering valid/invalid scenarios

### Changed

- **FulHash Telemetry**: Migrated from aggregated metrics to granular Prometheus-compatible metrics
  - Replaced `fulhash_hash_count` with algorithm-specific counters: `fulhash_operations_total_xxh3_128`, `fulhash_operations_total_sha256`
  - Added `fulhash_hash_string_total` for string hashing operations
  - Added `fulhash_bytes_hashed_total` for total bytes processed
  - Added `fulhash_operation_ms` histogram for operation latency
  - Updated error telemetry to emit `fulhash_errors_count` with error_type tags
  - Migrated from deprecated `SetTelemetrySystem()` to `telemetry.SetGlobalSystem()`
  - All 12 FulHash tests passing with new metrics
- **Crucible Dependency**: Updated from v0.2.6 to v0.2.7 (adds Prometheus metrics taxonomy)
- **Agent Workflow**: Added Pre-Commit Checklist to AGENTS.md for consistent commit quality

### Fixed

- **Listen Testing Strategy**: Documented approach for signal Listen() integration testing in adjunct brief (deferred implementation to test polish phase)

## [Unreleased]

### Added

### Changed

### Fixed

## [0.1.8] - 2025-11-03

### Added

- **Schema Export Utilities** - Export Crucible schemas with provenance metadata
  - **API Package**: `schema/export` with `Export()`, `ValidateExportedSchema()` functions
  - **CLI Tool**: `cmd/gofulmen-export-schema` with full flag support (--schema-id, --out, --format, --provenance-style, --force)
  - **Provenance Metadata**: Automatic tracking of schema_id, crucible_version, gofulmen_version, git_revision, exported_at
  - **Multiple Formats**: JSON and YAML output with auto-detection from file extension
  - **Provenance Styles**: Object (default), comment, or none - flexible embedding options
  - **Safety Features**: Overwrite protection, path validation, parent directory creation
  - **Exit Codes**: Proper foundry exit codes for CLI (ExitInvalidArgument, ExitFileWriteError, ExitDataInvalid)
  - **Comprehensive Tests**: Unit tests (14 test cases) + CLI integration tests (6 test cases)
  - **Documentation**: Full API and CLI documentation in `docs/schema/export.md`
  - **Makefile Targets**: `make export-schema` and `make export-schema-example` for convenience
  - **Files Added**: 7 new files (schema/export/\*.go, cmd/gofulmen-export-schema/main.go, tests, docs)
  - **100% Lint Health**: All code passes golangci-lint with zero issues

- **Foundry Exit Codes Integration** - Complete exit codes integration from Crucible v0.2.3
  - **54 Exit Code Constants**: Re-exported from `github.com/fulmenhq/crucible/foundry` (ExitSuccess, ExitFailure, ExitConfigInvalid, etc.)
  - **Metadata Access Layer**: `GetExitCodeInfo()`, `LookupExitCode()`, `ListExitCodes()` with parsed catalog data
  - **Platform Detection**: `SupportsSignalExitCodes()` with WSL detection for Windows compatibility
  - **Provenance Reporting**: `GofulmenVersion()`, `CrucibleVersion()`, `ExitCodesVersion()` using Crucible constants (no hardcoded values)
  - **Simplified Mode Mapping**: `MapToSimplified()` supporting 3-code basic mode and 8-code severity mode (catalog-derived)
  - **BSD Compatibility**: `MapToBSD()`, `MapFromBSD()`, `GetBSDCodeInfo()` for sysexits.h compatibility
  - **Snapshot Parity Test**: Automatic drift detection via `exit-codes.snapshot.json` comparison (54/54 codes verified)
  - **Efficient Implementation**: Uses `sort.Ints()` for sorting, correct YAML parsing with `maps_from` field
  - **100% Test Coverage**: Comprehensive tests for all APIs including platform detection and BSD mapping
  - **Files Added**: 11 new files (exit_codes.go, exit_codes_metadata.go, exit_codes_test.go, platform.go, platform_test.go, version.go, bsd.go, bsd_test.go, simplified_modes.go, simplified_modes_test.go, exit_codes_snapshot_test.go)

### Changed

- **Crucible Dependency**: Updated from v0.2.1 to v0.2.3 (adds exit codes catalog and snapshot)

## [0.1.7] - 2025-10-29

### Added

- **GitHub Actions CI Workflow** - Automated testing on Go 1.21, 1.22, 1.23
  - Runs tests, lint, and build on every push/PR
  - Bootstrap installs goneat v0.3.2 from public GitHub releases
  - External install test job (disabled until repo is public)
  - All dependencies are public and accessible

### Fixed

- **Prometheus Metric Naming** - Removed automatic suffix duplication in telemetry/exporters
  - Fixed `writeCounterMetrics()` and `writeGaugeMetrics()` to not append `_total` and `_gauge` suffixes
  - Metric names now follow Prometheus conventions (callers provide suffixes)
  - Fixes CI test failures: `TestPrometheusMetricTypeRouting` and `TestPrometheusMetricTypeRoutingInHandler`
- **RFC3339Nano Timestamp Test** - Fixed Go 1.21 compatibility in logging/golden_test.go
  - Use fixed timestamp with non-zero nanoseconds for consistent test behavior
  - Adjusted minimum length check to account for RFC3339Nano trailing zero omission
  - Fixes CI test failure: `TestGolden_RFC3339NanoTimestampCompatibility`

## [0.1.6] - 2025-10-29

### Added

- **Crucible v0.2.1 Config Embedding** - Clean architecture eliminating config duplication
  - **Direct Config Access**: Foundry now accesses config via `crucible.ConfigRegistry.Library().Foundry().*()` methods
  - **No More Sync Needed**: Config version automatically aligned with Crucible module version
  - **Single Source of Truth**: Removed local embedding of foundry assets, now uses Crucible's embedded config
  - **Comprehensive Tests**: Added `crucible/config_test.go` with tests that will fail if config embedding breaks
  - **Config API Re-exports**: Added `ConfigRegistry`, `GetConfig()`, and `ListConfigs()` to crucible package
  - **Architecture Cleanup**: Removed `//go:embed assets/*.yaml` from foundry, removed duplicated config files

### Changed

- **Crucible Dependency**: Updated from v0.2.0 to v0.2.1 (adds config embedding support)
- **Foundry Package**: Refactored `catalog.go` to use Crucible's embedded config instead of local assets
  - Removed `configFiles embed.FS` variable
  - Updated `loadYAML()` to call Crucible config API
  - Updated documentation to reflect new architecture
- **Build Process**: No longer requires `make sync` for foundry config (still useful for docs reference)

### Fixed

- **External Installation**: Now works correctly for forge-workhorse-groningen and other external consumers
  - Standard `go get github.com/fulmenhq/gofulmen` installation works
  - All foundry config accessible out of the box
  - No special configuration or sync required

## [0.1.5] - 2025-10-27

### Added

- **Similarity v2 API** - Unified algorithm-specific distance and score calculations (Crucible v2.0.0)
  - **DistanceWithAlgorithm()**: Calculate edit distance with algorithm selection (Levenshtein, Damerau OSA, Damerau Unrestricted)
  - **ScoreWithAlgorithm()**: Calculate normalized similarity scores with algorithm selection (all algorithms + Jaro-Winkler, Substring)
  - **5 Algorithm Support**: Levenshtein, Damerau OSA (Optimal String Alignment), Damerau Unrestricted, Jaro-Winkler, Substring matching
  - **Native OSA Implementation**: Replaced buggy matchr.OSA() with native Go implementation based on rapidfuzz-cpp
  - **100% Fixture Compliance**: All 30 Crucible v2.0.0 fixtures passing (was 28/30 with matchr bug)
  - **Cross-Language Consistency**: Validated against PyFulmen (RapidFuzz) and TSFulmen (strsim-wasm)
  - **Performance Optimized**: Native OSA expected to match Levenshtein pattern (1.24-1.76x faster than external libraries)
  - **ADR-0002**: Similarity algorithm implementation strategy with benchmark data
  - **ADR-0003**: Native OSA implementation decision and rationale

- **Similarity Telemetry** - Opt-in counter-only instrumentation (ADR-0008 Pattern 1)
  - **Counter-Only Metrics**: Algorithm usage, string length distribution, fast paths, edge cases, API misuse tracking
  - **Zero Overhead by Default**: Telemetry disabled unless explicitly enabled via EnableTelemetry()
  - **Acceptable Overhead**: ~1μs per operation when enabled (negligible for typical CLI/spell-check use cases)
  - **6 Metric Types**: distance.calls, score.calls, string_length, fast_path, edge_case, error counters
  - **Production Visibility**: Understand algorithm usage patterns and performance characteristics in production
  - **NO Histograms/Tracing**: Follows ADR-0008 Pattern 1 for performance-sensitive hot-loop code

- **Error Handling Module** - Structured error envelopes with validation and strategies
  - **Error Envelope System**: Structured errors with severity levels, correlation IDs, and context support
  - **Validation Strategies**: Configurable error handling (LogWarning, AppendToMessage, FailFast, Silent)
  - **Severity Levels**: Info, Low, Medium, High, Critical with automatic validation
  - **Context Support**: Structured context data with validation for error metadata
  - **Safe Helpers**: Production-safe wrappers that handle validation errors gracefully
  - **Cross-Language Patterns**: Consistent error handling patterns for TS/Py team implementation
  - **Enterprise Integration**: Ready for distributed tracing and error monitoring systems

- **Telemetry Phase 5** - Advanced Features & Ecosystem Integration
  - **Gauge Metrics Support**: Real-time value metrics (CPU %, memory usage, temperature) with proper type routing
  - **Custom Exporters**: Full-featured Prometheus exporter with HTTP server and proper metric formatting
  - **Metric Type Routing**: MetricsEvent carries Type field ensuring counters→Counter, gauges→Gauge, histograms→Histogram
  - **Prometheus Format Compliance**: Counters as \_total, gauges as \_gauge, histograms with full \_bucket/\_sum/\_count series
  - **+Inf Histogram Bucket**: ADR-0007 buckets plus +Inf ensuring long-running samples aren't lost
  - **Cross-Language Patterns**: Implementation ready for TS/Py teams with consistent APIs and error handling
  - **Enterprise Integration**: Ready for Prometheus, Datadog, and other monitoring systems
  - **Comprehensive Testing**: New routing tests verify correct metric type handling and Prometheus output format
  - **Performance Maintained**: <5% overhead target achieved with efficient data structures
  - **Schema Validation**: Metrics validate against canonical Crucible taxonomy and observability schemas

### Changed

- **Similarity Algorithm Strategy**: Updated to use native OSA instead of matchr.OSA() (bug fix)
- **Telemetry Core**: Enhanced MetricsEvent with Type field for proper metric routing
- **Prometheus Exporter**: Rewritten to handle all metric types with proper Prometheus conventions
- **Histogram Implementation**: Added +Inf bucket for complete sample coverage

### Fixed

- **Similarity OSA Bug**: Native OSA implementation fixes start-of-string transposition bugs in matchr library
  - `"hello"/"ehllo"`: Now correctly returns distance=1 (was 2 with matchr)
  - `"algorithm"/"lagorithm"`: Now correctly returns distance=1 (was 2 with matchr)
  - Validated against PyFulmen using RapidFuzz (canonical Rust strsim-rs implementation)
- **Metric Type Routing**: Resolved issue where gauges were incorrectly routed to counter methods
- **Histogram Buckets**: Fixed sample loss for durations exceeding largest ADR-0007 boundary
- **Prometheus Output**: Correct metric naming conventions (\_total, \_gauge, \_bucket/\_sum/\_count)

## [0.1.4] - 2025-10-23

### Added

- **FulHash** package - Enterprise-grade hashing utilities with xxh3-128 and sha256 support
  - Block and streaming hash APIs with `Hash()`, `HashString()`, `HashReader()`, and `NewHasher()`
  - Digest metadata format (`<algorithm>:<hex>`) with `FormatDigest()` and `ParseDigest()` helpers
  - Thread-safe streaming hashers for large file processing
  - Comprehensive test coverage (90%+) with Crucible fixture validation
  - Performance optimized: xxh3-128 for speed, sha256 for cryptographic needs
  - Schema-backed fixture validation against synced Crucible standards
  - Cross-language API parity with pyfulmen and tsfulmen
- **Pathfinder checksum support** - Opt-in checksum calculation using FulHash package
  - New `CalculateChecksums` and `ChecksumAlgorithm` options in `FindQuery`
  - Supports xxh3-128 (default) and sha256 algorithms
  - Maintains <10% performance overhead with streaming implementation
  - Schema-compliant metadata: `checksum`, `checksumAlgorithm`, `checksumError`
  - Backward compatible - disabled by default
  - Enables integrity verification for downstream tooling (goneat, docscribe)
- **Pathfinder feature enhancements** - Pre-migration upscale for goneat v0.2.0 compatibility
  - `PathTransform` function type for result remapping (strip prefixes, flatten paths, logical mounting)
  - `FindStream()` method for channel-based streaming results (Go-idiomatic, memory efficient)
  - `SkipDirs` option for simple directory exclusions without complex glob patterns
  - `SizeRange` and `TimeRange` filtering for file size and modification time constraints
  - `FindFilesByType()` utility method supporting common file types (go, javascript, python, java, config, docs, images)
  - `GetDirectoryStats()` for repository analytics and disk usage reporting
  - Worker pool support for concurrent directory traversal (configurable `Workers` field)
  - All features backward compatible with existing API
- **Code quality improvements** - Comprehensive polish addressing 37 assessment issues
  - Fixed high-priority security issue (G304 potential file inclusion in pathfinder)
  - Resolved 33 golangci-lint issues across bootstrap and test files
  - Updated goneat tools configuration with proper schema validation
  - Improved error handling in bootstrap package (unchecked Close() and error handling)
  - Added proper date to CHANGELOG.md entries
  - All precommit/prepush checks now pass with 100% health

### Fixed

- **Pathfinder security** - Resolved G304 potential file inclusion vulnerability in finder.go
- **Bootstrap reliability** - Fixed unchecked Close() errors and improved error handling in download/extract operations
- **Test code quality** - Addressed lint issues in foundry and config test files

### Security

- **Pathfinder** - Fixed G304 potential file inclusion via path traversal protection

## [0.1.3] - 2025-10-22

### Added

- **Similarity** package - Text similarity scoring and normalization utilities (foundry/similarity)
  - Levenshtein distance calculation with Wagner-Fischer algorithm (Unicode-aware via rune counting)
  - Normalized similarity scoring (0.0 to 1.0 range) with score formula: 1 - distance/max(len(a), len(b))
  - Suggestion API with ranked fuzzy matching and configurable thresholds (MinScore=0.6, MaxSuggestions=3)
  - Unicode-aware text normalization (trim → casefold → optional accent stripping)
  - Turkish locale support for dotted/dotless i case folding (İ→i, I→ı)
  - Accent stripping via NFD decomposition and combining mark filtering
  - Case-insensitive comparison helpers (EqualsIgnoreCase, Casefold, StripAccents)
  - Performance: ~28 μs/op for 128-char strings (17x faster than 0.5ms target)
  - 100% test coverage on all implementation files (similarity.go, normalize.go, suggest.go)
  - 15 benchmark functions for ongoing performance regression testing
  - 26 Crucible fixture tests passing (distance, normalization, suggestions)
  - Cross-language API parity with pyfulmen and tsfulmen
  - golang.org/x/text v0.30.0 dependency for Unicode normalization
- **Docscribe** package - Lightweight markdown and YAML documentation processing
  - Frontmatter extraction and metadata parsing (YAML frontmatter with error recovery)
  - Markdown header extraction with anchors and line numbers (ATX and Setext styles)
  - Format detection (markdown, YAML, JSON, TOML) with heuristic-based sniffing
  - Multi-document splitting (YAML streams, concatenated markdown with double-delimiter handling)
  - Document inspection with <1ms performance target for 100KB documents
  - Source-agnostic design works with Crucible, Cosmography, or any content source
  - Integrates with crucible.GetDoc() for SSOT asset access
  - Performance targets: InspectDocument <1ms, ParseFrontmatter <5ms, SplitDocuments <10ms
  - Comprehensive test coverage (14 test functions, 56 assertions)
  - Test fixtures for frontmatter, headers, format detection, and multi-doc scenarios
- **Crucible Overview** section added to README.md per helper library standard
  - Explains what Crucible is and why the shim/docscribe module matters
  - Provides learning resources for SSOT relationship
- **Crucible SSOT Sync** to version 2025.10.2
  - Similarity module with fixtures, schema, and standard documentation
  - Docscribe module manifest entry and standard documentation
  - Updated helper library standard with Crucible Overview requirement
  - Added Fulmen Forge workhorse standard
  - Module catalog updates

### Fixed

- **Docscribe** - Double-delimiter pattern handling (---\n\n---) for proper markdown document separation in multi-document splits

## [0.1.2] - 2025-10-20

### Added

- **Logging** - Progressive logging system with profiles, middleware pipeline, and policy enforcement
  - **Progressive Profiles**: SIMPLE (minimal), STRUCTURED (JSON + middleware), ENTERPRISE (full observability), CUSTOM (flexible)
  - **Middleware Pipeline**: Pluggable event processing with correlation, redaction, and throttling middleware
  - **Correlation Middleware**: UUIDv7 correlation ID injection for distributed tracing
  - **Redaction Middleware**: Pattern-based PII and secrets redaction
  - **Throttling Middleware**: Token bucket rate limiting with configurable drop policies
  - **Policy Enforcement**: YAML-based logging governance with environment-specific rules
  - **Config Normalization**: Case-insensitive profiles, automatic defaults, middleware deduplication
  - **Full Crucible Envelope**: 20+ fields including traceId, spanId, contextId, requestId
  - **Profile Validation**: Strict requirements for middleware, throttling, and policy enforcement
  - **Integration Tests**: End-to-end testing for all profiles and middleware combinations
  - **Golden Tests**: Cross-language compatibility validation (pyfulmen/tsfulmen alignment)
  - **Godoc Examples**: 10+ comprehensive examples for all major features
  - **Test Coverage**: 89.2% with 190+ tests
  - **Documentation**: Complete progressive logging guide with migration paths
- **Schema** - Catalog metadata, offline metaschema validation, structured diagnostics, shared validator cache, composition/diff helpers, and CLI shim with optional goneat bridge
- **Config** - Three-layer configuration loader (defaults → user → runtime) with schema validation helpers and environment override parsing

### Fixed

- **Pathfinder** - Root boundary enforcement preventing path traversal via glob patterns
- **Pathfinder** - Hidden file filtering now checks all path segments, not just basename
- **Pathfinder** - Metadata now populated with file size and mtime (RFC3339Nano)
- **Pathfinder** - .fulmenignore support with gitignore-style pattern matching

### Security

- **Pathfinder** - Path traversal protection via ValidatePathWithinRoot preventing escapes through patterns like `../**/*.go`

## [0.1.1] - 2025-10-17

### Added

- **Foundry** package - Enterprise-grade foundation utilities from Crucible catalogs
  - RFC3339Nano timestamps with nanosecond precision for cross-language compatibility
  - UUIDv7 correlation IDs for distributed tracing (time-sortable, globally unique)
  - Pattern matching (regex, glob, literal) from Crucible catalogs (20+ curated patterns)
  - MIME type detection (content-based sniffing + extension lookup)
  - HTTP status helpers (status code grouping, validation, reason phrases)
  - Country code validation (ISO 3166-1 Alpha2, Alpha3, Numeric with case-insensitive lookups)
  - Embedded catalogs for offline operation (no network dependencies)
  - Test coverage: 85.8%
- **Guardian hooks integration** with goneat for commit/push protection
  - Pre-commit and pre-push hooks with browser-based approval workflow
  - Scoped operation policies (git.commit, git.push)
  - Expiring approval sessions (5-15 minutes)
  - .goneat/hooks.yaml manifest with make target integration
- **ADR system** synchronized with Crucible ecosystem standards
  - Two-tier ADR structure (project ADRs vs. ecosystem ADRs)
  - Ecosystem ADRs synced from Crucible (ADR-0002 through ADR-0005)
- Comprehensive test and coverage improvements
  - Overall coverage: 87.7% → 89.3% (+1.6pp)
  - Schema package tests (9.9% → 53.5%, +43.6pp)
  - Logging package tests (50.8% → 54.2%, +3.4pp)
  - Foundry catalog verification tests per Foundry Interfaces standard
  - Added schema/loader_test.go and schema/errors_test.go
  - Added testdata fixtures for schema validation
  - New logging method tests (Trace, Debug, Named, WithContext)
- Documentation enhancements
  - docs/development/bootstrap.md - Complete v0.1.0-v0.1.1 journey documentation
  - docs/development/operations.md - Version management clarifications
  - AGENTS.md - Commit message style guidance with Crucible SOP cross-links
  - .plans/active/v0.1.2/pathfinder-improvements.md - Feature brief for pathfinder

### Fixed

- **Bootstrap symlink creation** - installLink() now creates proper symlinks instead of copies
  - Replaced io.Copy with os.Symlink for type:link tools
  - bin/goneat now correctly tracks source without re-bootstrap

### Security

- **Foundry security audit remediation** - Resolved all audit findings
  - Country code: Added explicit ASCII-range validation and uppercase enforcement
  - UUIDv7: Strict version checking with timestamp monotonicity awareness
  - Numeric: IEEE 754 special-value handling (NaN, Infinity) with epsilon tolerance
  - HTTP: Configurable RoundTripper support for request interception

### Changed

- **Crucible sync** to 2025.10.2 release
  - Updated standards, SOPs, and schemas
  - Added assessment schemas (severity definitions)
  - Added version policy schema for goneat
  - Enhanced agentic attribution standard with commit message style guidance

## [0.1.0] - 2025-10-13

### Added

- Initial bootstrap of gofulmen library with 7 core packages
- **Bootstrap** package - goneat installation with download, link, and verify methods
- **Config** package - XDG Base Directory support and Fulmen configuration paths
- **Logging** package - Structured logging with RFC3339Nano timestamps and severity filtering
- **Schema** package - JSON Schema validation (draft 2020-12) with Crucible integration
- **Crucible** package - Embedded access to SSOT assets (docs, schemas, config)
- **Pathfinder** package - Safe filesystem discovery with path traversal prevention
- **ASCII** package - Terminal utilities, box drawing, Unicode analysis, and terminal overrides
- Comprehensive test coverage (85% average across all packages)
- Goneat integration for SSOT sync and version management
- Documentation, operational runbooks, and MIT licensing
- Repository safety protocols and agent attribution standards
