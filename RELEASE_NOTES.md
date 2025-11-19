# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.18] - 2025-11-19

### Crucible v0.2.19 Sync – DevSecOps Secrets Schema Hardening

**Release Type**: Dependency Update (Crucible SSOT Sync)  
**Status**: ✅ Ready for Release

#### Overview

This release syncs gofulmen to Crucible v0.2.19, which introduces comprehensive DevSecOps secrets schema hardening with DoS protection, structured credentials, and enhanced metadata support. No gofulmen code changes beyond the SSOT sync.

#### Changes

**DevSecOps Secrets Schema Hardening** (Primary Update):

- **DoS Protection**: Defensive size limits to prevent resource exhaustion
  - 256 projects per file maximum
  - 1,024 credentials per project maximum
  - 65,536 character limit for credential values (64KB, UTF-8 encoded)
  - 2,048 character limit for external references (vault URIs, ARNs)
  - 4,096 character limit for descriptions (file, project, credential levels)
  - 255 character limit for environment variable names (POSIX standard)
- **Structured Credentials**: Migrated from flat `secrets` (string values) to `credentials` (objects)
  - Type field: `api_key`, `password`, or `token` (determines masking behavior)
  - Value field: Plaintext credential value (mutually exclusive with `ref`)
  - Ref field: External reference for vault integration (mutually exclusive with `value`)
  - Description field: Audit-friendly documentation for each credential
- **Enhanced Metadata**:
  - Global `env_prefix` field for all projects (e.g., `MYAPP_`)
  - Per-project `env_prefix` override capability
  - Description fields at file, project, and credential levels (compliance documentation)
- **Improved Patterns**:
  - Enhanced `project_slug` pattern: Now allows underscores alongside hyphens (`my_service-v2`)
  - Start/end must be alphanumeric: `^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`

**Additional Updates**:

- Updated telemetry metrics taxonomy with latest definitions
- Updated metrics documentation with enhanced module standards

#### Files Changed

```
.crucible/metadata/metadata.yaml                   # Updated metadata
.goneat/ssot-consumer.yaml                         # Updated to v0.2.19 ref
.goneat/ssot/provenance.json                       # Updated provenance (commit f17e5fa)
VERSION                                            # v0.1.18
config/crucible-go/devsecops/secrets/v1.0.0/defaults.yaml         # Enhanced with structured credentials
config/crucible-go/taxonomy/metrics.yaml                          # Updated taxonomy
docs/crucible-go/standards/devsecops/project-secrets.md           # +348 lines: Size limits, credential objects
docs/crucible-go/standards/library/modules/telemetry-metrics.md  # +552 lines: Enhanced documentation
schemas/crucible-go/devsecops/secrets/v1.0.0/secrets.schema.json # +358 lines: Hardened schema
```

**Total**: 9 files changed, +1424 insertions, -179 deletions

#### Impact

**For Secrets Management Users**:

- ✅ Enhanced security with DoS protection limits
- ✅ Structured credentials enable type-aware masking
- ✅ External reference support for vault/secrets-manager integration
- ✅ Compliance-friendly with description fields at all levels
- ⚠️ Schema changes require update to fulmen-secrets v0.1.1+ (if using secrets tooling)

**For All Users**:

- ✅ No breaking changes to gofulmen APIs
- ✅ Updated Crucible standards available via `crucible` package
- ✅ Enhanced documentation for DevSecOps workflows

#### Verification

- ✅ All tests passing (no code changes, sync only)
- ✅ `make check-all`: 100% health (0 issues)
- ✅ Crucible provenance confirmed: commit f17e5fa (v0.2.19)
- ✅ Schema validation: All embedded schemas valid

---

## [0.1.17] - 2025-11-17

### HTTP Server Metrics Middleware – Production-Ready Performance

**Release Type**: New Feature + Performance Optimization  
**Status**: ✅ Ready for Release

#### Overview

This release introduces comprehensive HTTP server metrics middleware with production-ready performance (~21% overhead) and proper cardinality control. The implementation provides all 5 HTTP metrics from the Crucible v0.2.18 taxonomy with framework integration support and enterprise-grade reliability.

#### Key Features

**Complete HTTP Metrics Collection**:

- All 5 HTTP metrics: requests total, duration, request size, response size, active requests
- Proper histogram bucket mathematics for size metrics
- Minimal label sets for cardinality control per taxonomy
- Thread-safe concurrent operation with atomic counters

**Route Normalization & Cardinality Control**:

- UUID segment normalization: `/api/users/550e8400-e29b-41d4-a716-446655440000` → `/api/users/{uuid}`
- Numeric segment normalization: `/users/123` → `/users/{id}`
- Query parameter stripping: `/api/search?q=test` → `/api/search`
- Configurable custom route normalizers

**Performance Optimization**:

- **~21% overhead** (reduced from 55-84% during development)
- Tag pooling using `sync.Pool` to reduce allocations
- Histogram bucket pooling for size metrics
- Pre-compiled UUID regex patterns
- Optimized fast-path handling for simple routes

**Framework Integration**:

- Native net/http support
- Chi router integration patterns
- Gin router integration patterns
- Easy middleware composition

#### Critical Fixes

**UUID Normalization Bug**:

- **Issue**: Route normalizer checked for UUID patterns but only replaced hardcoded UUID string
- **Impact**: Real UUIDs would slip through unchanged, causing cardinality explosion
- **Fix**: Implemented proper regex replacement with `uuidPattern.ReplaceAllString(path, "{uuid}")`
- **Result**: All UUID segments now correctly normalized to prevent metric explosion

**Duration Buckets API Cleanup**:

- **Issue**: `DurationBuckets` option was settable but never used (emitter-driven)
- **Impact**: Misleading API suggesting configurable duration buckets
- **Fix**: Removed unused option and renamed `WithCustomBuckets()` to `WithCustomSizeBuckets()`
- **Result**: Honest API design with clear emitter-driven behavior documentation

#### API Changes

```go
// New HTTP metrics middleware
middleware := telemetry.HTTPMetricsMiddleware(
    emitter,
    telemetry.WithServiceName("my-api"),
    telemetry.WithCustomSizeBuckets([]float64{1024, 10240, 102400, 1048576}),
)

// Framework integration examples documented in README.md
```

**Removed**:

- `WithCustomBuckets()` function (replaced with `WithCustomSizeBuckets()`)
- `DurationBuckets` field from `HTTPMetricsConfig`
- `DefaultHTTPDurationBuckets` constant

#### Performance Benchmarks

```
Baseline (no middleware):     ~1336 ns/op
With HTTP metrics middleware: ~1726 ns/op
Overhead: ~21% (390ns additional)
Memory: ~5.5KB per request
Allocations: ~22 per request
```

#### Testing & Quality

- **Comprehensive test coverage**: All HTTP metrics, route normalization, framework integration
- **Performance validation**: Hotspot analysis and optimization verification
- **Cross-language consistency**: 95% alignment with expected patterns
- **Schema compliance**: Full Crucible v0.2.18 taxonomy alignment
- **Framework validation**: Chi and Gin integration patterns tested

#### Documentation

- Updated `telemetry/README.md` with HTTP metrics section and performance claims
- Added `telemetry/HTTP_METRICS_MIGRATION.md` with comprehensive migration guide
- Updated `telemetry/CROSS_LANGUAGE_CONSISTENCY.md` with performance analysis
- Framework integration examples and best practices included

#### Migration

No breaking changes for existing users. New HTTP metrics functionality is opt-in:

```go
// Add to existing HTTP server
middleware := telemetry.HTTPMetricsMiddleware(emitter)
handler := middleware(existingHandler)
```

#### Dependencies

No new dependencies added. Uses existing telemetry infrastructure.

---

## [0.1.15] - 2025-11-16

### Logging Redaction Middleware + Pathfinder Repository Root Discovery

**Release Type**: Major Feature Release  
**Status**: ✅ Ready for Release

#### Overview

This release adds schema-compliant logging redaction middleware for PII/secrets protection and repository root discovery for pathfinder. The redaction middleware enables automatic filtering of sensitive data in logs per Crucible v0.2.16, while repository root discovery fixes schema validation from test subdirectories and provides a foundation for tooling that needs repository context.

#### Changes

**Logging Redaction Middleware**:

- **Schema-Compliant Redaction**: Implements Crucible v0.2.16 logging middleware specification
  - Pattern-based filtering: API keys, tokens, passwords, SSNs, credit cards (regex)
  - Field-based filtering: password, token, secret, apiKey, ssn, creditCard
  - Replacement modes: `[REDACTED]` (text) or SHA-256 hash prefix
  - Opt-in design: No behavioral changes unless explicitly configured
- **Helper Functions**: `WithRedaction()`, `WithDefaultRedaction()` for easy configuration
- **Bundle Helpers**: `BundleSimpleWithRedaction()`, `BundleStructuredWithRedaction()` for common patterns
- **Backward Compatibility**: Pipeline builder maps legacy `name`→`type`, `order`→`priority` fields
- **Documentation**: 80+ lines in logging/README.md with before/after examples

**Pathfinder Repository Root Discovery**:

- **FindRepositoryRoot() API**: Safe upward traversal per Crucible v0.2.15 specification
  - Predefined markers: Git (`.git`), Go (`go.mod`), Node (`package.json`), Python (`pyproject.toml`), Monorepo (`pnpm-workspace.yaml`)
  - Safety boundaries: Home directory ceiling (default), configurable boundaries, max depth (10)
  - Filesystem root detection: `/` (Unix), `C:\` (Windows), UNC paths (`\\server\share`)
- **Security Features**:
  - Symlink loop detection with `TRAVERSAL_LOOP` error (critical severity)
  - Boundary enforcement prevents traversal outside designated areas
  - Multi-tenant isolation (cross-tenant data access prevention)
  - Container escape prevention
- **Functional Options**: `WithMaxDepth`, `WithBoundary`, `WithFollowSymlinks`, `WithMarkers`, `WithStrictBoundary`
- **Performance**: All operations <30µs (well under Crucible spec targets of <5-20ms)
- **Test Coverage**: 36 tests (9 basic, 17 security, 10 benchmarks)
- **Documentation**: 150+ lines in pathfinder/README.md with usage examples

**Schema Validator Fixes**:

- **Repository Root Resolution**: Added `findRepoRoot()` helper to compute paths from repository root
- **Fixed Path Mapping**: `mapSchemaURLToPath()` now handles relative schema references correctly
- **Version Directory Detection**: Prevents duplicate path segments (e.g., `/v1.0.0/v1.0.0/`)
- **Subdirectory Testing**: All schema validation tests pass from any subdirectory

**Crucible v0.2.16 Update**:

- Updated logging schemas with middleware `type` and `priority` fields
- Added pathfinder repository root discovery specification
- New ADR-0012: Schema reference IDs standard
- Updated DevSecOps taxonomy with modules schema
- Updated metrics taxonomy

#### Impact

**For Logging Users**:

- ✅ Automatic PII/secrets redaction available (opt-in)
- ✅ Schema-compliant middleware configuration
- ✅ 100% backward compatibility with existing configurations
- ✅ Default patterns cover common sensitive data (API keys, passwords, SSNs, credit cards)

**For Pathfinder Users**:

- ✅ Repository root discovery for tooling that needs repository context
- ✅ Safe upward traversal with multiple safety boundaries
- ✅ Cross-language parity with tsfulmen v0.1.9 (symlink loop detection)
- ✅ Security-first design prevents data leakage

**For Schema Validator Users**:

- ✅ Schema validation works correctly from test subdirectories
- ✅ Relative schema references resolve properly
- ✅ Version directory detection prevents path issues

#### Files Changed

```
logging/middleware_redaction_v2.go       # NEW: 241 lines - Redaction middleware
logging/helpers.go                        # NEW: 78 lines - Helper functions
logging/config.go                         # Updated: +25 lines - RedactionConfig, MiddlewareConfig updates
logging/logger.go                         # Updated: +56 lines - Pipeline builder compatibility
logging/logger_test.go                    # Updated: +7 lines - Test updates
logging/config_test.go                    # Updated: +4 lines - Config test updates
logging/README.md                         # Updated: +150 lines - Redaction docs

pathfinder/repo_root.go                   # NEW: 385 lines - FindRepositoryRoot implementation
pathfinder/repo_root_test.go              # NEW: 147 lines - Basic functionality tests
pathfinder/repo_root_security_test.go     # NEW: 497 lines - Security test suite
pathfinder/repo_root_bench_test.go        # NEW: 217 lines - Performance benchmarks
pathfinder/README.md                      # Updated: +168 lines - Repository root docs

schema/validator.go                       # Updated: +173 lines - Path resolution fixes

.goneat/ssot-consumer.yaml                # Updated to v0.2.16 ref
.goneat/ssot/provenance.json              # Updated provenance tracking
.crucible/metadata/metadata.yaml          # Updated metadata
VERSION                                   # v0.1.15
go.mod                                    # Updated to Crucible v0.2.16
go.sum                                    # Updated with v0.2.16 hashes

docs/crucible-go/standards/observability/logging.md  # Updated: +193 lines
docs/crucible-go/standards/library/extensions/pathfinder.md  # Updated: +338 lines
docs/crucible-go/decisions/ADR-0012-schema-ref-ids.md  # NEW: 44 lines
schemas/crucible-go/logging/v1.0.0/logger-config.schema.json  # Updated: Middleware fields
schemas/crucible-go/devsecops/v1.0.0/devsecops-module-entry.schema.json  # NEW: 117 lines
config/crucible-go/taxonomy/devsecops/modules/v1.0.0/modules.yaml  # NEW: 46 lines
config/crucible-go/taxonomy/library/platform-modules/v1.0.0/modules.yaml  # Updated: +45 lines
config/crucible-go/taxonomy/metrics.yaml  # Updated: +2 lines
```

**Total**: 27 files changed, +2905 lines, -85 lines (6 new files, 21 updates)

#### Verification

- ✅ All tests passing (36 pathfinder tests + existing suite)
- ✅ Precommit checks: 0 issues (100% health)
- ✅ Cross-language audit: Aligned with tsfulmen v0.1.9 (symlink loop detection)
- ✅ Backward compatibility: 100% for logging configurations
- ✅ Performance: All pathfinder benchmarks well under spec targets
- ✅ Security: 17 security tests covering boundary enforcement, isolation, escape prevention

---

## [0.1.14] - 2025-11-15 (Archived)

Fulpack Module Complete + Crucible v0.2.14 Update. See `docs/releases/v0.1.14.md`

---

## Archived Releases

Older releases (v0.1.13 and earlier) are archived in `docs/releases/`. See those files for complete release documentation.

## [0.1.13] - 2025-11-13 (Archived)

Windows Build Compatibility & Crucible v0.2.11 Update. See `docs/releases/v0.1.13.md`

## [0.1.12] - 2025-11-10 (Archived)

Critical Dependency Fix - Updated go.mod to Crucible v0.2.9. See `docs/releases/v0.1.12.md`

## [0.1.11] - 2025-11-10 (Archived)

Crucible v0.2.9 sync with enhanced documentation. See `docs/releases/v0.1.11.md`

## [0.1.10] - 2025-11-09 (Archived)

Signals package migration to top-level. See `docs/releases/v0.1.10.md`

## [0.1.9] - 2025-11-08 (Archived)

Prometheus exporter, App Identity module, and Signal Handling module. See `docs/releases/v0.1.9.md`

## [0.1.8] - 2025-11-03 (Archived)

Schema export utilities and Foundry exit codes integration. See `docs/releases/v0.1.8.md`

## [0.1.7] - 2025-10-29 (Archived)

GitHub Actions CI infrastructure + test fixes. See `docs/releases/v0.1.7.md`

## [0.1.6] - 2025-10-29 (Archived)

Crucible v0.2.1 config embedding. See `docs/releases/v0.1.6.md`

## [0.1.5] - 2025-10-27 (Archived)

Similarity v2 API + Telemetry + Error Handling. See `docs/releases/v0.1.5.md`

## [0.1.4] - 2025-10-23 (Archived)

FulHash package + Pathfinder enhancements. See `docs/releases/v0.1.4.md`

## [0.1.3] - 2025-10-22 (Archived)

Similarity & Docscribe modules + Crucible SSOT sync. See `docs/releases/v0.1.3.md`

## [0.1.2] - 2025-10-20 (Archived)

Progressive logging + Schema + Config packages. See `docs/releases/v0.1.2.md`

## [0.1.1] - 2025-10-17 (Archived)

Foundry package + Guardian hooks. See `docs/releases/v0.1.1.md`

## [0.1.0] - 2025-10-13 (Archived)

Initial bootstrap with 7 core packages. See `docs/releases/v0.1.0.md`

---

**Note**: For complete release documentation of archived releases, see the individual release files in `docs/releases/`.
