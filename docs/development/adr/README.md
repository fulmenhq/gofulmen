# Gofulmen Architecture Decision Records (ADRs)

This directory contains **local ADRs** specific to gofulmen implementation decisions. For ecosystem-wide ADRs that affect all Fulmen helper libraries, see [`docs/crucible-go/architecture/decisions/`](../../crucible-go/architecture/decisions/).

## 📋 ADR Index

| ID                                                  | Title                                                                   | Status   | Date       | Tags                                                   |
| --------------------------------------------------- | ----------------------------------------------------------------------- | -------- | ---------- | ------------------------------------------------------ |
| [ADR-0001](./ADR-0001-base-layer-error-wrapping.md) | Base Layer Packages Use Standard Go Errors, Callers Wrap with Envelopes | accepted | 2025-10-24 | error-handling, architecture, import-cycles, telemetry |

## 🎯 When to Write a Local ADR

Write a local ADR for decisions that are **Go-specific** and don't affect other language implementations:

### ✅ Write Local ADR For:

- Implementation details unique to Go (e.g., "use sync.Pool for event buffers")
- Go-specific library/dependency choices (e.g., "use santhosh-tekuri/jsonschema for validation")
- Performance optimizations specific to Go runtime (e.g., "pre-allocate slices for catalog indexes")
- Go idiom preferences (e.g., "use functional options pattern for constructors")
- Test framework choices (e.g., "use testify for assertions")
- Build/packaging decisions (e.g., "use goneat for tooling bootstrap")

### 🚀 Promote to Ecosystem ADR When:

- Decision affects API contracts between Go/Python/TypeScript
- Pattern must be consistent across all helper libraries
- Schema structure or field naming is involved
- Foundry catalog structure/completeness rules
- Other languages must implement the same behavior
- Testing fixture parity required across languages

When in doubt, discuss in `#fulmen-architecture` before creating an ecosystem ADR.

## 📚 Ecosystem ADRs (Reference)

Ecosystem ADRs are maintained in Crucible SSOT and synced to this repository at:

**[`docs/crucible-go/architecture/decisions/`](../../crucible-go/architecture/decisions/)**

These ADRs define cross-language patterns and contracts that gofulmen must implement. Current ecosystem ADRs:

- [**ADR-0001**: Two-Tier ADR System](../../crucible-go/architecture/decisions/ADR-0001-two-tier-adr-system.md) - This ADR pattern itself

## 📝 ADR Format

Use the standard template from [`docs/crucible-go/architecture/decisions/template.md`](../../crucible-go/architecture/decisions/template.md).

**Required Frontmatter:**

```yaml
---
id: "ADR-XXXX"
title: "Brief Descriptive Title"
status: "proposal" # proposal | experimental | accepted | stable | deprecated | superseded | retired
date: "YYYY-MM-DD"
deciders: ["@username"]
scope: "gofulmen" # Always "gofulmen" for local ADRs
tags: ["tag1", "tag2"]
---
```

**Standard Sections:**

1. **Context** - Problem, constraints, requirements
2. **Decision** - What we decided (with code examples if relevant)
3. **Rationale** - Why this is the right choice
4. **Alternatives Considered** - Other options evaluated
5. **Consequences** - Positive, negative, and neutral impacts
6. **Implementation** - Files modified, PRs, validation approach
7. **Related Ecosystem ADRs** - Cross-references to ecosystem decisions
8. **References** - Specs, docs, external resources

## 🔄 Promotion Path: Local → Ecosystem

If a local ADR reveals cross-language impact during implementation or review:

1. **Recognize**: Decision affects other language implementations
2. **Propose**: Create proposal in Crucible `.plans/` referencing this local ADR
3. **Coordinate**: Discuss in `#fulmen-architecture`, get buy-in from pyfulmen/tsfulmen maintainers
4. **Promote**: Create ecosystem ADR in Crucible `docs/architecture/decisions/`
5. **Update Local ADR**: Mark as "Superseded by [ADR-XXXX]" with clear link
6. **Sync**: Run `make sync` in Crucible to propagate to all language wrappers

**Example Update When Promoted:**

```markdown
**Status**: Superseded by [ADR-0012](../../crucible-go/architecture/decisions/ADR-0012-title.md)

⚠️ This local decision has been promoted to ecosystem ADR-0012 for cross-language consistency.
```

## 🤝 Contributing ADRs

1. **Draft**: Create `.md` file in this directory following the template
2. **Review**: Open PR, tag gofulmen maintainers
3. **Frontmatter**: Ensure all required fields are present and valid
4. **Cross-Reference**: Link to related ecosystem ADRs when applicable
5. **Index Update**: Add entry to the table above
6. **Merge**: Maintainer approval required

## 🔗 External Resources

- [ADR Pattern by Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [Crucible ADR Schemas](../../schemas/crucible-go/config/standards/v1.0.0/)
- [Fulmen Helper Library Standard](../../crucible-go/architecture/fulmen-helper-library-standard.md#architecture-decision-records-adrs)
- [ADR GitHub Organization](https://adr.github.io/)

---

_Local ADRs maintained by gofulmen team • Ecosystem ADRs synced from Crucible SSOT_
