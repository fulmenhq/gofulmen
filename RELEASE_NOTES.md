# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

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

## Archived Releases

Older release notes are archived under `docs/releases/`.
