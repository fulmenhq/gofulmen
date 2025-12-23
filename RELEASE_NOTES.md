# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.26] - 2025-12-23

### Crucible v0.2.27 sync + fixture directory markers

**Release Type**: Dependency Sync / Repo Hygiene

#### Overview

This release syncs Crucible SSOT assets to `v0.2.27` and adds `.gitkeep` marker files for intentionally empty fixture directories so they persist across clean checkouts.

#### Highlights

- **Crucible SSOT sync** – Updated embedded Crucible assets and `go.mod` dependency to Crucible `v0.2.27`.
- **Fixture markers** – Added `.gitkeep` to empty fixture directories used for edge-case tests.

#### Testing

- ✅ `go test ./...`

---

## [0.1.25] - 2025-12-20

### goneat hooks compatibility guard

**Release Type**: Tooling Hardening

#### Overview

This release adds a small guardrail to catch known `goneat hooks generate --with-guardian` template regressions before they can break local development workflows.

#### Highlights

- **Compatibility checks** – `make verify-hooks-compat` detects stray brace patterns and missing `set -f` (noglob) in generated hooks.
- **Temporary fixup** – `scripts/fixup-hooks-noglob.sh` injects `set -f` into hooks until upstream templates include it consistently.
- **Check-all integration** – `verify-hooks-compat` is part of `make check-all`.
- **Crucible SSOT sync** – Updated embedded Crucible assets and `go.mod` dependency to Crucible `v0.2.26`.

#### Testing

- ✅ `make verify-hooks-compat`
- ✅ `make precommit`

---

## [0.1.24] - 2025-12-19

### Crucible v0.2.25 sync

**Release Type**: Dependency Sync

#### Overview

This release updates embedded Crucible assets and the Crucible Go module dependency to `v0.2.25`.

#### Highlights

- **Crucible SSOT sync** – Updated embedded Crucible docs/config/schemas and `go.mod` dependency to Crucible `v0.2.25`.

#### Testing

- ✅ `go test ./...`

---

## Archived Releases

Older release notes are archived under `docs/releases/`.
