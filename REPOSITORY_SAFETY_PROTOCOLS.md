# Gofulmen – Repository Safety Protocols

This document outlines the safety protocols for gofulmen repository operations. For detailed operational guidelines, see the [Repository Operations SOP](docs/crucible-go/sop/repository-operations-sop.md).

## Quick Reference

- **Human Oversight Required**: All merges, tags, and publishes need @3leapsdave approval.
- **Use Make Targets**: Prefer `make` commands for consistency and safety.
- **Plan Changes**: Document work plans in `.plans/` before structural changes.
- **High-Risk Operations**: See [Repository Operations SOP](docs/crucible-go/sop/repository-operations-sop.md#high-risk-operations) for protocols.
- **Incident Response**: Follow the process in [Repository Operations SOP](docs/crucible-go/sop/repository-operations-sop.md#incident-response).

## Library-Specific Safety Guidelines

### API Stability

- **Breaking Changes**: Require major version bump and migration guide
- **Deprecation**: Mark deprecated APIs with godoc comments and maintain for at least one minor version
- **Backward Compatibility**: All changes must maintain compatibility with existing consumers

### Test Coverage

- **Minimum Coverage**: 80% for all packages
- **New Features**: Must include comprehensive tests before merge
- **Regression Tests**: Add tests for all bug fixes

### Dependency Management

- **Minimal Dependencies**: Prefer standard library when possible
- **Vetted Dependencies**: All new dependencies require @3leapsdave approval
- **Security Scanning**: Run `make license-audit` before adding dependencies

### Bootstrap Safety

- **Tool Verification**: Always verify tools after bootstrap changes
- **Checksum Validation**: Update checksums when changing tool versions
- **Local Override Pattern**: Never commit `.goneat/tools.local.yaml`

### Sync Operations

- **Crucible Assets**: Never manually edit synced files in `docs/crucible-go/`, `schemas/crucible-go/`, `config/crucible-go/`
- **Sync Before Changes**: Run `make sync` to get latest Crucible assets before making changes
- **Verify After Sync**: Run `make test` after sync to ensure compatibility

## High-Risk Operations

### Version Bumps

- **Process**: Use `make version-bump-{patch|minor|major}` only
- **Verification**: Run `make check-all` after version bump
- **Approval**: Major version bumps require @3leapsdave approval

### Release Operations

- **Pre-Release**: Run `make release-check` to validate readiness
- **Tagging**: Only @3leapsdave can create release tags
- **Publishing**: Coordinate with @3leapsdave for Go module publishing

### Structural Changes

- **Package Reorganization**: Requires architecture review with @3leapsdave
- **New Modules**: Document in `.plans/` and get approval before implementation
- **Breaking Changes**: Require migration guide and major version bump

## Incident Response

### Build Failures

1. Check `make check-all` output for specific failures
2. Fix process issues (Makefile, scripts) before code changes
3. Verify fix with full test suite
4. Document root cause in commit message

### Test Failures

1. Isolate failing test with `go test -v ./package/...`
2. Fix code or update test as appropriate
3. Ensure all tests pass before commit
4. Add regression test if bug fix

### Sync Failures

1. Check network connectivity to Crucible repository
2. Verify `.goneat/ssot-consumer.yaml` configuration
3. Try local sync with `.goneat/ssot-consumer.local.yaml`
4. Escalate to @3leapsdave if persistent

## References

- [Repository Operations SOP](docs/crucible-go/sop/repository-operations-sop.md) (canonical standard)
- `AGENTS.md`
- `MAINTAINERS.md`
- `docs/crucible-go/standards/makefile-standard.md`
- `docs/crucible-go/standards/release-checklist-standard.md`
- `docs/crucible-go/standards/coding/go.md`
- `docs/crucible-go/sop/cicd-operations.md`
- `docs/crucible-go/sop/repository-structure.md`
