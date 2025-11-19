# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.19] - 2025-11-19

### Crucible Version Synchronization Fix + Guardrail

**Release Type**: Critical Bug Fix + Process Improvement  
**Status**: ✅ Ready for Release

#### Overview

This release fixes the v0.1.18 Crucible version mismatch and implements guardrails to prevent future occurrences. The mismatch caused `go.mod` to require v0.2.18 while embedded assets came from v0.2.19, primarily affecting users of Crucible's DevSecOps secrets schema.

#### Critical Fix

**Crucible Version Mismatch (v0.1.18)**:

- **Issue**: go.mod required `github.com/fulmenhq/crucible v0.2.18` but synced assets were from v0.2.19
- **Impact**: Version reporting incorrect, DevSecOps secrets schema users got v0.2.19 docs but v0.2.18 runtime
- **Root Cause**: Sync process updated assets but forgot to run `go get github.com/fulmenhq/crucible@v0.2.19`
- **Fix**: Updated go.mod to v0.2.19, aligned all three synchronization points (ssot-consumer.yaml, provenance.json, go.mod)

#### Guardrails Implemented

**Automated Detection** - `TestCrucibleVersionMatchesMetadata`:

- New test fails CI/builds when go.mod and metadata versions disagree
- Parses go.mod using `golang.org/x/mod/modfile` (no shell dependencies)
- Parses metadata.yaml to extract synced Crucible version
- Normalizes versions to handle format differences ("v0.2.19" vs "0.2.19")
- Beautiful error message using `ascii.DrawBox()` shows exact mismatch and fix
- Dogfoods gofulmen: uses `pathfinder.FindRepositoryRoot()` for repo discovery
- Runs as part of `make check-all`, `make test`, and CI

**Automated Workflow** - `make crucible-update VERSION=v0.2.X`:

- Single command atomically updates all three synchronization points:
  1. Updates `.goneat/ssot-consumer.yaml` ref via sed
  2. Runs `make sync` to update provenance timestamp
  3. Runs `go get github.com/fulmenhq/crucible@<version>` + `go mod tidy`
  4. Runs verification test to confirm success
- Self-documenting with progress messages and next steps
- Prevents partial updates that cause mismatches

**Manual Verification Guide** (ADR-0007):

Quick 3-point check for code reviewers:

```bash
# 1. Check sync ref
grep "ref:" .goneat/ssot-consumer.yaml  # Expected: ref: v0.2.19

# 2. Check go.mod
grep "github.com/fulmenhq/crucible" go.mod  # Expected: v0.2.19

# 3. Check metadata
grep "version:" .crucible/metadata/metadata.yaml | head -2  # Expected: 0.2.19
```

#### Changes

**Fixed**:

- Crucible version mismatch: go.mod v0.2.18 → v0.2.19 (aligns with embedded assets)
- Updated provenance timestamp to reflect current sync state

**Added**:

- `crucible/version_guard_test.go` - Guardrail test (uses pathfinder + ASCII)
- `make crucible-update` - Atomic Crucible update workflow
- `docs/development/adr/ADR-0007-crucible-version-synchronization.md` - Process documentation
- Dependency: `golang.org/x/mod v0.30.0` for go.mod parsing

#### Files Changed

```
crucible/version_guard_test.go                     # NEW: 152 lines - Guardrail test
docs/development/adr/ADR-0007-crucible-version-synchronization.md  # NEW: 350 lines - ADR
Makefile                                           # +30 lines: crucible-update target
go.mod                                             # crucible v0.2.18 → v0.2.19, +golang.org/x/mod
go.sum                                             # Updated checksums
.goneat/ssot/provenance.json                       # Updated timestamp
.crucible/metadata/metadata.yaml                   # Updated timestamp
VERSION                                            # v0.1.19
docs/crucible-go/guides/consuming-crucible-assets.md  # +112 lines: Practical examples
```

**Total**: 9 files changed, +644 insertions, -24 deletions (2 new files, 7 updates)

#### Testing

- ✅ `TestCrucibleVersionMatchesMetadata` PASSES (versions now match)
- ✅ `make check-all` PASSES (all quality checks)
- ✅ No regressions in test suite
- ✅ Verified manual 3-point check shows alignment

#### Impact

**For All Users**:

- ✅ Correct Crucible version reporting via `crucible.GetVersionString()`
- ✅ Runtime behavior matches embedded documentation
- ✅ Future releases protected by guardrail test

**For Contributors**:

- ✅ Simple workflow: `make crucible-update VERSION=v0.2.X`
- ✅ Automated verification catches mistakes before commit
- ✅ Clear documentation in ADR-0007

#### Lessons Learned

This is the **second time** this mistake was made, proving that:

1. **Process > Memory**: Automated workflows prevent human error better than documentation
2. **Fail Fast**: Tests that catch mistakes before release are invaluable
3. **Dogfooding Works**: Using our own libraries (pathfinder, ASCII) improved test quality

---

## [0.1.18] - 2025-11-19

### Known Issues

⚠️ **Version mismatch bug**: This release has mismatched Crucible versions (go.mod requires v0.2.18 but embedded docs/schemas are v0.2.19). This primarily affects users of Crucible's DevSecOps secrets schema. Upgrade to v0.1.19 for correct alignment.

### Crucible v0.2.19 Sync – DevSecOps Secrets Schema Hardening

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

## [0.1.15] - 2025-11-16 (Archived)

Logging Redaction Middleware + Pathfinder Repository Root Discovery. See `docs/releases/v0.1.15.md`

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
