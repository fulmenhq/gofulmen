# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

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

## [0.1.23] - 2025-12-18

### AppIdentity embedded identity fallback (standalone artifacts) + Crucible v0.2.24

**Release Type**: Bug Fix / DX Contract Hardening

#### Overview

This release closes the remaining gap in App Identity portability: distributed artifacts can now ship a compiled-in `.fulmen/app.yaml` so basic commands work even when the binary/package runs outside a repository checkout.

#### Highlights

- **Embedded identity fallback** – Applications can call `appidentity.RegisterEmbeddedIdentityYAML()` to register a compiled-in identity used only after explicit path/env overrides and filesystem discovery fail.
- **Keeps overrides authoritative** – `FULMEN_APP_IDENTITY_PATH` and `Options.ExplicitPath` continue to short-circuit discovery and will not fall back to embedded identity.
- **Template-ready contract** – Supports the template pattern of mirroring `.fulmen/app.yaml` to an embeddable location (e.g. `internal/assets/appidentity/app.yaml`) and embedding it via `//go:embed`.
- **Crucible SSOT sync** – Updated embedded Crucible assets and `go.mod` dependency to Crucible `v0.2.24`.

#### Testing

- ✅ `go test ./appidentity -v`
- ✅ `go test ./...`

---

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

## Archived Releases

Older release notes are archived under `docs/releases/`.
