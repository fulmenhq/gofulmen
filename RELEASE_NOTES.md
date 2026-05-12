# Release Notes

This document tracks release notes and checklists for gofulmen releases.

> **Convention**: Keep only latest 3 releases here to prevent file bloat. Older releases are archived in `docs/releases/`.

## [0.3.5] - 2026-05-12

### Release hygiene, appidentity precedence, and dependency refresh

**Release Type**: Patch Release (Reliability + Maintenance)

#### Overview

This release tightens the v0.3.x foundation before downstream adoption: embedded appidentity now wins over ambient checkout discovery, CI and local lint formatting are aligned, release-blocking high security findings are cleared, and the direct dependency wave is current.

No new exported config utility primitives are included in v0.3.5. That API design was deferred to v0.3.6 so the semantics can be settled against confirmed downstream needs without expanding the v0.3.5 release scope.

#### Highlights

- **appidentity**: Embedded identities now short-circuit CWD and executable-directory ancestor discovery, so packaged binaries do not self-identify as a nearby foreign repository.
- **security**: Added checked ZIP size conversions, permission-mode normalization, validated path/binary resolution, CRC byte serialization helpers, and related hardening to clear release-blocking high findings.
- **CI/tooling**: Updated the Fulmen toolbox runner, aligned `.yamlfmt`/`.yamllint` behavior with goneat, and kept the external installation test active on pull requests.
- **Git hooks**: Regenerated goneat hooks without guardian browser interception now that the repository is moving from direct-push safeguards toward merge-policy protection.
- **Lint cleanup**: Moved larger Make recipes into scripts and replaced `WriteString(fmt.Sprintf(...))` patterns with direct `fmt.Fprintf` calls.
- **Dependency updates**:
  - `go.uber.org/zap` v1.27.1 → v1.28.0
  - `github.com/mattn/go-runewidth` v0.0.20 → v0.0.23
  - `golang.org/x/mod` v0.33.0 → v0.36.0
  - `golang.org/x/text` v0.34.0 → v0.37.0
  - `golang.org/x/time` v0.14.0 → v0.15.0

#### For Library Consumers

Consumers that embed appidentity data in shipped binaries get safer default behavior: the embedded identity is preferred over ambient filesystem discovery unless an explicit path or `FULMEN_APP_IDENTITY_PATH` is provided.

No public API migration is required for v0.3.5.

#### Testing

- `make check-all`
- `make precommit`
- Pull-request CI: `Test (container)` and `External Installation Test`

---

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

## Archived releases

Older release notes are archived under `docs/releases/`.
