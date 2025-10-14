# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [0.1.1] - 2025-10-14

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
- Comprehensive test improvements
  - Schema package tests (9.9% → 53.5%, +43.6pp)
  - Logging package tests (50.8% → 54.2%, +3.4pp)
  - Added schema/loader_test.go and schema/errors_test.go
  - Added testdata fixtures for schema validation
  - New logging method tests (Trace, Debug, Named, WithContext)
- Overall test coverage improved from 49.6% to 52.1% (+2.5pp)

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
