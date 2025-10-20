# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Schema** - Catalog metadata, offline metaschema validation, structured diagnostics, shared validator cache, and schema composition/diff helpers
- **Config** - Three-layer configuration loader (defaults → user → runtime) with schema validation helpers and environment override parsing

### Changed

### Deprecated

### Removed

### Fixed

- **Pathfinder** - Root boundary enforcement preventing path traversal via glob patterns
- **Pathfinder** - Hidden file filtering now checks all path segments, not just basename
- **Pathfinder** - Metadata now populated with file size and mtime (RFC3339Nano)
- **Pathfinder** - .fulmenignore support with gitignore-style pattern matching

### Security

- **Pathfinder** - Path traversal protection via ValidatePathWithinRoot preventing escapes through patterns like `../**/*.go`

## [0.1.1] - 2025-10-17

### Added

- **Foundry** package - Enterprise-grade foundation utilities from Crucible catalogs
  - RFC3339Nano timestamps with nanosecond precision for cross-language compatibility
  - UUIDv7 correlation IDs for distributed tracing (time-sortable, globally unique)
  - Pattern matching (regex, glob, literal) from Crucible catalogs (20+ curated patterns)
  - MIME type detection (content-based sniffing + extension lookup)
  - HTTP status helpers (status code grouping, validation, reason phrases)
  - Country code validation (ISO 3166-1 Alpha2, Alpha3, Numeric with case-insensitive lookups)
  - Embedded catalogs for offline operation (no network dependencies)
  - Test coverage: 85.8%
- **Guardian hooks integration** with goneat for commit/push protection
  - Pre-commit and pre-push hooks with browser-based approval workflow
  - Scoped operation policies (git.commit, git.push)
  - Expiring approval sessions (5-15 minutes)
  - .goneat/hooks.yaml manifest with make target integration
- **ADR system** synchronized with Crucible ecosystem standards
  - Two-tier ADR structure (project ADRs vs. ecosystem ADRs)
  - Ecosystem ADRs synced from Crucible (ADR-0002 through ADR-0005)
- Comprehensive test and coverage improvements
  - Overall coverage: 87.7% → 89.3% (+1.6pp)
  - Schema package tests (9.9% → 53.5%, +43.6pp)
  - Logging package tests (50.8% → 54.2%, +3.4pp)
  - Foundry catalog verification tests per Foundry Interfaces standard
  - Added schema/loader_test.go and schema/errors_test.go
  - Added testdata fixtures for schema validation
  - New logging method tests (Trace, Debug, Named, WithContext)
- Documentation enhancements
  - docs/development/bootstrap.md - Complete v0.1.0-v0.1.1 journey documentation
  - docs/development/operations.md - Version management clarifications
  - AGENTS.md - Commit message style guidance with Crucible SOP cross-links
  - .plans/active/v0.1.2/pathfinder-improvements.md - Feature brief for pathfinder

### Fixed

- **Bootstrap symlink creation** - installLink() now creates proper symlinks instead of copies
  - Replaced io.Copy with os.Symlink for type:link tools
  - bin/goneat now correctly tracks source without re-bootstrap

### Security

- **Foundry security audit remediation** - Resolved all audit findings
  - Country code: Added explicit ASCII-range validation and uppercase enforcement
  - UUIDv7: Strict version checking with timestamp monotonicity awareness
  - Numeric: IEEE 754 special-value handling (NaN, Infinity) with epsilon tolerance
  - HTTP: Configurable RoundTripper support for request interception

### Changed

- **Crucible sync** to 2025.10.2 release
  - Updated standards, SOPs, and schemas
  - Added assessment schemas (severity definitions)
  - Added version policy schema for goneat
  - Enhanced agentic attribution standard with commit message style guidance

## [0.1.0] - 2025-10-13

### Added

- Initial bootstrap of gofulmen library with 7 core packages
- **Bootstrap** package - goneat installation with download, link, and verify methods
- **Config** package - XDG Base Directory support and Fulmen configuration paths
- **Logging** package - Structured logging with RFC3339Nano timestamps and severity filtering
- **Schema** package - JSON Schema validation (draft 2020-12) with Crucible integration
- **Crucible** package - Embedded access to SSOT assets (docs, schemas, config)
- **Pathfinder** package - Safe filesystem discovery with path traversal prevention
- **ASCII** package - Terminal utilities, box drawing, Unicode analysis, and terminal overrides
- Comprehensive test coverage (85% average across all packages)
- Goneat integration for SSOT sync and version management
- Documentation, operational runbooks, and MIT licensing
- Repository safety protocols and agent attribution standards
