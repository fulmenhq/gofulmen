---
id: "ADR-0007"
title: "Crucible Version Synchronization and Verification"
status: "accepted"
date: "2025-11-19"
deciders: ["@3leapsdave", "@foundation-forge"]
scope: "gofulmen"
tags: ["crucible", "dependencies", "testing", "quality", "process"]
---

## Context

Gofulmen embeds Crucible assets (docs, schemas, configs) via direct imports and SSOT sync, unlike other Fulmen libraries that only reference Crucible. This creates a three-way synchronization requirement:

1. **`.goneat/ssot-consumer.yaml`** - Declares which Crucible version to sync
2. **`.goneat/ssot/provenance.json`** - Records when/what was synced
3. **`go.mod`** - Declares which Crucible version to import at runtime

**Problem**: On two separate occasions (v0.1.17 and v0.1.18), releases shipped with mismatched versions:

- SSOT sync pulled v0.2.19 assets (docs/schemas/configs)
- But `go.mod` still required v0.2.18 for runtime imports
- Result: Embedded documentation referenced v0.2.19 features but runtime behavior used v0.2.18 code

This mismatch causes:

- Incorrect version reporting via `crucible.GetVersionString()`
- Potential runtime errors if schemas/APIs changed between versions
- Confusion for consumers trying to debug behavior
- Loss of confidence in our release process

## Decision

Implement a three-part solution to prevent and detect Crucible version mismatches:

### 1. Automated Workflow: `make crucible-update`

Create a single Makefile target that atomically updates all three synchronization points:

```makefile
crucible-update: ## Update Crucible dependency to specific version (usage: make crucible-update VERSION=v0.2.19)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION not specified. Usage: make crucible-update VERSION=v0.2.19"; \
		exit 1; \
	fi
	@echo "Updating Crucible to $(VERSION)..."
	@echo ""
	@echo "Step 1: Updating .goneat/ssot-consumer.yaml..."
	@sed -i.bak 's|ref: v[0-9]*\.[0-9]*\.[0-9]*|ref: $(VERSION)|' .goneat/ssot-consumer.yaml && rm .goneat/ssot-consumer.yaml.bak
	@echo "✅ Updated ssot-consumer.yaml ref to $(VERSION)"
	@echo ""
	@echo "Step 2: Running make sync to update provenance..."
	@$(MAKE) sync
	@echo ""
	@echo "Step 3: Updating go.mod..."
	@go get github.com/fulmenhq/crucible@$(VERSION)
	@go mod tidy
	@echo "✅ Updated go.mod to $(VERSION)"
	@echo ""
	@echo "Step 4: Running tests to verify compatibility..."
	@go test ./crucible -run TestCrucibleVersionMatchesMetadata -v
	@echo ""
	@echo "✅ Crucible updated successfully to $(VERSION)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review changes: git diff"
	@echo "  2. Run full checks: make check-all"
	@echo "  3. Commit changes with proper attribution"
```

**Benefits**:

- Single command updates all three points atomically
- Self-documenting via echo statements
- Runs verification test automatically
- Provides clear next steps for developer

### 2. Guardrail Test: `TestCrucibleVersionMatchesMetadata`

Create a test that fails CI/builds when versions are mismatched:

```go
func TestCrucibleVersionMatchesMetadata(t *testing.T) {
	// Find repository root using pathfinder (dogfooding)
	cwd, _ := os.Getwd()
	root, _ := pathfinder.FindRepositoryRoot(cwd, pathfinder.GoModMarkers)

	// Read Crucible version from go.mod
	goModVersion, _ := readCrucibleVersionFromGoMod(filepath.Join(root, "go.mod"))

	// Read Crucible version from metadata
	metadataVersion, _ := readCrucibleVersionFromMetadata(
		filepath.Join(root, ".crucible", "metadata", "metadata.yaml"))

	// Normalize versions (metadata: "0.2.19", go.mod: "v0.2.19")
	normalizedGoMod := strings.TrimPrefix(goModVersion, "v")
	normalizedMetadata := strings.TrimPrefix(metadataVersion, "v")

	// They must match
	if normalizedGoMod != normalizedMetadata {
		// Error message formatted using ascii.DrawBox() for clarity
		box := ascii.DrawBox(errorMessage, 0)
		t.Fatalf("\n%s", box)
	}
}
```

**Key Design Decisions**:

- **Dogfoods gofulmen libraries**: Uses `pathfinder.FindRepositoryRoot()` and `ascii.DrawBox()`
- **Version normalization**: Handles format mismatch (go.mod has "v" prefix, metadata doesn't)
- **Clear error messages**: ASCII-boxed output shows exact mismatch and remediation steps
- **Runs in CI**: Part of `make check-all` and `make test`, blocks bad releases

**Dependencies**:

- `golang.org/x/mod/modfile` - Parse go.mod without shell commands
- `gopkg.in/yaml.v3` - Parse metadata.yaml
- `github.com/fulmenhq/gofulmen/pathfinder` - Repository root discovery
- `github.com/fulmenhq/gofulmen/ascii` - Formatted error output

### 3. Manual Verification Guide

When reviewing a PR or troubleshooting, developers can verify version alignment manually:

#### Quick Check (3 locations)

```bash
# 1. Check ssot-consumer.yaml sync ref
grep "ref:" .goneat/ssot-consumer.yaml
# Expected: ref: v0.2.19

# 2. Check go.mod require line
grep "github.com/fulmenhq/crucible" go.mod
# Expected: github.com/fulmenhq/crucible v0.2.19

# 3. Check metadata synced version
grep "version:" .crucible/metadata/metadata.yaml | head -2
# Expected: version: 0.2.19 (note: no 'v' prefix)
```

All three should show the same semantic version (ignoring 'v' prefix differences).

#### Deep Verification (provenance)

```bash
# Check provenance record for sync timestamp and commit hash
jq '.sources[] | select(.name=="crucible") | {version, commit, generated_at}' \
  .goneat/ssot/provenance.json

# Expected output:
# {
#   "version": "0.2.19",
#   "commit": "f17e5fa553d9bef5be4a7b323651ca893978e364",
#   "generated_at": "2025-11-19T17:44:59.707715Z"
# }
```

Verify:

- `version` matches go.mod (without 'v')
- `commit` hash matches the Crucible v0.2.19 tag
- `generated_at` is recent (indicates fresh sync)

#### Automated Verification

```bash
# Run the guardrail test directly
go test ./crucible -run TestCrucibleVersionMatchesMetadata -v

# Or run full quality checks (includes the test)
make check-all
```

## Consequences

### Positive

1. **Prevents Release Errors**: Guardrail test blocks releases with version mismatches
2. **Atomic Updates**: Single `make crucible-update` command prevents partial updates
3. **Self-Documenting**: Test error messages show exact problem and fix
4. **Dogfooding**: Solution uses our own pathfinder and ASCII libraries
5. **CI Integration**: Runs automatically on every PR via `make check-all`
6. **Clear Manual Verification**: Three-point check is fast and unambiguous

### Negative

1. **Additional Test Dependency**: Requires `golang.org/x/mod/modfile` package
2. **Makefile Complexity**: Adds 30-line target (but well-documented)
3. **Learning Curve**: Developers must use `make crucible-update` instead of manual `go get`

### Neutral

1. **Version Normalization**: Test must strip 'v' prefix to handle format differences
2. **Provenance Churn**: `make sync` updates timestamp even when version unchanged
   - Acceptable: Provenance timestamp indicates freshness
   - Mitigation: Only run sync when actually updating Crucible version

## Implementation Notes

### Test Placement

Located in `crucible/version_guard_test.go` because:

- Tests the crucible package's version reporting
- Natural location for Crucible-related quality checks
- Runs as part of `go test ./crucible`

### Error Message Format

Uses `ascii.DrawBox()` for consistent, professional formatting:

```
┌──────────────────────────────────────────────────────────────────────┐
│ CRITICAL: Crucible Version Mismatch Detected                         │
│ ════════════════════════════════════════════════════════════════════ │
│                                                                      │
│ go.mod requires:     github.com/fulmenhq/crucible v0.2.18            │
│ metadata synced:     0.2.19                                          │
│                                                                      │
│ This means your embedded assets (docs/schemas/configs) are from      │
│ Crucible 0.2.19 but runtime imports will use v0.2.18                 │
│                                                                      │
│ ════════════════════════════════════════════════════════════════════ │
│ TO FIX:                                                              │
│ ════════════════════════════════════════════════════════════════════ │
│                                                                      │
│ After running 'make sync', you must ALSO update go.mod:              │
│                                                                      │
│   go get github.com/fulmenhq/crucible@0.2.19                         │
│   go mod tidy                                                        │
│   make check-all                                                     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

Clear separation between problem description and remediation steps.

### Future Enhancements

1. **Pre-commit Hook**: Add guardrail test to pre-commit to catch mistakes before commit
2. **Version Bump Integration**: Extend `make crucible-update` to optionally bump gofulmen version
3. **Changelog Automation**: Auto-generate changelog entry when Crucible version changes
4. **Diff Report**: Show what changed between Crucible versions (schemas/docs modified)

## References

- [Crucible Repository](https://github.com/fulmenhq/crucible)
- [Goneat SSOT Sync Documentation](../../GONEAT.md)
- [Repository Safety Protocols](../../REPOSITORY_SAFETY_PROTOCOLS.md)
- [Feature Brief: v0.1.19 Crucible Version Mismatch Fix](../../../.plans/active/v0.1.19/crucible-version-mismatch-fix.md)

## Related ADRs

- ADR-0004: Crucible Runtime Dependency (established the embedding pattern)

## Lessons Learned

1. **Process Over Memory**: Automated workflows prevent human error better than documentation
2. **Fail Fast**: Tests that catch mistakes before release are invaluable
3. **Dogfooding Works**: Using our own libraries (pathfinder, ASCII) improved test quality
4. **Clear Error Messages**: Investing in formatted, actionable errors pays off immediately

---

**Status**: Accepted and implemented in v0.1.19  
**Approvers**: @3leapsdave (human supervisor), Foundation Forge (AI co-maintainer)
