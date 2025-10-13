# Bootstrap Strategy for Gofulmen

## The Problem: Bootstrap the Bootstrap

Gofulmen faces a unique challenge in the Fulmen ecosystem:

1. **We're a foundation library** that other tools (like goneat, fulward) depend on
2. **We need DX tooling** (fuldx for version management, SSOT sync)
3. **We can't create circular dependencies** (goneat depends on gofulmen, but goneat is the sophisticated tool manager)
4. **We need to work in CI/CD** without manual installation steps

## The Solution: Minimal FulDX Bootstrap + Synced Assets

### Key Principles

1. **Commit synced assets** (docs, schemas, configs from Crucible)
   - Pattern: `docs/crucible-go/`, `schemas/crucible-go/`
   - Regenerated via `make sync` but committed to repo
   - Provides self-contained documentation and schemas
   - Matches fuldx pattern (`docs/crucible-ts/`, `schemas/crucible-ts/`)

2. **Single entry in tools.yaml** for fuldx
   - Platform-specific URLs and checksums
   - Matches Crucible's approach for goneat
   - Works in CI/CD without additional setup

3. **Local override for development iteration**
   - `.crucible/tools.local.yaml` (gitignored)
   - Points to local fuldx build during active development
   - Precommit hook prevents leaking local paths

## Bootstrap Hierarchy

```
┌─────────────────────────────────────────────┐
│  git, go (system dependencies)              │
│  ✓ Verified via tools.yaml                  │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  gofulmen/bootstrap (Go-based)              │
│  ✓ Minimal tool installer                   │
│  ✓ No external dependencies                 │
│  ✓ Reads .crucible/tools.yaml               │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  fuldx (microtool - rarely changes)         │
│  ✓ Installed via bootstrap                  │
│  ✓ Version management (CalVer/SemVer)       │
│  ✓ SSOT sync from Crucible                  │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  goneat (sophisticated tool manager)        │
│  ✓ Installed via bootstrap                  │
│  ✓ Formatting, linting, security            │
│  ✓ Complex tool orchestration               │
└─────────────────────────────────────────────┘
```

## Workflows

### CI/CD (Production)

```bash
# 1. Clone repo (includes synced assets)
git clone https://github.com/fulmenhq/gofulmen

# 2. Bootstrap tools (uses tools.yaml)
make bootstrap

# 3. Sync latest from Crucible (optional, uses committed assets otherwise)
make sync

# 4. Run tests, build, etc.
make test
```

**Why this works:**

- ✅ Self-contained (synced assets committed)
- ✅ No manual installation steps
- ✅ Deterministic (checksums in tools.yaml)
- ✅ Fast (no dynamic discovery)

### Local Development (Iteration on Gofulmen + FulDX)

```bash
# 1. Clone both repos
git clone https://github.com/fulmenhq/gofulmen
git clone https://github.com/fulmenhq/fuldx

# 2. Build fuldx locally
cd fuldx && bun install && bun run build

# 3. Create local tools override
cd ../gofulmen
cp .crucible/tools.local.yaml.example .crucible/tools.local.yaml
# Edit to point to local fuldx build

# 4. Bootstrap (uses local override)
make bootstrap

# 5. Iterate
# - Make changes to fuldx
# - Rebuild fuldx
# - Test in gofulmen via `make sync`, `make version-bump`, etc.
```

**Why this works:**

- ✅ Fast iteration (no publish/download cycle)
- ✅ Safe (tools.local.yaml is gitignored)
- ✅ Flexible (can work on fuldx features needed by gofulmen)

### Local Development (Stable FulDX)

```bash
# 1. Clone repo
git clone https://github.com/fulmenhq/gofulmen

# 2. Bootstrap (uses tools.yaml, downloads released fuldx)
make bootstrap

# 3. Develop gofulmen
# - Use committed synced assets
# - Run `make sync` to update from Crucible if needed
```

## File Structure

```
gofulmen/
├── .crucible/
│   ├── tools.yaml                 # Production tool definitions (committed)
│   ├── tools.local.yaml          # Local overrides (gitignored)
│   ├── tools.local.yaml.example  # Example local config (committed)
│   ├── metadata/                 # Sync metadata (committed, from fuldx sync)
│   │   ├── sync-keys.yaml
│   │   └── sync-consumer-config.yaml
├── .fuldx/
│   └── sync-consumer.yaml        # FulDX sync configuration (committed)
├── docs/
│   ├── crucible-go/              # Synced docs (committed, regenerated via sync)
│   │   ├── architecture/
│   │   ├── standards/
│   │   └── guides/
│   ├── FULDX.md                  # FulDX integration guide
│   └── BOOTSTRAP-STRATEGY.md     # This document
├── schemas/
│   └── crucible-go/              # Synced schemas (committed, regenerated)
│       ├── config/
│       ├── observability/
│       └── pathfinder/
├── config/
│   └── crucible-go/              # Synced configs (committed, regenerated)
│       └── terminal/
└── bootstrap/                    # Minimal Go-based tool installer
    └── *.go
```

## Comparison: Option 1 vs Option 2

### Option 1: Assume FulDX in PATH

```bash
# User must do:
brew install fuldx  # (if we published to brew)
# or
curl -L https://github.com/.../fuldx-... -o /usr/local/bin/fuldx
```

**Pros:**

- ✅ Simple Makefile (just call `fuldx`)

**Cons:**

- ❌ Manual installation step
- ❌ Breaks in CI/CD (must install before checkout)
- ❌ Version drift (users may have different fuldx versions)
- ❌ Not self-contained

### Option 2: Bootstrap FulDX via tools.yaml (Chosen)

```yaml
# .crucible/tools.yaml
- id: fuldx
  install:
    type: download
    url: https://github.com/fulmenhq/fuldx/releases/download/v0.1.1/fuldx-{{os}}-{{arch}}
    checksum:
      darwin-arm64: abc123...
      linux-amd64: def456...
```

**Pros:**

- ✅ Self-contained (everything in repo)
- ✅ Works in CI/CD (no manual steps)
- ✅ Version pinned (reproducible)
- ✅ Local override for iteration
- ✅ Matches Crucible pattern

**Cons:**

- ⚠️ Need to maintain checksums (one-time per release)
- ⚠️ Need local override pattern for dev (solved via tools.local.yaml)

## Safety: Preventing Local Path Leaks

### Precommit Hook (TODO)

```bash
# Check for local paths in tools.yaml
if grep -q "source:.*Users" .crucible/tools.yaml; then
  echo "❌ Local path found in tools.yaml"
  echo "Use tools.local.yaml for local development"
  exit 1
fi
```

### Best Practices

1. **Never edit tools.yaml with local paths**
   - Use `tools.local.yaml` instead
2. **Don't commit tools.local.yaml**
   - Already in `.gitignore`
3. **Use the example as template**
   - Copy `tools.local.yaml.example` → `tools.local.yaml`

## Why This Works for Gofulmen

1. **Foundation library status**
   - Other tools depend on us, not vice versa
   - FulDX is a microtool, not a library dependency
   - No circular import issues

2. **Minimal external tooling**
   - Only need: git, go, fuldx, goneat
   - All manageable via simple bootstrap script
   - No need for sophisticated tool orchestration (that's goneat's job)

3. **Committed synced assets**
   - Docs, schemas, configs always available
   - No network dependency for development
   - Regenerated via `make sync` when needed

4. **Matches ecosystem patterns**
   - Crucible: Uses bootstrap for goneat
   - FulDX: Commits synced assets (`docs/crucible-ts/`)
   - Gofulmen: Combines both patterns

## Future Evolution

### When FulDX Stabilizes

1. Uncomment fuldx in `tools.yaml`
2. Add real checksums
3. Remove "coming soon" warnings
4. Document as primary workflow

### If FulDX Complexity Grows

Consider:

- Publishing to package managers (brew, scoop)
- Standalone installer script
- Docker container with all tools

But for now: **Option 2 is the right choice** ✅

## References

- [Crucible Bootstrap Pattern](../../crucible/.crucible/tools.yaml)
- [FulDX Sync Pattern](../../fuldx/.fuldx/sync-consumer.yaml)
- [Fulmen Helper Library Standard](docs/crucible-go/architecture/fulmen-helper-library-standard.md)
- [FULDX Integration Guide](FULDX.md)
