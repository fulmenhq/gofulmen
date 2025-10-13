# Foundry Asset Sync - Post-Sync Hook Extension

## Overview

The Foundry package uses Go's `embed` directive to bundle Crucible config files at compile time, enabling offline operation in compiled binaries. However, Go's embed cannot access parent directories (`..`), so Foundry YAML files must be copied from `config/crucible-go/library/foundry/` to `foundry/assets/` for embedding.

This document describes the post-sync hook pattern we use to keep these embedded files synchronized with the Crucible SSOT.

## Problem

**Standard SSOT Sync:** `make sync` runs `goneat ssot sync` to update `config/crucible-go/` from the Crucible SSOT repository.

**Embed Limitation:** Go's `//go:embed` directive cannot use relative paths like `../config/crucible-go/library/foundry/*.yaml` to access parent directories.

**Solution Required:** After each sync, Foundry assets must be copied to `foundry/assets/` so the embed directive `//go:embed assets/*.yaml` works.

## Implementation

### Makefile Post-Sync Hook

We extended the `sync` target to automatically copy Foundry assets after the standard SSOT sync:

```makefile
sync: ## Sync assets from Crucible SSOT
	@if [ ! -f $(GONEAT) ]; then \
		echo "❌ goneat not found. Run 'make bootstrap' first."; \
		exit 1; \
	fi
	@echo "Syncing assets from Crucible..."
	@$(GONEAT) ssot sync
	@$(MAKE) sync-foundry-assets    # <-- Post-sync hook
	@echo "✅ Sync completed"

sync-foundry-assets: ## Copy foundry YAML assets to embedded location (post-sync hook)
	@echo "Copying foundry assets for embed..."
	@mkdir -p foundry/assets
	@cp config/crucible-go/library/foundry/*.yaml foundry/assets/
	@echo "✅ Foundry assets synchronized"
```

### Usage

**Developer workflow:**
```bash
# Standard sync (automatically copies foundry assets)
make sync

# Manual foundry asset sync (if needed)
make sync-foundry-assets
```

**CI/CD workflow:**
```bash
# Run sync before build/test to ensure latest assets
make sync
make test
```

## Pattern Benefits

1. **Automatic Synchronization:** Assets stay in sync without manual intervention
2. **Standard Compliance:** Uses standard `make sync` command
3. **Explicit Target:** `sync-foundry-assets` can be called independently
4. **Source of Truth:** `config/crucible-go/` remains the SSOT, `foundry/assets/` is derived

## Asset Files

The following files are synchronized:

- `config/crucible-go/library/foundry/patterns.yaml` → `foundry/assets/patterns.yaml`
- `config/crucible-go/library/foundry/mime-types.yaml` → `foundry/assets/mime-types.yaml`
- `config/crucible-go/library/foundry/http-statuses.yaml` → `foundry/assets/http-statuses.yaml`
- `config/crucible-go/library/foundry/country-codes.yaml` → `foundry/assets/country-codes.yaml`

## Verification

After running `make sync`, verify assets are current:

```bash
# Check timestamps match
ls -l config/crucible-go/library/foundry/*.yaml
ls -l foundry/assets/*.yaml

# Verify content matches
diff config/crucible-go/library/foundry/patterns.yaml foundry/assets/patterns.yaml
```

## Git Handling

**`.gitignore` Status:** `foundry/assets/` is **tracked** in git (not gitignored) because:
- Embedded files are required for build
- Allows git to detect sync drift
- Enables code review of SSOT changes

**Commit Practice:** When `make sync` updates Crucible assets, commit both locations together:
```bash
git add config/crucible-go/ foundry/assets/
git commit -m "sync: update Crucible assets to vX.Y.Z"
```

## Extending the Pattern

If other packages need embedded Crucible assets, follow this pattern:

1. **Create package assets directory:** `mkdir -p <package>/assets`
2. **Add embed directive:** `//go:embed assets/*.yaml` in package code
3. **Add sync target:**
   ```makefile
   sync-<package>-assets:
       @cp config/crucible-go/<path>/*.yaml <package>/assets/
   ```
4. **Hook into sync:** Add `@$(MAKE) sync-<package>-assets` to `sync` target
5. **Document:** Create `docs/development/<package>-asset-sync.md`

## Related Standards

- [Crucible SSOT Sync Model](../crucible-go/architecture/sync-model.md)
- [Makefile Standard](../crucible-go/standards/makefile-standard.md)
- [Foundry Library Standard](../crucible-go/standards/library/foundry/)

## Maintenance Notes

- **Audit Reminder:** Always verify `foundry/assets/` matches `config/crucible-go/library/foundry/` after sync
- **CI Check:** Consider adding a CI check to ensure assets are synchronized
- **Future Enhancement:** Could automate verification with a `make verify-sync` target
