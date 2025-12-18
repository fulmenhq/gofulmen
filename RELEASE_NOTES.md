# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.22] - 2025-12-18

### AppIdentity portable discovery (binary outside repo) + Crucible v0.2.23

**Release Type**: Bug Fix / Ergonomics

#### Overview

This release closes a portability gap in `appidentity` discovery: `.fulmen/app.yaml` can now be discovered even when an installed binary is executed from a working directory outside its repository tree.

#### Highlights

- **Executable-dir fallback** – After the normal CWD/repo-root ancestor search fails with `NotFound`, discovery also searches ancestors of `os.Executable()`.
- **Env var remains authoritative** – If `FULMEN_APP_IDENTITY_PATH` is set (even to a missing path), discovery short-circuits to that result and does not fall back.
- **Better not-found diagnostics** – NotFound errors now include both primary and fallback search traces.
- **No cached NotFound** – Only successful identity loads are cached; errors are not cached so callers can retry after fixing runtime conditions.
- **Crucible SSOT sync** – Updated embedded Crucible assets and `go.mod` dependency to Crucible `v0.2.23`.

#### Testing

- ✅ `go test ./... -count=1`
- ✅ `FULMEN_APP_IDENTITY_PATH=/definitely/missing.yaml go test ./appidentity -v -count=1`

---

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
- **Signing Key Env Alignment** – Release scripts use `GOFULMEN_GPG_HOMEDIR` (deprecated alias: `GOFULMEN_GPG_HOME`).
- **Optional Minisign Attestation** – `make release-tag` can write `dist/release/vX.Y.Z.tag.txt.minisig` when `GOFULMEN_MINISIGN_KEY` + `GOFULMEN_MINISIGN_PUB` are set.
- **Version/Tag Guard** – New `make release-guard-tag-version` to fail when `VERSION` and release tag diverge.
- **Provenance Visibility** – New `make release-provenance-check` prints Crucible provenance from `.goneat/ssot/provenance.json` and `.crucible/metadata/metadata.yaml`.
- **Stable CI Repo Root Discovery** – New `pathfinder.DetectCIBoundaryHint()` helps applications constrain `FindRepositoryRoot` to the CI workspace without adding a production env bypass.

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

## Archived Releases

Older release notes are archived under `docs/releases/`.
