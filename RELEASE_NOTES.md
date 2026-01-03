# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.2.1] - 2026-01-03

### Pre-push license compliance via goneat assess

**Release Type**: DX Improvement

#### Overview

This release closes a workflow gap where license violations could slip through if developers only ran `make prepush` or relied on git hooks. The `dependencies` category is now included in the pre-push hook, ensuring `goneat assess --categories dependencies` runs automatically and catches forbidden licenses (GPL, LGPL, AGPL, MPL, CDDL) before code reaches the remote.

#### Highlights

- **Pre-push hook updated** – `.goneat/hooks.yaml` pre-push now includes `format,lint,security,dependencies`.
- **Automatic license checks** – No manual `make license-audit` required; dependencies category runs via assess.
- **Unified workflow** – License compliance is now part of the standard goneat assess pipeline.

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

## [0.1.29] - 2026-01-03

### Crucible v0.3.0 sync + agentic roles adoption

**Release Type**: Dependency Sync / Repo Governance

#### Overview

This release syncs gofulmen's embedded Crucible SSOT assets to `v0.3.0` and updates `go.mod` to match.

It also completes gofulmen's migration away from an identity-based agent scheme in favor of Crucible's role-based agent interface. Evergreen governance docs now reference roles (`devlead`, `devrev`, `infoarch`, `secrev`) and standardized attribution (model + interface + role), with an ADR documenting the transition.

#### Highlights

- **Crucible SSOT v0.3.0** – Synced `docs/crucible-go/`, `schemas/crucible-go/`, and `config/crucible-go/` and updated `go.mod` to `github.com/fulmenhq/crucible v0.3.0`.
- **Role catalog embedded** – Added the synced role catalog under `config/crucible-go/agentic/roles/`.
- **Roles-first governance** – Updated `AGENTS.md` and `MAINTAINERS.md` to remove identity-based agent naming and standardize on roles + attribution.
- **Secrev escalation** – Explicitly calls out `secrev` for security-sensitive changes (secrets, crypto, supply chain).
- **Tooling minimums** – Updated `Makefile` documented minimum `GONEAT_VERSION` to `v0.4.0`.

#### Testing

- `make check-all`
- `.goneat/hooks/pre-push`

---

## Archived releases

Older release notes are archived under `docs/releases/`.
