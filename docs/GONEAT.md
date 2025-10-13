# Goneat Integration Guide

This guide explains how goneat is integrated into gofulmen for development tooling and workflows.

## Overview

Gofulmen uses the [goneat](https://github.com/fulmenhq/goneat) CLI for standardized development operations:

- **Bootstrap**: Install and manage development tools
- **Sync**: Synchronize assets from Crucible SSOT
- **Version**: Manage version bumps and releases
- **Format**: Code formatting and linting
- **Assess**: Code quality assessment

**Important**: Gofulmen is a dependency of goneat, creating a circular dependency. Therefore, goneat is installed as a binary download (not via go install) to avoid the circular dependency.

## Bootstrap Pattern

### Tool Manifest

The `.goneat/tools.yaml` file defines the tools required for development, including goneat itself as a binary download:

```yaml
version: v0.3.0
binDir: ./bin
tools:
  - id: goneat
    description: Fulmen schema validation and automation CLI
    required: true
    install:
      type: download
      url: https://github.com/fulmenhq/goneat/releases/download/v0.3.0/goneat-{{os}}-{{arch}}
      binName: goneat
      destination: ./bin
      checksum:
        darwin-arm64: "0" # TODO: Replace with actual checksums
        darwin-amd64: "0" # TODO: Replace with actual checksums
        linux-amd64: "0" # TODO: Replace with actual checksums
        linux-arm64: "0" # TODO: Replace with actual checksums
```

### Installing Goneat

Goneat is installed via bootstrap as a binary download to avoid circular dependency (since gofulmen is a dependency of goneat, we can't use `go install`).

For local development, use `.goneat/tools.local.yaml` to link to your local build:

```yaml
version: v0.3.0-dev
binDir: ./bin
tools:
  - id: goneat
    description: Fulmen schema validation and automation CLI (local dev)
    required: true
    install:
      type: link
      source: ../goneat/dist/goneat
      binName: goneat
      destination: ./bin
```

### Local Development Overrides

For local development, the bootstrap automatically uses `.goneat/tools.local.yaml` if it exists (this file is gitignored). This allows you to:

1. Link to your local goneat build during development
2. Override any other tools with local versions

The `.goneat/tools.local.yaml.example` file provides a template.

## Makefile Integration

The Makefile provides convenient targets:

```bash
# Install tools (including goneat as binary download)
make bootstrap

# Force reinstall
make bootstrap-force

# Verify tools are available
make tools

# Sync assets from Crucible
make sync

# Bump version
make version-bump TYPE=patch
```

## Development Workflow

### Initial Setup

1. Clone the repository
2. Run `make bootstrap` to install tools (including goneat as binary)
3. Run `make sync` to get latest Crucible assets

### Daily Development

1. Make changes
2. Run `make fmt` to format code
3. Run `make test` to run tests
4. Run `make check-all` for full quality checks

### Release Process

1. Update VERSION file if needed
2. Run `make version-bump TYPE=patch|minor|major`
3. Update CHANGELOG.md
4. Create release PR

## Integration with Crucible

Goneat handles synchronization with Crucible SSOT:

- Documentation: `docs/crucible-go/`
- Schemas: `schemas/crucible-go/`
- Config defaults: `config/crucible-go/`

These assets are committed to version control for offline availability.

## Troubleshooting

### Bootstrap Issues

If bootstrap fails:

1. Check if the goneat release exists
2. Verify checksums in `.goneat/tools.yaml`
3. Check network connectivity
4. Try `make bootstrap-force` to reinstall

### Sync Issues

If sync fails:

1. Ensure goneat is installed: `make bootstrap`
2. Check Crucible repository accessibility
3. Verify `.fuldx/sync-consumer.yaml` configuration

### Version Issues

If version bump fails:

1. Ensure working directory is clean
2. Check if VERSION file exists
3. Verify goneat is installed

## Migration from FulDX

This project has migrated from FulDX to goneat:

- `.crucible/` → `.goneat/`
- `fuldx` commands → `goneat` commands
- Updated Makefile targets
- New bootstrap pattern

The old FulDX files have been removed and replaced with the goneat equivalent.
