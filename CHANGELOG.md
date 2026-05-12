# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Convention**: Keep only latest 10 releases here (reverse-chronological) to prevent file bloat. Older releases are archived in `docs/releases/`.

## [Unreleased]

## [0.3.5] - 2026-05-12

### Fixed

- **appidentity**: Embedded application identity now takes precedence over CWD and executable-directory discovery, preventing shipped binaries from accidentally adopting a foreign checkout's `.fulmen/app.yaml`.
- **security**: Cleared release-blocking high findings by adding checked ZIP size conversions, permission-mode normalization, validated path/binary resolution, and CRC byte serialization helpers.

### Changed

- **go.uber.org/zap** v1.27.1 → v1.28.0 – Minor.
- **github.com/mattn/go-runewidth** v0.0.20 → v0.0.23 – Patch.
- **golang.org/x/mod** v0.33.0 → v0.36.0 – Minor.
- **golang.org/x/text** v0.34.0 → v0.37.0 – Minor.
- **golang.org/x/time** v0.14.0 → v0.15.0 – Minor.
- **CI/tooling**: Updated the Fulmen toolbox runner, aligned YAML lint configuration with goneat formatting rules, and kept the pull-request external installation test active.
- **Git hooks**: Regenerated and installed goneat hooks without guardian browser interception; repository protection is moving to merge-policy controls instead of local commit/push approval prompts.
- **Makefile**: Moved longer recipes into helper scripts to reduce lint churn while preserving public Make targets.

### Documentation

- **docs/GONEAT.md**: Documented local/container format and lint alignment workflow.
- **README.md**: Refreshed release-sensitive version examples and release validation guidance.
- **Release planning**: Deferred config utility primitive exports to v0.3.6 after confirming they no longer gate v0.3.5.

## [0.3.4] - 2026-02-20

### Added

- **crucible**: Typed role catalog API via shim re-exports from Crucible v0.4.12.
  - `LoadRole(slug)`, `ListRoleSlugs()`, `LoadRoleCatalog()` wrapper functions.
  - Type aliases: `RolePrompt`, `RoleMindset`, `RoleEscalation`, `RoleExample`, `RoleRequiredReading`, `RoleRequiredReadingFile`, `AgenticConfig`.
  - Raw YAML access via `ConfigRegistry.Agentic().Role(slug)`.
  - 12 tests exercising the shim path (invariant-based, not exact-count).

### Changed

- **Crucible v0.4.12** – Synced SSOT assets including 3 new agentic roles (`cxotech`, `deliverylead`, `infraeng`) and `domains` field on all existing roles.
- **go.uber.org/zap** v1.27.0 → v1.27.1 – Patch.
- **github.com/mattn/go-runewidth** v0.0.19 → v0.0.20 – Patch.
- **golang.org/x/mod** v0.32.0 → v0.33.0 – Minor.
- **golang.org/x/text** v0.33.0 → v0.34.0 – Minor.
- **github.com/bmatcuk/doublestar/v4** v4.9.1 → v4.10.0 – Minor.
- **github.com/zeebo/xxh3** v1.0.2 → v1.1.0 – Minor (pulls cpuid v2.2.10, x/sys v0.30.0).
- **github.com/stretchr/testify** v1.8.1 → v1.11.1 – Test-only minor.

### Documentation

- **crucible/README.md** – Expanded role catalog section with per-task snippets, RolePrompt field reference table, and "check if role exists" pattern. Fixed stale version numbers.
- **AGENTS.md** – Added explicit copy-paste commit trailer template to prevent agents from omitting `Role:` and `Committer-of-Record:` lines.

## [0.3.3] - 2026-02-04

### Added

- **signals**: `HTTPConfig.AllowedSignals` to restrict which signals can be requested via `HTTPHandler`.
- **signals**: `HTTPConfig.AllowClientGracePeriod` to control whether the handler honors `grace_period_seconds`.
- **config**: `EnvVarSpecWithAliases` plus `LoadEnvOverridesWithReport` / `LoadEnvOverridesWithEnvelopeAndReport` for alias precedence + conflict diagnostics (sensitive values masked by default).
- **logging**: `IsSensitiveKey` helper for envinfo/doctor-style masking.

### Changed

- **signals**: `HTTPHandler` now ignores client `grace_period_seconds` unless `AllowClientGracePeriod` is enabled.
- **signals + telemetry/exporters**: bearer token auth now uses normalized parsing and constant-time compare.
- **Crucible v0.4.10** – Synced SSOT assets and updated `go.mod` dependency.

### Fixed

- **foundry + crucible**: version reporting now reflects the actual gofulmen module version in downstream binaries.

## [0.3.2] - 2026-01-28

### Added

- **Fulencode Module** – New `fulencode/` package providing canonical encoding/decoding with security protections.
  - **Encode**: Base64/Base64URL/Base64-raw, Base32/Base32hex, Hex + UTF-8/UTF-16LE/UTF-16BE/ISO-8859-1/CP1252/ASCII
  - **Decode**: All formats with expansion ratio limits, max size protection, checksum support
  - **Detect**: BOM detection, UTF-16 null-pattern heuristic, UTF-8 validation, confidence scoring
  - **Normalize**: NFC/NFD/NFKC/NFKD + `text_safe` profile (control char rejection, bidi/zero-width filtering, combining mark limits)
  - **BOM Helpers**: `DetectBOM`, `RemoveBOM`, `AddBOM` for byte order mark handling
  - Uses Crucible SSOT enums via `github.com/fulmenhq/crucible/fulencode`
  - Fixture-backed tests using Crucible v0.4.9 test vectors
- **JSON Schema Meta-Draft Support** – End-to-end validation for Draft-04 through Draft-2020-12.
  - Updated `schema/validator.go` loader to resolve json-schema.org metaschemas for all drafts
  - Added `schema/meta_drafts_test.go` regression tests using Crucible meta fixtures

### Changed

- **Go 1.25.5** – Updated toolchain from 1.25.1 for vulnerability remediation in standard library.
- **golang.org/x/mod v0.32.0** – Updated from v0.30.0.
- **golang.org/x/text v0.33.0** – Updated from v0.30.0.
- **Crucible v0.4.9** – Synced SSOT assets including:
  - New QA agentic role (`config/crucible-go/agentic/roles/qa.yaml`)
  - Fulencode library schemas, fixtures, and text-safe standard
  - JSON Schema meta-schema drafts (04, 06, 2019-09) with fixtures
  - Classifiers framework with dimension schemas and standards
  - Foundation schemas (error-response, lifecycle-phases, release-phase)
  - Reorganized upstream 3leaps schemas under `crucible/` namespace

## [0.3.1] - 2026-01-07

### Changed

- **Crucible v0.4.3 Sync** – Updated embedded Crucible assets and `go.mod` dependency to `github.com/fulmenhq/crucible v0.4.3`.

## [0.3.0] - 2026-01-07

### Changed

- **Crucible v0.4.2 Sync** – Updated embedded Crucible assets and `go.mod` dependency to `github.com/fulmenhq/crucible v0.4.2`.
- **Canonical URI Resolution** – All schema `$id` values now use module-qualified URIs (`schemas.fulmenhq.dev/crucible/...`).
- **Schema Resolver Fix** – `schema/validator.go` now explicitly handles only `crucible/` module schemas; other modules return clear error.
- **Similarity Module Promotion** – Similarity fixtures and schemas moved from `library/foundry/` to `library/similarity/` as a standalone module.
- **Module Schema v1.1.0** – Synced new module registry schema with `weight`, `default_inclusion`, and per-language `notes` fields.
- **Fixture Standard** – New standard for test infrastructure repositories (`docs/crucible-go/architecture/fulmen-fixture-standard.md`).

### Removed

- **Enact Schemas** – 11 schemas moved to enacthq organization.
- **Goneat Schemas** – 6 schemas moved to goneat repository.

### Breaking Changes

- Schema `$id` URIs now include `crucible/` module prefix. Resolvers using `$id` lookup must handle the new pattern.
- Synced content paths changed for similarity:
  - `config/crucible-go/library/foundry/similarity-fixtures.yaml` → `config/crucible-go/library/similarity/fixtures.yaml`
  - `schemas/crucible-go/library/foundry/v2.0.0/similarity.schema.json` → `schemas/crucible-go/library/similarity/v2.0.0/similarity.schema.json`
- Consumers referencing synced files directly will need to update paths. The Go API (`foundry/similarity` package) is unchanged.

## [0.2.1] - 2026-01-03

### Changed

- **Pre-push License Compliance** – Added `dependencies` category to `.goneat/hooks.yaml` pre-push hook, ensuring license compliance checks run automatically via `goneat assess --categories dependencies` on every push.
- **Crucible v0.3.1 Sync** – Updated embedded Crucible assets and `go.mod` dependency to `github.com/fulmenhq/crucible v0.3.1`.

## [0.2.0] - 2026-01-03

### Changed

- **Native Similarity Algorithms** – Jaro-Winkler and unrestricted Damerau-Levenshtein now use native Go implementations, replacing external dependency for MIT-compatible SBOM.
- **License Compliance** – Removed GPL-2.0 licensed dependency, enabling clean license audits for downstream consumers.
- **Quality Gate** – Added `license-audit` to `make check-all` target to prevent future license regressions.

### Added

- `foundry/similarity/jaro_winkler.go` – Native Jaro-Winkler implementation.
- `foundry/similarity/golden_matchr_test.go` – Golden tests for algorithm verification (58 test cases).

### Removed

- `foundry/similarity/levenshtein_benchmark_comparison_test.go` – External library comparison no longer needed.

## [0.1.29] - 2026-01-03

### Added

- **Agentic Role Catalog** – Synced the Crucible role catalog (`config/crucible-go/agentic/roles/`) including `devlead`, `devrev`, `infoarch`, `secrev`, and other ecosystem roles.
- **Role Adoption ADR** – Added `docs/development/adr/ADR-0008-agentic-roles-adoption.md` documenting the gofulmen transition to role-based agent operation.

### Changed

- **Crucible v0.3.0 Sync** – Updated embedded Crucible assets and `go.mod` dependency to `github.com/fulmenhq/crucible v0.3.0`.
- **Agentic Guidance** – Updated evergreen governance docs (`AGENTS.md`, `MAINTAINERS.md`) to remove identity-based agent naming and standardize on roles + attribution.
- **Tooling Minimums** – Updated `Makefile` documented minimum `GONEAT_VERSION` to `v0.4.0`.

## [0.1.28] - 2025-12-28

### Fixed

- **Config Telemetry Stdout Leak** – `config.LoadLayeredConfig*` no longer instantiates an enabled-by-default telemetry system that emits JSON metrics to stdout.
  - Uses `telemetry.GetGlobalSystem()` (disabled-by-default) unless the caller injects `LayeredConfigOptions.TelemetrySystem`.
  - Adds a regression test to ensure config load produces no stdout output by default.

## Archived releases

Older release notes (v0.1.27 and earlier) are archived under `docs/releases/`.
