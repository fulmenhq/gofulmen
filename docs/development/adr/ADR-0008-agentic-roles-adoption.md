# ADR-0008: Adopt role-based agentic interface (Crucible v0.3.0)

Date: 2026-01-03

## Status

Accepted

## Context

Gofulmen previously experimented with an identity-based scheme for AI agents (a typed name/handle intended for human coordination). In practice, both humans and agents perceived these identities as anthropomorphizing. That was not the intent, and it created confusion about accountability and authorship.

Crucible `v0.3.0` introduces a standardized agentic interface built around:

- Schema-validated role prompts (`config/crucible-go/agentic/roles/*.yaml`)
- Standardized attribution (model + interface + role)

This provides a clear, composable framework for supervised and autonomous workflows without requiring “personas”.

## Decision

Effective with gofulmen syncing Crucible `v0.3.0` (and updating `go.mod` accordingly), gofulmen adopts **role-based prompts and attribution**.

- Evergreen operational docs should reference roles (e.g., `devlead`, `devrev`, `infoarch`) rather than named identities.
- Commit attribution follows `docs/crucible-go/standards/agentic-attribution.md` and requires a `Role:` trailer.
- We do not rewrite historical/release-specific artifacts (release notes, dated memos, etc.). Those remain as historical record.

## Legacy identity → role mapping

This mapping is provided for continuity while reading older discussions/logs.

| Legacy identity / handle | Primary role going forward |
| ------------------------ | -------------------------- |
| Foundation Forge         | `devlead`                  |
| @foundation-forge        | `devlead`                  |
| EA Steward               | `entarch`                  |
| @fulmen-ea-steward       | `entarch`                  |

## Consequences

### Positive

- Removes anthropomorphism pressure and clarifies accountability.
- Aligns gofulmen with Crucible’s ecosystem-wide standards.
- Enables autonomous agent coordination via explicit roles.

### Trade-offs

- Humans lose a simple “typed name” for routing requests (now use role + model/interface).
- Cross-repo discussions referencing older identities may need the mapping table.

## Notes

- If future coordination channels still require a typed label, prefer role-first labels like `agent-devlead`, not persona names.
- If repository-specific overrides are needed, follow the Crucible guidance: use a local override directory (gitignored) rather than editing synced role YAML.
