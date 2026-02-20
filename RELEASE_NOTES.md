# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.3.4] - 2026-02-20

### Typed role catalog API + dependency updates

**Release Type**: Patch Release (Feature + Maintenance)

#### Overview

This release adds the typed role catalog API to the crucible shim — exposing `LoadRole`, `ListRoleSlugs`, and `LoadRoleCatalog` with full Go types — and brings all direct dependencies to their latest versions. Crucible is updated to v0.4.12, which adds 3 new roles and the `domains` field across the catalog.

#### Highlights

- **crucible**: Typed role catalog shim re-exports from Crucible v0.4.12.
  - `LoadRole(slug)` — load and parse a single role by slug.
  - `ListRoleSlugs()` — sorted list of all available role slugs.
  - `LoadRoleCatalog()` — full catalog as `map[string]*RolePrompt`.
  - 7 type aliases: `RolePrompt`, `RoleMindset`, `RoleEscalation`, `RoleExample`, `RoleRequiredReading`, `RoleRequiredReadingFile`, `AgenticConfig`.
  - Raw YAML access via `ConfigRegistry.Agentic().Role(slug)`.
  - 12 tests exercising the shim path with invariant-based assertions.
- **Crucible v0.4.12**: 3 new roles (`cxotech`, `deliverylead`, `infraeng`), `domains` field on all roles.
- **Dependency updates**: All direct dependencies updated to latest (0 vulnerabilities).
  - zap v1.27.1, go-runewidth v0.0.20, x/mod v0.33.0, x/text v0.34.0, doublestar v4.10.0, xxh3 v1.1.0, testify v1.11.1.
- **Documentation**: Expanded crucible/README.md role catalog section with field reference table, per-task snippets, and fixed stale version numbers.
- **AGENTS.md**: Added explicit commit trailer template to prevent omitted `Role:` and `Committer-of-Record:` lines.

#### For Library Consumers

Access the role catalog through the gofulmen crucible shim:

```go
import "github.com/fulmenhq/gofulmen/crucible"

// Load a single role
role, err := crucible.LoadRole("devlead")
fmt.Printf("Role: %s — %s\n", role.Name, role.Description)

// List all available roles
slugs, _ := crucible.ListRoleSlugs()

// Load full catalog
catalog, _ := crucible.LoadRoleCatalog()

// Access orchestration fields
if role.Mindset != nil {
    fmt.Println("Focus:", role.Mindset.Focus)
}
for _, e := range role.EscalatesTo {
    fmt.Printf("Escalate to %s when: %s\n", e.Target, e.When)
}
```

#### Testing

- `make check-all`
- `make test` (all 27 packages pass)
- `goneat dependencies --vuln` (0 findings)

---

## [0.3.3] - 2026-02-04

### Control-plane hardening + diagnostics primitives

**Release Type**: Patch Release (Security Hardening)

#### Overview

This release hardens the `signals` HTTP admin endpoint and Prometheus exporter auth for control-plane usage, and adds small diagnostics primitives (env var alias/conflict reporting, sensitive-key masking) to reduce duplication across Fulmen templates.

#### Highlights

- **signals**: add `HTTPConfig.AllowedSignals` allowlist (optional; nil/empty preserves existing behavior).
- **signals**: add `HTTPConfig.AllowClientGracePeriod` (default `false`) to ignore client-provided `grace_period_seconds` unless explicitly enabled.
- **signals + telemetry/exporters**: bearer token parsing + constant-time compare.
- **config**: add env var alias specs + conflict diagnostics (`LoadEnvOverridesWithReport`); sensitive values are masked by default.
- **logging**: add `IsSensitiveKey(name string) bool` helper for envinfo/doctor-style masking.
- **foundry + crucible**: version reporting now reflects the actual installed gofulmen module version in downstream binaries.
- **Crucible v0.4.10**: synced SSOT assets and updated dependency.

#### Behavior change note

- `signals.HTTPHandler` no longer honors client `grace_period_seconds` by default. If you relied on that behavior, set `AllowClientGracePeriod: true`.

#### Testing

- `make fmt`
- `make test`
- `make lint`

---

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

## Archived releases

Older release notes are archived under `docs/releases/`.
