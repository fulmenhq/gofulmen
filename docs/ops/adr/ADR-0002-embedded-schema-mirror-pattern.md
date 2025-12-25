# ADR-0002: Embedded Schema Mirror Pattern (Go)

**Date**: 2025-12-24  
**Status**: Accepted

## Context

Some gofulmen modules validate user-facing configuration against JSON Schemas.

In Go, schema validation is implemented using `//go:embed` so the validator is self-contained and works in distributed binaries and libraries without requiring external schema files.

However, `//go:embed` has a key constraint:

- The embedded file must exist inside the module directory tree at build time.

Separately, Crucible is the schema SSOT and is synced into gofulmen under `schemas/crucible-go/...` via `make sync`.

This creates a tension:

- The SSOT schema lives under `schemas/crucible-go/...` (synced content).
- The app/module validator wants a schema file colocated with package code for reliable embedding.

A real bug surfaced:

- Crucible updated the AppIdentity vendor pattern to allow leading digits (e.g., `3leaps`, `37signals`).
- gofulmen synced the Crucible schema under `schemas/crucible-go/...` correctly.
- gofulmen’s embedded schema copy in `appidentity/app-identity.schema.json` drifted and still enforced the old pattern.
- Runtime validation used the embedded copy, causing legitimate identities to fail.

## Decision

Adopt an explicit **embedded schema mirror** pattern:

- Treat the Crucible-synced schema under `schemas/crucible-go/...` as the SSOT.
- Maintain a colocated schema mirror next to the Go package that embeds it (for `//go:embed`).
- The mirror is generated/copied from SSOT and MUST NOT be edited manually.

Guardrails are required:

1. Sync step

- `make sync` MUST update the embedded mirror(s) from the synced Crucible schema(s).

2. Drift test

- Add a unit test that fails if the embedded mirror diverges from the Crucible-synced copy.

## Consequences

### Positive

- Ensures Go validators always match Crucible SSOT.
- Eliminates “partial sync” failure modes where only one copy is updated.
- Establishes a repeatable pattern for any future embedded schemas.

### Negative

- Requires maintaining two files in the repo (SSOT + embed mirror), with explicit automation.
- Adds a small amount of build/sync complexity.

## Implementation

For AppIdentity:

- SSOT schema (synced):
  - `schemas/crucible-go/config/repository/app-identity/v1.0.0/app-identity.schema.json`
- Embedded mirror (Go embed):
  - `appidentity/app-identity.schema.json`
- Sync script:
  - `scripts/sync-appidentity-schema.sh`
- Drift guard test:
  - `appidentity/schema_sync_test.go`

## Notes for other languages

This ADR is Go-specific.

In other language libraries (e.g., tsfulmen, pyfulmen), schema packaging should prefer the synced SSOT schema directly as the runtime artifact where possible, rather than duplicating it.

If duplication is unavoidable for packaging reasons, adopt the same principles:

- SSOT remains Crucible.
- Mirrors are generated.
- Add drift guards.

## References

- Crucible ADR (related Go constraints): `docs/crucible-go/architecture/decisions/ADR-0009-go-module-root-relocation.md`
- AppIdentity schema SSOT: `schemas/crucible-go/config/repository/app-identity/v1.0.0/app-identity.schema.json`
