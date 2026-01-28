# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.3.2] - 2026-01-28

### Fulencode module + JSON Schema meta-draft support

**Release Type**: Minor Release (New Feature)

#### Overview

This release introduces the **fulencode** module—a canonical encoding/decoding library with built-in security protections—and expands JSON Schema validation to support all drafts from Draft-04 through Draft-2020-12. Go toolchain updated to 1.25.5 for vulnerability remediation.

#### Highlights

- **Fulencode Module** – New `fulencode/` package for encoding operations across the Fulmen ecosystem.
  - 12 encoding formats: Base64 variants, Base32 variants, Hex, UTF-8/16, ISO-8859-1, CP1252, ASCII
  - Security by default: Expansion ratio limits, max size protection, encoding bomb detection
  - Detection with confidence scoring: BOM, UTF-16 null patterns, UTF-8 validation
  - Normalization profiles: NFC/NFD/NFKC/NFKD + security-focused `text_safe` profile
  - SSOT integration: Uses Crucible-generated enums for cross-language consistency
- **JSON Schema Meta-Draft Support** – Validate schemas using any draft (04, 06, 07, 2019-09, 2020-12).
- **Go 1.25.5** – Toolchain update addresses stdlib vulnerabilities in Go 1.25.1.
- **golang.org/x providers** – Updated x/mod (v0.32.0) and x/text (v0.33.0).
- **Crucible v0.4.9** – New QA role, fulencode schemas/fixtures, classifiers framework, foundation schemas.

#### For Library Consumers

The new `fulencode/` package provides consistent encoding operations:

```go
import "github.com/fulmenhq/gofulmen/fulencode"

// Encode bytes to base64
result, _ := fulencode.Encode(data, fulencode.BASE64, nil)

// Decode with security limits
decoded, _ := fulencode.DecodeString(encoded, fulencode.BASE64, &fulencode.DecodeOptions{
    MaxDecodedSize:    100 * 1024 * 1024,  // 100MB limit
    MaxExpansionRatio: 10.0,                // Bomb protection
})

// Detect encoding with confidence
detection, _ := fulencode.Detect(unknownBytes, nil)
fmt.Printf("Encoding: %s (%.0f%% confidence)\n", *detection.Encoding, detection.Confidence*100)

// Normalize text safely (reject control chars, bidi, zero-width)
normalized, _ := fulencode.Normalize(userInput, fulencode.TEXT_SAFE, nil)
```

#### Testing

- `make check-all`
- `make test`
- All fixture tests pass against Crucible v0.4.9 test vectors

---

## [0.3.1] - 2026-01-07

### Crucible v0.4.3 sync

**Release Type**: Patch Release (Dependency Update)

#### Overview

This release updates Crucible to v0.4.3, keeping gofulmen aligned with the latest SSOT assets.

#### Highlights

- **Crucible v0.4.3** – Updated embedded Crucible assets and `go.mod` dependency.

#### Testing

- `make check-all`
- `make test`

---

## [0.3.0] - 2026-01-07

### Crucible v0.4.2 sync + canonical URI resolution

**Release Type**: Minor Release (Breaking Change)

#### Overview

This release syncs Crucible v0.4.2 which establishes the Canonical URI Resolution Standard. All schema `$id` values now use module-qualified URIs (`schemas.fulmenhq.dev/crucible/...`). The schema resolver in `schema/validator.go` has been updated to explicitly handle only crucible-module schemas.

#### Highlights

- **Crucible v0.4.2** – Updated embedded Crucible assets and `go.mod` dependency.
- **Canonical URIs** – All ~63 schemas updated with `crucible/` module prefix in `$id`.
- **Resolver fix** – `localLoader.Load()` now rejects non-crucible modules with clear error.
- **Fixture Standard** – New standard for test infrastructure repositories.
- **Similarity promotion** – Fixtures and schemas moved from `library/foundry/` to `library/similarity/`.
- **Module schema v1.1.0** – New fields: `weight`, `default_inclusion`, per-language `notes`.

#### Breaking Changes

- Schema `$id` URIs changed: `schemas.fulmenhq.dev/<topic>/...` → `schemas.fulmenhq.dev/crucible/<topic>/...`
- Enact schemas (11 files) removed – moved to enacthq organization.
- Goneat schemas (6 files) removed – moved to goneat repository.
- Similarity paths changed: `library/foundry/` → `library/similarity/`.

The Go API (`foundry/similarity` package) and schema validation are unchanged.

#### Testing

- `make check-all`
- `make test`

---

## Archived releases

Older release notes are archived under `docs/releases/`.
