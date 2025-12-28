# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

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

## Archived releases

Older release notes are archived under `docs/releases/`.
