# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.21] - 2025-12-13

### Signed-Tag Releases + sfetch Trust Anchor (Crucible v0.2.21)

**Release Type**: Release Process + Tooling Hardening  
**Status**: ✅ Ready for Release

#### Overview

This release upgrades gofulmen’s bootstrap and release workflow for higher supply-chain assurance:

- `make bootstrap` now uses `sfetch` as the trust anchor and installs `goneat` via `sfetch` with **minisign required**.
- Releases are now standardized as **GPG-signed annotated git tags** (no binaries), with a provenance check and a
  guard that enforces `tag == v$(cat VERSION)`.

#### Highlights

- **Bootstrap Trust Pyramid** – `make bootstrap` verifies `sfetch` and prints `sfetch --self-verify --json`, then installs `goneat` via `sfetch --require-minisign`.
- **No Local goneat Assumptions** – Make targets resolve `goneat` from `PATH`/user bin dir (no `./bin/goneat`).
- **Signed Tag Workflow** – New `make release-tag` + `make release-verify-tag` and a contributor-friendly `RELEASE_CHECKLIST.md`.
- **Version/Tag Guard** – New `make release-guard-tag-version` to fail when `VERSION` and release tag diverge.
- **Provenance Visibility** – New `make release-provenance-check` prints Crucible provenance from `.goneat/ssot/provenance.json` and `.crucible/metadata/metadata.yaml`.

#### Testing

- ✅ `make bootstrap` (installs/validates `sfetch`, installs signed `goneat`)
- ✅ `make sync` (provenance unchanged for Crucible v0.2.21)
- ✅ `make tools`
- ✅ `make -n release-*` (Makefile sanity)

---

## [0.1.20] - 2025-11-26

### FulHash CRC + MultiHash + Verify (Crucible v0.2.20)

**Release Type**: Feature + Dependency Sync  
**Status**: ✅ Ready for Release

#### Overview

This release finishes the FulHash workstream: CRC32/CRC32C support, single-pass MultiHash helpers, Crucible-compatible Verify utilities, and refreshed telemetry. All APIs pull algorithms directly from Crucible v0.2.20 so FulHash stays SSOT-aligned with pyfulmen and tsfulmen.

#### Highlights

- **CRC Algorithms** – Added IEEE and Castagnoli CRC32 implementations across block, streaming, and pooled hashers with fixture coverage and streaming parity tests.
- **MultiHash Helpers** – `MultiHash`, `MultiHashString`, and `MultiHashReader` dedupe algorithms, fan out via `io.MultiWriter`, emit per-algorithm counters, and record bytes only once.
- **Verify Helpers** – `Verify`, `VerifyString`, and `VerifyReader` parse Crucible-formatted digests, compute the required algorithm once, and emit `result=match|mismatch` telemetry plus mismatch counters.
- **Crucible Digest Interop** – `Digest.ToCrucible()`/`FromCrucible()` bridge SSOT digests, decoding hex when Crucible omits raw bytes and round-tripping every algorithm in regression tests.
- **Telemetry** – New counters for CRC32/CRC32C, verify result tagging, and mismatch error counters so dashboards surface digest drift quickly.

#### Files Changed

```
fulhash/*                    # CRC hashers, multi-hash fanout, verify helpers, options, tests, benches
telemetry/metrics/*          # CRC metrics + taxonomy tests
schemas/config/docs          # Crucible v0.2.20 fulhash taxonomy + documentation sync
.crucible/.goneat/VERSION    # Provenance + version bump aligned with Crucible sync
go.mod / go.sum              # Dependency update to github.com/fulmenhq/crucible v0.2.20
```

#### Testing

- ✅ `make check-all`
- ✅ `go test ./fulhash -run .`
- ✅ `go test ./...` (implicit via make target)

#### Impact

- FulHash consumers now have CRC32/CRC32C plus helper utilities without double-reading data.
- Telemetry dashboards can differentiate successes vs mismatches with algorithm tags.
- Future FulHash algorithms ship automatically via Crucible taxonomy imports.

---

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

## Archived Releases

Older release notes are archived under `docs/releases/`. Refer to those files for versions prior to v0.1.18.

---

**Note**: For complete release documentation of archived releases, see the individual release files in `docs/releases/`.
