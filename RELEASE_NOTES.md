# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.14] - 2025-11-15

### Fulpack Module Complete + Crucible v0.2.14 Update

**Release Type**: Major Feature Release + Dependency Update  
**Status**: ✅ Ready for Release

#### Overview

This release completes the fulpack archive module implementation (all 5 operations now functional) and updates to Crucible v0.2.14. The fulpack module now provides production-ready archive creation, extraction, and verification with mandatory security protections.

#### Changes

**Fulpack Archive Module - Complete Implementation**:

- **Create()**: Archive creation with pathfinder source selection and fulhash checksums
  - Supports TAR, TAR.GZ, ZIP, GZIP formats
  - Pathfinder integration for glob-based source filtering (include/exclude patterns)
  - Fulhash checksum generation (SHA-256 default, XXH3-128 for speed)
  - Configurable compression levels (1-9, ignored for TAR)
  - Proper algorithm labeling (unsupported algorithms fallback to SHA-256 with correct label)
- **Extract()**: Secure extraction with **mandatory** security protections
  - Path traversal prevention (rejects `../` and absolute paths via `isPathTraversal()`)
  - Symlink validation (ensures targets stay within destination via `isWithinBounds()`)
  - **Decompression bomb detection during extraction** (enforces compression ratio, size, and entry limits)
  - Overwrite policy support (error/skip/overwrite modes)
  - **Include/exclude pattern filtering** during extraction
  - MaxSize (1GB default) and MaxEntries (10000 default) enforcement
- **Verify()**: Integrity and security validation
  - Archive structure validation (corrupt archive detection)
  - Path traversal scanning across all entries
  - Decompression bomb characteristic detection
  - Symlink safety validation
  - Checksum presence detection
- **All 5 Operations Complete**: Info, Scan, Create, Extract, Verify
- **22 Comprehensive Tests**: Covering all formats, security scenarios, and edge cases
- **Spec Compliance**: 100% compliant with Fulpack Archive Module Standard v1.0.0

**Security Fixes (Audit Findings)**:

- Fixed exclude patterns now honored during extraction (was being ignored)
- Added decompression bomb detection to extract operation (was only in verify)
- Fixed checksum algorithm mislabeling (now correctly reports actual algorithm used)

**Crucible v0.2.14 Update**:

- **Dependency Update**: Updated with comprehensive verification process
  - Updated `go.mod` from v0.2.13 to v0.2.14 and verified via `go list -m github.com/fulmenhq/crucible`
  - Updated `.goneat/ssot-consumer.yaml` sync configuration to use v0.2.14 ref
  - Verified no vendor directory drift (clean dependency management)
  - Added DevSecOps secrets management standards (docs + schema + defaults)
  - Updated metrics taxonomy with latest definitions
  - Updated provenance tracking with latest Crucible metadata (commit 089b4c7)

#### Files Changed

```
fulpack/create.go                    # NEW: Archive creation implementation
fulpack/extract.go                   # NEW: Secure extraction implementation
fulpack/verify.go                    # NEW: Security validation implementation
fulpack/types.go                     # Added ExcludePatterns field
fulpack/api.go                       # Wired up Create/Extract/Verify operations
fulpack/fulpack_test.go              # +513 lines: Comprehensive test coverage
.goneat/ssot-consumer.yaml           # Updated to v0.2.14 ref
.goneat/ssot/provenance.json         # Updated provenance tracking (commit 089b4c7)
VERSION                              # v0.1.14
go.mod                               # Updated to Crucible v0.2.14
go.sum                               # Updated with v0.2.14 hashes
docs/crucible-go/standards/devsecops/secrets-management.md # NEW: Secrets management standard
schemas/crucible-go/devsecops/v1.0.0/secrets.schema.json  # NEW: Secrets schema
config/crucible-go/devsecops/secrets-defaults.yaml        # NEW: Secrets defaults
```

**Total**: 14 files changed (+3 new fulpack ops, +3 new DevSecOps assets, +8 updates)

---

## [0.1.13] - 2025-11-13

### Windows Build Compatibility & Crucible v0.2.11 Update

**Release Type**: Bug Fix + Dependency Update  
**Status**: ✅ Ready for Release

#### Overview

This release resolves critical Windows build failures in the signals package and updates to Crucible v0.2.11 with full verification. The Windows fix enables cross-platform compatibility while maintaining Unix functionality, and the Crucible update brings the latest fulpack type generation framework.

#### Changes

**Windows Build Fix**:

- **Platform-Specific Signal Handling**: Implemented build tag solution for cross-platform compatibility
  - Added `signals/platform_signals_unix.go` with SIGUSR1/2 definitions for Unix systems
  - Added `signals/platform_signals_windows.go` with empty map for Windows compatibility
  - Updated `signals/http.go` to use dynamic signal map composition
  - Resolved undefined symbol errors on Windows for `syscall.SIGUSR1/SIGUSR2`
  - Maintains full Unix functionality while enabling Windows builds

**Crucible v0.2.11 Update**:

- **Dependency Verification**: Updated with comprehensive verification process
  - Updated `go.mod` from v0.2.9 to v0.2.11 and verified via `go list -m github.com/fulmenhq/crucible`
  - Updated `.goneat/ssot-consumer.yaml` sync configuration to use v0.2.11 ref
  - Updated sync configuration: changed `sync_path_base` from `lang/go` to `"./"` per ADR-0004
  - Confirmed `go.sum` contains v0.2.11 hashes and removed stale vendor directory
  - Enhanced fulpack type generation framework for cross-language consistency
  - Updated provenance tracking with latest Crucible metadata (commit 631e8b7)
  - Synced latest Crucible assets to `docs/crucible-go/`, `config/crucible-go/`, `schemas/crucible-go/`

#### Impact

**For Windows Users**:

- ✅ gofulmen now builds successfully on Windows platforms
- ✅ All signals package functionality available on Windows (platform-appropriate)
- ✅ No breaking changes for Unix/Linux/macOS users

**For All Users**:

- ✅ Latest Crucible v0.2.11 assets and schemas available
- ✅ Enhanced fulpack type generation framework for future cross-language support
- ✅ Improved provenance tracking and metadata

#### Verification

- ✅ `go test ./signals/...` passes on all platforms
- ✅ Windows build compatibility confirmed
- ✅ Crucible v0.2.11 dependency verified through multiple methods
- ✅ Precommit checks pass with updated formatting
- ✅ Cross-platform signal handling tested

#### Files Changed

```
.goneat/ssot-consumer.local.yaml    # Removed: Local override (explicit coupling only)
.goneat/ssot-consumer.yaml          # Updated to v0.2.11 ref
.goneat/ssot/provenance.json        # Updated provenance tracking
go.mod                               # Updated to Crucible v0.2.11
go.sum                               # Updated with v0.2.11 hashes
signals/http.go                       # Updated for platform-specific signals
signals/platform_signals_unix.go       # New file: Unix signal definitions
signals/platform_signals_windows.go    # New file: Windows compatibility
VERSION                              # v0.1.13
```

**Total**: 8 files changed, +45 lines, -15 lines (1 removed)

---

## [0.1.12] - 2025-11-10

### Critical Dependency Fix

**Release Type**: Bug Fix  
**Status**: ✅ Ready for Release

#### Overview

This release fixes a critical dependency issue where go.mod was still referencing Crucible v0.2.8 despite documentation claiming v0.2.9 sync. Downstream teams were not receiving the correct Crucible dependency.

#### Changes

**Dependency Fix**:

- Updated go.mod from `github.com/fulmenhq/crucible v0.2.8` to `v0.2.9`
- Ensured downstream teams receive correct Crucible v0.2.9 dependency
- Maintained all v0.1.11 functionality with proper dependency resolution

**Impact**: Downstream teams should now receive the correct Crucible v0.2.9 dependency when updating gofulmen.

---

## Archived Releases

Older releases (v0.1.11 and earlier) are archived in `docs/releases/`. See those files for complete release documentation.

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
