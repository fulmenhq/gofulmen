# Operations Documentation

This directory contains repository-specific operational documentation. These artifacts are **not embedded** in the library binary - they are for repository maintainers and contributors.

## Directory Structure

### `adr/` - Architecture Decision Records

Documents significant architectural decisions made during development.

**Format**: `YYYYMMDD-short-title.md`

**Template**:

```markdown
# ADR-001: Title

**Date**: YYYY-MM-DD
**Status**: Proposed | Accepted | Deprecated | Superseded

## Context

What is the issue we're seeing that is motivating this decision?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

What becomes easier or more difficult to do because of this change?
```

### `decisions/` - Repository Decisions

Decisions about repository operations, exceptions to normal processes, patterns where we override standard QA checks.

**Examples**:

- Exceptions to commit message format
- Decisions to rewrite history
- Special handling for specific files/patterns
- Temporary overrides of quality gates

### `runbooks/` - Operational Runbooks

Step-by-step procedures for common operational tasks.

**Examples**:

- Emergency hotfix process
- Release process
- Rollback procedures
- CI/CD troubleshooting

## Usage

### For Repository Maintainers

Reference these docs when:

- Making architectural decisions (add new ADR)
- Overriding standard processes (document in `decisions/`)
- Performing operational tasks (follow `runbooks/`)

### For Contributors

Review relevant ADRs and decisions before:

- Proposing significant changes
- Questioning existing patterns
- Understanding historical context

## See Also

- [INTEGRATION.md](../docs/INTEGRATION.md) - For library consumers
- [FULDX.md](../docs/FULDX.md) - FulDX development experience
- [BOOTSTRAP-STRATEGY.md](../docs/BOOTSTRAP-STRATEGY.md) - Bootstrap architecture (for repo maintainers)
