# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.3.0] - 2026-01-06

### Crucible v0.4.1 sync + similarity module promotion

**Release Type**: Minor Release (Structural Change)

#### Overview

This release syncs Crucible v0.4.1 which promotes the similarity module from a foundry sub-package to a standalone module. The module registry schema (v1.1.0) now includes `weight` and `default_inclusion` fields to support feature-gating in downstream consumers.

#### Highlights

- **Crucible v0.4.1** – Updated embedded Crucible assets and `go.mod` dependency.
- **Similarity promotion** – Fixtures and schemas moved from `library/foundry/` to `library/similarity/`.
- **Module schema v1.1.0** – New fields: `weight` (light/heavy), `default_inclusion` (bool), per-language `notes`.
- **Signals catalog** – Now contains 9 signals (added new signal).
- **Test path updates** – Fixture and schema paths updated to new locations.

#### Breaking Changes

- Synced content paths changed: `config/crucible-go/library/foundry/similarity-fixtures.yaml` → `config/crucible-go/library/similarity/fixtures.yaml`
- Schema paths changed: `schemas/crucible-go/library/foundry/v2.0.0/` → `schemas/crucible-go/library/similarity/v2.0.0/`

Consumers referencing synced files directly will need to update paths. The Go API (`foundry/similarity` package) is unchanged.

#### Testing

- `make check-all`
- `make release-provenance-check`

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

## [0.2.0] - 2026-01-03

### Native similarity algorithms for MIT-compatible SBOM

**Release Type**: Minor Release (License Compliance)

#### Overview

This release replaces GPL-2.0 licensed external dependencies in the similarity package with native Go implementations. Downstream consumers can now use gofulmen with fully MIT/Apache-2.0 compatible dependency trees.

#### Highlights

- **Native Jaro-Winkler** – New `jaro_winkler.go` implementing standard Jaro-Winkler similarity.
- **Native Damerau-Levenshtein** – Unrestricted variant now uses native implementation with proper transposition handling.
- **Clean SBOM** – `make license-audit` passes with no GPL/LGPL/AGPL/MPL/CDDL licenses.
- **Quality Gate** – Added `license-audit` to `make check-all` to prevent future license regressions.
- **Golden Tests** – 58 test cases verifying algorithm correctness.

#### Testing

- `make check-all` (now includes license-audit)
- `make license-audit`
- All Crucible fixture tests pass
- All golden tests pass

---

## Archived releases

Older release notes are archived under `docs/releases/`.
