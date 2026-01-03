# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

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

## [0.1.28] - 2025-12-28

### Config telemetry stdout hygiene

**Release Type**: Bug Fix

#### Overview

This release fixes an output hygiene issue for CLI tooling and stdio-based MCP servers: `config.LoadLayeredConfig*` previously instantiated an internal telemetry system with telemetry enabled and no emitter, which caused JSON metrics to be written to stdout.

Config loading now uses `telemetry.GetGlobalSystem()` by default (disabled unless the application explicitly enables it), and callers can inject a telemetry system explicitly via `LayeredConfigOptions.TelemetrySystem`.

#### Highlights

- **No stdout by default** – Config loading no longer writes telemetry metrics to stdout unless telemetry is explicitly enabled.
- **Per-call override** – Added `LayeredConfigOptions.TelemetrySystem` to allow callers to route metrics to a non-stdout emitter.
- **Caller guidance** – Documented that enabling telemetry without setting `telemetry.Config.Emitter` will fall back to stdout emission.
- **Regression test** – Added `TestLoadLayeredConfig_DoesNotWriteToStdoutByDefault`.

#### Testing

- `go test ./config -run TestLoadLayeredConfig_DoesNotWriteToStdoutByDefault -v`
- `go test ./...`

---

## Archived releases

Older release notes are archived under `docs/releases/`.
