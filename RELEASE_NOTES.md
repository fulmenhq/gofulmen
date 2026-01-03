# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.1.29] - 2026-01-03

### Crucible v0.3.0 sync + agentic roles adoption

**Release Type**: Dependency Sync / Repo Governance

#### Overview

This release syncs gofulmen’s embedded Crucible SSOT assets to `v0.3.0` and updates `go.mod` to match.

It also completes gofulmen’s migration away from an identity-based agent scheme in favor of Crucible’s role-based agent interface. Evergreen governance docs now reference roles (`devlead`, `devrev`, `infoarch`, `secrev`) and standardized attribution (model + interface + role), with an ADR documenting the transition.

#### Highlights

- **Crucible SSOT v0.3.0** – Synced `docs/crucible-go/`, `schemas/crucible-go/`, and `config/crucible-go/` and updated `go.mod` to `github.com/fulmenhq/crucible v0.3.0`.
- **Role catalog embedded** – Added the synced role catalog under `config/crucible-go/agentic/roles/`.
- **Roles-first governance** – Updated `AGENTS.md` and `MAINTAINERS.md` to remove identity-based agent naming and standardize on roles + attribution.
- **Secrev escalation** – Explicitly calls out `secrev` for security-sensitive changes (secrets, crypto, supply chain).
- **Tooling minimums** – Updated `Makefile` documented minimum `GONEAT_VERSION` to `v0.4.0`.

#### Testing

- ✅ `make check-all`
- ✅ `.goneat/hooks/pre-push`

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

- ✅ `go test ./config -run TestLoadLayeredConfig_DoesNotWriteToStdoutByDefault -v`
- ✅ `go test ./...`

---

## [0.1.27] - 2025-12-24

### AppIdentity schema sync fix

**Release Type**: Bug Fix

#### Overview

This release fixes a schema drift bug where gofulmen’s embedded AppIdentity schema (`appidentity/app-identity.schema.json`) did not reflect the updated Crucible SSOT schema, causing vendor validation to incorrectly reject real-world vendor IDs that begin with digits (e.g. `3leaps`, `37signals`).

#### Highlights

- **Schema synced** – `appidentity/app-identity.schema.json` now matches Crucible `schemas/crucible-go/config/repository/app-identity/v1.0.0/app-identity.schema.json`.
- **Drift guard** – Added `appidentity/schema_sync_test.go` to fail CI if the embedded schema diverges.
- **Sync hook** – `make sync` now refreshes the embedded schema via `scripts/sync-appidentity-schema.sh`.

#### Testing

- ✅ `go test ./appidentity -run TestEmbeddedAppIdentitySchemaMatchesCrucible -v`

---

## Archived releases

Older release notes are archived under `docs/releases/`.
