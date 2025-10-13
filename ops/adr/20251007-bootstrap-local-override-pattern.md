# ADR-001: Bootstrap Local Override Pattern

**Date**: 2025-10-07  
**Status**: Accepted

## Context

Gofulmen is a foundation library that other tools (goneat, fulward) depend on. We need DX tooling (FulDX) for version management and SSOT sync, but face the "bootstrap the bootstrap" problem:

1. We can't depend on goneat (circular dependency)
2. We need FulDX during active development (rapid iteration)
3. FulDX binaries aren't published yet
4. We can't commit local absolute paths to `.crucible/tools.yaml`

## Decision

Implement a **local tools override pattern** with these components:

### 1. Manifest Resolution Order

Bootstrap prefers `.crucible/tools.local.yaml` if present, falls back to `tools.yaml`:

```go
func resolveManifestPath(defaultPath string) string {
    localPath := strings.Replace(defaultPath, ".yaml", ".local.yaml", 1)
    if _, err := os.Stat(localPath); err == nil {
        return localPath
    }
    return defaultPath
}
```

### 2. Local Override Template

Provide `.crucible/tools.local.yaml.example` (committed):

```yaml
version: v1.0.0
binDir: ./bin
tools:
  - id: fuldx
    install:
      type: link
      source: /Users/yourname/dev/fulmenhq/fuldx/dist/fuldx
      binName: fuldx
```

### 3. Gitignore Local Overrides

Never commit `.crucible/tools.local.yaml`:

```gitignore
# Local tool overrides (never commit - used for development iteration)
.crucible/tools.local.yaml
```

### 4. New Install Type: `type: link`

Add support for copying local binaries:

```go
// bootstrap/install_link.go
func installLink(tool *Tool) error {
    // Copy source to destination
    // Make executable
}
```

## Consequences

### Positive

- ✅ **Fast local iteration** - No rebuild/republish cycle for FulDX changes
- ✅ **Safe** - Local paths never committed (gitignored)
- ✅ **Flexible** - Can work on FulDX and gofulmen simultaneously
- ✅ **Self-documenting** - Example file shows pattern clearly
- ✅ **Production ready** - Same pattern works with published binaries (just remove local override)

### Negative

- ⚠️ **Extra file** - Adds `.crucible/tools.local.yaml.example` to repo
- ⚠️ **Manual setup** - Contributors must copy and edit example file
- ⚠️ **Schema extension needed** - Requires Crucible to add `type: link` and `source` field

### Neutral

- Uses copy instead of symlink (cross-platform compatibility)
- Bootstrap tool now has manifest resolution logic
- Pattern can be reused by other `*fulmen` libraries (tsfulmen, pyfulmen)

## Alternatives Considered

### Option 1: Assume FulDX in PATH

```bash
# User must manually install
brew install fuldx  # (if we published to brew)
```

**Rejected**: Breaks CI/CD, requires manual steps, version drift

### Option 2: Manual Binary Copy

```bash
# Manual script
cp ../fuldx/dist/fuldx bin/fuldx
chmod +x bin/fuldx
```

**Rejected**: Not automated, no integration with bootstrap, error-prone

### Option 3: Git Submodules

```bash
# Add fuldx as submodule
git submodule add https://github.com/fulmenhq/fuldx
```

**Rejected**: Too heavy, complicates repo, doesn't solve build issue

## Implementation

1. ✅ Add `Source` field to `Install` struct
2. ✅ Create `bootstrap/install_link.go`
3. ✅ Implement `resolveManifestPath()`
4. ✅ Create `.crucible/tools.local.yaml.example`
5. ✅ Update `.gitignore`
6. ⏳ Crucible team updates schema (pending)
7. ⏳ Precommit hook to prevent local path leaks (future)

## References

- [Bootstrap Strategy](bootstrap-strategy.md)
- [FULDX.md](../docs/FULDX.md)
- Crucible Schema Update Memo: `.plans/memos/crucible-bootstrap-schema-update.md`

## Supersedes

None (first ADR)

## Related

This pattern is now part of the **Fulmen Helper Library Standard** and should be adopted by:

- tsfulmen (TypeScript foundation library)
- pyfulmen (Python foundation library, future)
- Other language-specific foundation libraries
