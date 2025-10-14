---
title: "Fulmen Helper Library Standard"
description: "Standard structure and capabilities for gofulmen, tsfulmen, and future language helpers"
author: "Schema Cartographer"
date: "2025-10-02"
last_updated: "2025-10-10"
status: "draft"
tags: ["architecture", "helper-library", "multi-language", "local-development"]
---

# Fulmen Helper Library Standard

This document formalizes the expectations outlined in the gofulmen proposal so that every `*fulmen`
language foundation delivers the same core capabilities. The canonical list of languages and statuses
is maintained in the [Library Ecosystem taxonomy](library-ecosystem.md#language-foundations-taxonomy);
consult that table before proposing new foundations or changing lifecycle state.

## Scope

Applies to language-specific Fulmen helper libraries (gofulmen, tsfulmen, pyfulmen, csfulmen, rufulmen, etc.). Excludes SSOT repos (Crucible, Cosmography) and application/tool repos (Fulward, goneat).

## Mandatory Capabilities

1. **Goneat Bootstrap Pattern**
   - Follow the [Goneat Bootstrap Guide](../guides/bootstrap-goneat.md) so contributors and CI can install the CLI via package manager or `go install github.com/fulmenhq/goneat/cmd/goneat@latest`.
   - Provide a `make bootstrap` target (or equivalent script) that verifies goneat is available and documents the fallback steps in `docs/`.
   - Keep `.goneat/tools.yaml` and `.goneat/tools.local.yaml.example` templates for teams that still need download-based bootstraps; ensure `.goneat/tools.local.yaml` remains gitignored.
   - Once goneat is installed, use `goneat ssot sync` with a repository-specific `.goneat/ssot-consumer.yaml` (see below) to pull Crucible assets. Avoid bundling legacy FulDX binaries.

2. **SSOT Synchronization**
   - Include `.goneat/ssot-consumer.yaml` describing the Crucible assets your library consumes. Recommended shape:
     ```yaml
     version: "2025.10.2"
     sync_path_base: lang/<language> # e.g., lang/typescript, lang/python
     sources:
       - id: crucible
         repo: https://github.com/fulmenhq/crucible.git
         ref: main
         include:
           - docs/**/*.md
           - schemas/**/*
           - config/**/*.yaml
           - config/sync/sync-keys.yaml
         notes: "Use sync keys defined in config/sync/sync-keys.yaml"
     ```
     `sync_path_base` MUST match the directory under `lang/` (for example `lang/typescript`) so that goneat writes the synced artifacts to the correct tree.
   - Create `.goneat/ssot-consumer.local.yaml` for local development (gitignored) with overrides such as `local_path: ../crucible`.
   - Add to `.gitignore`:
     ```
     .goneat/*.local.yaml
     ```
   - Commit synced assets—for example `docs/crucible-<lang>`, `schemas/crucible-<lang>`, `config/crucible-<lang>`, and `.crucible/metadata`—so consumers can work offline.
   - Provide a `make sync` target (or equivalent) that runs `goneat ssot sync --manifest .goneat/ssot-consumer.yaml`.
   - Use glob patterns such as `schemas/**/*` to capture both `.json` and `.yaml` assets.
   - Refer to the [SSOT Sync Standard](../standards/library/modules/ssot-sync.md) for command surface and testing guidance.

3. **Crucible Shim**
   - Provide idiomatic access to Crucible assets (docs, schemas, config defaults).
   - Re-export version constants so consumers can log/report underlying Crucible snapshot.
   - Discover available categories (`ListAvailableDocs()`, `ListAvailableSchemas()`) via embedded metadata or generated index.
   - Refer to the [Crucible Shim Standard](../standards/library/modules/crucible-shim.md).

4. **Config Path API**
   - Implement `GetAppConfigDir`, `GetAppDataDir`, `GetAppCacheDir`, `GetAppConfigPaths`, and `GetXDGBaseDirs` (naming per language).
   - Expose Fulmen-specific helpers (`GetFulmenConfigDir`, etc.) aligned with [Fulmen Config Path Standard](../standards/config/fulmen-config-paths.md).
   - Respect platform defaults (Linux/macOS/Windows) and environment overrides.
   - Refer to the [Config Path API Standard](../standards/library/modules/config-path-api.md).

5. **Three-Layer Config Loading**
   - Layer 1: Embed Crucible defaults from `config/{category}/vX.Y.Z/*-defaults.yaml`.
   - Layer 2: Merge user overrides from `GetFulmenConfigDir()`.
   - Layer 3: Allow application-provided config (BYOC) with explicit API hooks.
   - Refer to the [Three-Layer Configuration Standard](../standards/library/modules/three-layer-config.md).

6. **Schema Validation Utilities**
   - Provide helpers to load, parse, and validate schemas shipped in Crucible.
   - Optional but recommended: integrate with language-native validation libraries.
   - Refer to the [Schema Validation Helper Standard](../standards/library/modules/schema-validation.md).

7. **Observability Integration**
   - Consume logging schemas/defaults from `config/observability/logging/`.
   - Map shared severity enum and throttling settings to language-specific logging implementation.
   - Refer to the [Fulmen Logging Standard](../standards/observability/logging.md).

## Optional (Recommended) Capabilities

- Pathfinder & ASCII helpers (for languages that can support them).
- Cosmography shims once that SSOT expands.
- Registry API clients if SSOT repos expose HTTP endpoints in the future.

Module requirement levels, coverage targets, and language overrides are tracked in
`config/library/v1.0.0/module-manifest.yaml` (validated by
`schemas/library/module-manifest/v1.0.0/module-manifest.schema.json`).

## Directory Structure

### Required Directories

```
<foundation-repo>/
├── .goneat/
│   ├── tools.yaml                    # Production tooling manifest (goneat ensured on PATH)
│   ├── tools.local.yaml.example      # Local override template (committed)
│   ├── ssot-consumer.yaml            # SSOT sync configuration (committed)
│   └── ssot-consumer.local.yaml.example  # Local override template (committed)
├── .crucible/
│   └── metadata/                     # Sync metadata produced by goneat ssot sync
├── docs/
│   └── crucible-<lang>/              # Synced docs (committed, regenerated via sync)
├── schemas/
│   └── crucible-<lang>/              # Synced schemas (committed, regenerated via sync)
├── config/
│   └── crucible-<lang>/              # Synced config defaults (committed, regenerated)
└── bin/                              # Optional vendored tools (gitignored if used)
    └── goneat
```

### Namespace Guidance

- SSOT-specific helpers should live under namespaces such as `crucible/logging`, `crucible/terminal`, `cosmography/maps`.
- Language-only utilities (e.g., Go reflection helpers) may live under `foundation/` namespaces.
- Provide package-level READMEs describing available modules.

## Bootstrap Strategy

### The Problem: Bootstrap the Bootstrap

Foundation libraries face a unique challenge:

- They're dependencies for sophisticated tools (goneat, fulward)
- They need DX tooling themselves (goneat for validation, SSOT sync, and release automation)
- They cannot create circular dependencies
- They must work in CI/CD without manual installation steps

### The Solution: Minimal Goneat Bootstrap + Synced Assets

1. **Commit synced assets** - Docs, schemas, configs from Crucible are committed and regenerated via `make sync`.
2. **Single bootstrap entry** - `.goneat/tools.yaml` ensures goneat is installed (via package manager or download).
3. **Local override pattern** - `.goneat/tools.local.yaml` (gitignored) for development iteration.
4. **Automated sync** - `goneat ssot sync --manifest .goneat/ssot-consumer.yaml` keeps mirror directories current.

### Workflows

**CI/CD (Production):**

```bash
git clone <foundation-repo>  # Includes synced assets
make bootstrap               # Installs/validates goneat
make sync                    # Optional: update from Crucible via goneat ssot sync
make test
```

**Local Development (Iteration):**

```bash
# Copy local override template
cp .goneat/tools.local.yaml.example .goneat/tools.local.yaml

# Edit to point to local goneat build if necessary
# source: /path/to/goneat/dist/goneat

# Bootstrap uses local override
make bootstrap
```

### Safety: Preventing Local Path Leaks

- Add `.goneat/tools.local.yaml` to `.gitignore`
- Implement precommit hook to validate no local paths in `tools.yaml`
- Provide `tools.local.yaml.example` as template

## Testing Expectations

- Each language foundation owns its unit/integration tests.
- Crucible supplies schemas and config defaults but does not ship tests for the shims.
- Tests should cover:
  - Config path resolution (including legacy fallbacks).
  - Embedding/parsing of Crucible defaults.
  - Schema validation wrappers.
  - Logging severity/middleware mapping.
  - Foundry catalog helpers normalizing alpha-2/alpha-3/numeric country codes with case-insensitive lookups backed by precomputed secondary indexes.
  - Bootstrap script functionality (both local and production paths).

## Documentation Requirements

- README with installation, quick start, links to Crucible standards, **and a prominent link to the overview document described below.**
- Goneat bootstrap notes (or reference this standard / `docs/guides/bootstrap-goneat.md`).
- Bootstrap strategy documentation (e.g., `docs/BOOTSTRAP-STRATEGY.md`).
- API reference comments/docstrings per language norms.
- Notes on dependency flow (SSOT → foundation → consumer) to prevent circular imports.
- Architecture overview stored at `docs/<lang>fulmen_overview.md` (for example, `docs/pyfulmen_overview.md`). Use the following template to ensure consistency:

  ```markdown
  # <Language> Fulmen Overview

  ## Purpose & Scope

  - Brief summary of the foundation library and supported environments.

  ## Module Catalog

  | Module          | Tier | Summary                                  | Spec Link                                                       |
  | --------------- | ---- | ---------------------------------------- | --------------------------------------------------------------- |
  | config-path-api | Core | Platform-aware config/data/cache helpers | [config-path-api](standards/library/modules/config-path-api.md) |
  | ...             | ...  | ...                                      | ...                                                             |

  ## Observability & Logging Integration

  - Default logging profile, policy search order, notable middleware/sinks enabled by default.

  ## Dependency Map

  | Artifact          | Description                | Source             |
  | ----------------- | -------------------------- | ------------------ |
  | Crucible schemas  | Synced SSOT assets         | `goneat ssot sync` |
  | Published package | e.g., `@fulmenhq/tsfulmen` | npm (link)         |
  | Optional extras   | e.g., telemetry plugins    | link               |

  ## Roadmap & Known Gaps

  - Bullet list of planned enhancements or known limitations.
  ```

- Development operations documentation located under `docs/development/`. Every library MUST provide:

  ```markdown
  docs/
  ├── <lang>fulmen*overview.md
  ├── development/
  │ ├── README.md # Index linking to operations/testing/bootstrap docs
  │ └── operations.md # Required operations guide
  └── crucible*<lang>/ # Synced Crucible docs (as today)
  ```

  `docs/development/operations.md` MUST cover:
  1. Development workflow (primary `make` targets, lint/typecheck commands)
  2. Release process (versioning strategy, changelog expectations, required checks)
  3. Testing strategy (coverage targets from the module manifest, tooling)
  4. Tooling reference (bootstrap, goneat usage, language-specific helpers)
  5. Community & support notes (contribution guidelines, communication channels)
  6. Security expectations (dependency scanning, vulnerability reporting)

  Libraries MAY add additional documents (`testing.md`, `release-process.md`, `bootstrap.md`, etc.) referenced from `docs/development/README.md`.

  ```

  ```

## Version Alignment

- Foundations MUST pin the Crucible version they embed and expose it publicly.
- Consumers depend on the foundation version; no direct Crucible import required.
- When Crucible publishes new assets, foundations should sync and bump versions in tandem.
- Use the repo-provided CalVer automation (for example `bun run scripts/update-version.ts --bump <patch|minor|major>`) to keep artifacts aligned.
- Declare repository maturity via `LIFECYCLE_PHASE`, following the [Repository Lifecycle Standard](../standards/repository-lifecycle.md). Coverage gates and CI policies MUST respect the thresholds defined for each phase.

## Makefile Targets (Required)

In addition to the standard targets, support overriding bootstrap with `FORCE=1` or provide a `bootstrap-force` alias so tool reinstall can be forced when iterating on local builds.

All foundation libraries MUST provide these targets:

```makefile
bootstrap:      # Install or verify goneat availability
sync:           # Sync assets from Crucible SSOT via goneat
version-bump:   # Bump version (requires TYPE/BUMP parameter)
test:           # Run all tests
fmt:            # Format code
lint:           # Lint/style checks
```

## Common Pitfalls & Solutions

### Pitfall 1: Syncing Only JSON Schemas

- **Issue**: Using `schemas/**/*.json` misses `.yaml` schemas.
- **Solution**: Use `schemas/**/*` or `schemas/**/*.{json,yaml}`.

### Pitfall 2: Outdated Goneat Version

- **Issue**: Older goneat binaries may lack `ssot sync` enhancements.
- **Solution**: Re-run `make bootstrap` or reinstall goneat (`go install ...@latest`, `brew install goneat`, etc.) before syncing.

### Pitfall 3: Local Overrides Committed

- **Issue**: `.goneat/tools.local.yaml` accidentally committed.
- **Solution**: Provide `.goneat/tools.local.yaml.example`, add `.goneat/tools.local.yaml` to `.gitignore`, enforce pre-commit checks.

### Pitfall 4: Circular Bootstrap Dependencies

- **Issue**: Using project-internal tools before goneat is installed can create cycles.
- **Solution**: Keep bootstrap focused on ensuring goneat is available; delegate other tooling installs to repo-specific scripts executed after goneat is ready.

### Pitfall 5: Missing Maintainer Docs

- **Issue**: Ops/ADR info lost in commits.
- **Solution**: Maintain `ops/` directory with ADRs, runbooks, bootstrap strategy.

## References

- [Fulmen Library Ecosystem](library-ecosystem.md)
- [Fulmen Config Path Standard](../standards/config/fulmen-config-paths.md)
- [Config Defaults README](../../config/README.md)
- [Goneat Repository](https://github.com/fulmenhq/goneat)
- `.plans/crucible/fulmen-helper-library-specification.md` (original proposal)

## Implementation Examples

- **gofulmen** - Reference implementation with Go-based bootstrap
- **tsfulmen** - TypeScript implementation (in development)

## Changelog

- **2025-10-10** - Migrated bootstrap + sync guidance from FulDX to goneat
- **2025-10-07** - Added FulDX bootstrap pattern, SSOT sync requirements, directory structure
- **2025-10-03** - Initial draft
