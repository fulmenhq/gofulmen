# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.3.1] - 2026-01-07

### Crucible v0.4.3 sync

**Release Type**: Patch Release (Dependency Update)

#### Overview

This release updates Crucible to v0.4.3, keeping gofulmen aligned with the latest SSOT assets.

#### Highlights

- **Crucible v0.4.3** – Updated embedded Crucible assets and `go.mod` dependency.

#### Testing

- `make check-all`
- `make test`

---

## [0.3.0] - 2026-01-07

### Crucible v0.4.2 sync + canonical URI resolution

**Release Type**: Minor Release (Breaking Change)

#### Overview

This release syncs Crucible v0.4.2 which establishes the Canonical URI Resolution Standard. All schema `$id` values now use module-qualified URIs (`schemas.fulmenhq.dev/crucible/...`). The schema resolver in `schema/validator.go` has been updated to explicitly handle only crucible-module schemas.

#### Highlights

- **Crucible v0.4.2** – Updated embedded Crucible assets and `go.mod` dependency.
- **Canonical URIs** – All ~63 schemas updated with `crucible/` module prefix in `$id`.
- **Resolver fix** – `localLoader.Load()` now rejects non-crucible modules with clear error.
- **Fixture Standard** – New standard for test infrastructure repositories.
- **Similarity promotion** – Fixtures and schemas moved from `library/foundry/` to `library/similarity/`.
- **Module schema v1.1.0** – New fields: `weight`, `default_inclusion`, per-language `notes`.

#### Breaking Changes

- Schema `$id` URIs changed: `schemas.fulmenhq.dev/<topic>/...` → `schemas.fulmenhq.dev/crucible/<topic>/...`
- Enact schemas (11 files) removed – moved to enacthq organization.
- Goneat schemas (6 files) removed – moved to goneat repository.
- Similarity paths changed: `library/foundry/` → `library/similarity/`.

The Go API (`foundry/similarity` package) and schema validation are unchanged.

#### Testing

- `make check-all`
- `make test`

---

## [0.2.1] - 2026-01-03

### Pre-push license compliance via goneat assess

**Release Type**: DX Improvement

#### Overview

This release closes a workflow gap where license violations could slip through if developers only ran `make prepush` or relied on git hooks. The `dependencies` category is now included in the pre-push hook, ensuring `goneat assess --categories dependencies` runs automatically and catches forbidden licenses (GPL, LGPL, AGPL, MPL, CDDL) before code reaches the remote.

#### Highlights

- **Pre-push hook updated** – `.goneat/hooks.yaml` pre-push now includes `format,lint,security,dependencies`.
- **Automatic license checks** – No manual `make license-audit` required; dependencies category runs via assess.
- **Unified workflow** – License compliance is now part of the standard goneat assess pipeline.
- **Crucible v0.3.1** – Updated embedded Crucible assets and `go.mod` dependency.

#### Testing

- `make prepush`
- `goneat assess --categories dependencies`

---

## Archived releases

Older release notes are archived under `docs/releases/`.
