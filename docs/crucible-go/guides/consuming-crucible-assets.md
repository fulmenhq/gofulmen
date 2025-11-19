---
title: "Consuming Crucible Assets"
description: "Guidance for Fulmen templates and applications on using Crucible schemas, configs, and generated artifacts via helper libraries or dual-hosting workflows"
author: "Schema Cartographer"
date: "2025-11-03"
last_updated: "2025-11-19"
status: "approved"
tags: ["guides", "consumers", "schemas", "provenance", "dual-hosting", "examples"]
---

# Consuming Crucible Assets

Crucible is the SSOT for FulmenHQ schemas, configs, and generated bindings. Helper libraries (`gofulmen`, `pyfulmen`, `tsfulmen`, …) surface those assets so templates and applications can integrate without duplicating logic. This guide explains:

1. How to consume assets directly through helper-library APIs.
2. When and how to **dual-host** schemas/configs inside your own repo while preserving provenance.
3. Recommended checks to detect drift when Crucible versions advance.

> **Audience**: Template repositories (e.g., forge-codex-pulsar) and applications that depend on Fulmen helper libraries.

---

## 0. Practical Examples (Go)

**Problem**: The fulmen-secrets team needs to read `project-secrets.md` from Crucible but doesn't know how.

**Solution**: Use gofulmen's `crucible` package - it provides three generic access functions:

```go
import "github.com/fulmenhq/gofulmen/crucible"

// 1. Read a SCHEMA from schemas/crucible-go/
schemaBytes, err := crucible.GetSchema("devsecops/secrets/v1.0.0/secrets.schema.json")
if err != nil {
    log.Fatal(err)
}

// 2. Read a DOC from docs/crucible-go/
docContent, err := crucible.GetDoc("standards/devsecops/project-secrets.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println(docContent) // Full markdown content

// 3. Read a CONFIG from config/crucible-go/
configBytes, err := crucible.GetConfig("devsecops/secrets/v1.0.0/defaults.yaml")
if err != nil {
    log.Fatal(err)
}
```

### Real Example: fulmen-secrets Reading Project Secrets Documentation

```go
package main

import (
    "fmt"
    "log"
    "strings"
    
    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Get the project secrets documentation (latest version from gofulmen)
    secretsDocs, err := crucible.GetDoc("standards/devsecops/project-secrets.md")
    if err != nil {
        log.Fatalf("Failed to read project-secrets.md: %v", err)
    }
    
    // Get the secrets schema
    secretsSchema, err := crucible.GetSchema("devsecops/secrets/v1.0.0/secrets.schema.json")
    if err != nil {
        log.Fatalf("Failed to read secrets schema: %v", err)
    }
    
    // Get the default configuration
    defaultsConfig, err := crucible.GetConfig("devsecops/secrets/v1.0.0/defaults.yaml")
    if err != nil {
        log.Fatalf("Failed to read secrets defaults: %v", err)
    }
    
    // Now you have all three assets
    fmt.Printf("Documentation: %d characters\n", len(secretsDocs))
    fmt.Printf("Schema: %d bytes\n", len(secretsSchema))
    fmt.Printf("Defaults config: %d bytes\n", len(defaultsConfig))
    
    // Extract the first paragraph from docs
    lines := strings.Split(secretsDocs, "\n")
    fmt.Println("\nFirst 5 lines of documentation:")
    for i := 0; i < 5 && i < len(lines); i++ {
        fmt.Println(lines[i])
    }
}
```

### Path Rules

All paths are relative to their root directory:

| Function | Root Directory | Example Path |
|----------|----------------|--------------|
| `GetSchema()` | `schemas/crucible-go/` | `"devsecops/secrets/v1.0.0/secrets.schema.json"` |
| `GetDoc()` | `docs/crucible-go/` | `"standards/devsecops/project-secrets.md"` |
| `GetConfig()` | `config/crucible-go/` | `"devsecops/secrets/v1.0.0/defaults.yaml"` |

**Do NOT** include the root directory in your path - it's added automatically!

```go
// ✅ CORRECT
crucible.GetDoc("standards/devsecops/project-secrets.md")

// ❌ WRONG - will fail with "not found"
crucible.GetDoc("docs/crucible-go/standards/devsecops/project-secrets.md")
```

### Version Tracking

Always log which Crucible version you're using for support tickets:

```go
version := crucible.GetVersionString()
log.Printf("Using %s", version)
// Output: "gofulmen/v0.1.19 crucible/v0.2.19"
```

---

## 1. Library-First Consumption

Every helper library exposes Crucible catalogs via dedicated modules. Examples:

| Asset                  | Go (`gofulmen`)                                | Python (`pyfulmen`)           | TypeScript (`tsfulmen`)                |
| ---------------------- | ---------------------------------------------- | ----------------------------- | -------------------------------------- |
| Exit codes             | `github.com/fulmenhq/gofulmen/pkg/foundry`     | `pyfulmen.foundry.exit_codes` | `@fulmenhq/tsfulmen/foundry/exitCodes` |
| Signals (planned)      | `github.com/fulmenhq/gofulmen/pkg/signals`     | `pyfulmen.signals`            | `@fulmenhq/tsfulmen/signals`           |
| App identity (planned) | `github.com/fulmenhq/gofulmen/pkg/appidentity` | `pyfulmen.appidentity`        | `@fulmenhq/tsfulmen/appidentity`       |

Python consumers can introspect Crucible provenance and metadata without touching the filesystem:

```python
from pyfulmen.foundry import exit_codes

info = exit_codes.get_exit_code_info(exit_codes.ExitCode.EXIT_CONFIG_INVALID)
print(info["category"])          # => "configuration"
print(exit_codes.EXIT_CODES_VERSION)  # => e.g. "v1.0.0"
```

TypeScript consumers can retrieve the same metadata without leaving the helper boundary:

```typescript
import {
  exitCodes,
  getExitCodeInfo,
  EXIT_CODES_VERSION,
} from "@fulmenhq/tsfulmen/foundry/exitCodes";

const info = getExitCodeInfo(exitCodes.EXIT_CONFIG_INVALID);
console.log(info?.category); // => "configuration"
console.log(EXIT_CODES_VERSION); // => e.g. "v1.0.0"
```

For Go consumers, always import from the `pkg/...` path exposed by `gofulmen`. Crucible may generate root-level bindings for internal use, but the `pkg` re-exports are the compatibility layer we keep stable for templates and applications.

TypeScript validation helpers (e.g., `validateDataBySchemaId()` in `@fulmenhq/tsfulmen/schema/validator`) automatically capture provenance and emit telemetry. Prefer those over bespoke AJV instances so ecosystem metrics and migrations stay aligned.

When possible:

- Call the helper module to retrieve schemas/configs or typed bindings.
- Inspect provenance helpers (`foundry.ExitCodesVersion`, `signals.Version()`, etc.) to log catalog version, revision hash, or last-reviewed date.
- Use the helper’s validation utilities (AJV harness in TypeScript, `goneat` in Python, etc.) to enforce SSOT compliance in your build pipeline.

**Advantages**: No repo-level maintenance; automatic updates when you bump the helper library version; provenance recorded by the module.

---

## 2. Advanced: Dual-Hosting Workflow

Sometimes you need the schema/config **in your repository** for visibility, auditing, or to run tools that require local files (e.g., static site builders, component demos). Follow this workflow to stay aligned with Crucible:

### Step 1 – Export the Asset

- Run the helper’s sync/export command if provided (e.g., `pyfulmen export-schema`, `tsfulmen scripts/sync-schemas`).
- If no helper command exists yet, fetch from Crucible directly:
  ```bash
  curl -sS https://raw.githubusercontent.com/fulmenhq/crucible/<version>/schemas/web/branding/v1.0.0/site-branding.schema.json \
    -o vendor/crucible/schemas/web/branding/v1.0.0/site-branding.schema.json
  ```
- TypeScript projects can wire a small Bun/tsx helper that exports **and validates** in one pass:
  ```bash
  bunx tsx scripts/export-schema.ts \
    --schema-id web/branding/v1.0.0/site-branding \
    --out vendor/crucible/schemas/web/branding/v1.0.0/site-branding.schema.json
  ```
  Inside `export-schema.ts`, call `validateDataBySchemaId()` from `@fulmenhq/tsfulmen/schema/validator` to guarantee the exported file still matches the SSOT snapshot.

### Step 2 – Preserve Provenance

For every dual-hosted file, store metadata alongside it. Options:

- YAML front-matter or comment header:
  ```yaml
  # x-crucible-source:
  #   catalog: web/site-branding
  #   version: v1.0.0
  #   revision: 5ae105bd (exit codes example)
  #   retrieved: 2025-11-03
  ```
- A companion `.provenance.yaml` file per directory summarizing source/version/hash.

### Step 3 – Track in Vendor Space

Keep dual-hosted files under a dedicated directory (`vendor/crucible/…`, `third_party/crucible/…`) so local changes do not blend with your primary SSOT.

Python packages that ship vendored Crucible assets must list those directories in `pyproject.toml` (`[tool.setuptools.package-data]` or equivalent) so `importlib.resources.files()` can locate them at runtime. Forgetting to mark the data files means your production wheels will serve stale or missing catalogs even if the repo copy looks correct.

### Step 4 – Wire Local Validation

Configure your repo to validate both the library-provided copy **and** your dual-hosted file:

- Add CI job (AJV, goneat, etc.) pointing at `vendor/crucible/...`.
- Ensure your build/test pipeline still imports via helper module to catch updates.
- In TypeScript, prefer `validateFileBySchemaId()` or `validateDataBySchemaId()` from `@fulmenhq/tsfulmen/schema/validator` instead of maintaining a parallel AJV instance:

  ```typescript
  import { validateFileBySchemaId } from "@fulmenhq/tsfulmen/schema/validator";

  await validateFileBySchemaId(
    "web/branding/v1.0.0/site-branding",
    "vendor/crucible/schemas/web/branding/v1.0.0/site-branding.schema.json",
  );
  ```

  This keeps telemetry, provenance, and future Crucible schema migrations consistent with the helper library.

```python
# Python CI example
from pathlib import Path

from pyfulmen.schema.validator import validate_file, format_diagnostics

result = validate_file(
    "web/branding/site-branding@v1.0.0",
    Path("vendor/crucible/schemas/web/branding/v1.0.0/site-branding.schema.json"),
    use_goneat=False,
)
if not result.is_valid:
    raise SystemExit(format_diagnostics(result.diagnostics))
```

**Go-specific tip**:

- If you vendor YAML catalogs for bootstrap tooling, embed them with Go’s `//go:embed` in a dedicated package (e.g., `internal/crucibleassets`). Keep the embed path aligned with the helper’s copy so `go test ./...` exercises the same data the helper exposes. Regenerate/refresh the vendored files before each release and document the expected Crucible tag in a `doc.go` header.

**TypeScript-specific tip**:

- Add a vitest (or Bun) parity test that loads the helper’s runtime catalog and compares it to your vendored copy. This catches drift without duplicating parsing logic:

  ```typescript
  import { loadPatternCatalog } from "@fulmenhq/tsfulmen/foundry/loader";
  import vendorCatalog from "../vendor/crucible/config/library/foundry/patterns.json";

  test("vendor catalog matches helper", async () => {
    const canonical = await loadPatternCatalog();
    expect(vendorCatalog).toStrictEqual(canonical);
  });
  ```

---

## 3. Advanced: Drift Detection & Updates

When Crucible publishes an update (new version or schema change):

1. **Bump the helper library** in your repo (`go.mod`, `pyproject.toml`, `package.json`).
2. **Run sync/export** again to refresh dual-hosted files.
3. **Diff** your local copy against the helper/library view:
   - Use helper commands (`pyfulmen diff-schema` upcoming) or build your own script that compares your vendor copy to `tsfulmen`’s in-memory catalog.
   - For Go, add a parity check in `go test` that compares the embedded vendor bytes to `foundry.Catalog()` (when exposed) so CI fails if they diverge silently.
4. **Update provenance metadata** to reflect the new version and revision hash.

Consider adding a periodic CI job that:

- Downloads the latest Crucible schema (for the version you depend on).
- Compares it to your vendor copy and fails if they diverge.

---

## 4. Troubleshooting Common Issues

### "File not found" Error

**Problem**: `crucible.GetDoc("docs/crucible-go/standards/devsecops/project-secrets.md")` returns "file not found"

**Solution**: Remove the `docs/crucible-go/` prefix - the function adds it automatically:

```go
// ❌ WRONG
crucible.GetDoc("docs/crucible-go/standards/devsecops/project-secrets.md")

// ✅ CORRECT  
crucible.GetDoc("standards/devsecops/project-secrets.md")
```

### "Which version of the doc am I getting?"

**Problem**: You updated gofulmen but don't know if you have the latest Crucible docs.

**Solution**: Check the version string and compare to [Crucible releases](https://github.com/fulmenhq/crucible/releases):

```go
fmt.Println(crucible.GetVersionString())
// Output: gofulmen/v0.1.19 crucible/v0.2.19

// This means you have Crucible v0.2.19 embedded in gofulmen v0.1.19
```

### "I need a newer Crucible version"

**Problem**: Crucible v0.2.20 was released but gofulmen is still on v0.2.19.

**Solution**: Wait for gofulmen maintainers to sync, OR temporarily dual-host (see Section 2):

```bash
# Temporary workaround: fetch directly from GitHub
curl -sS https://raw.githubusercontent.com/fulmenhq/crucible/v0.2.20/docs/standards/devsecops/project-secrets.md \
  -o vendor/crucible/docs/project-secrets.md
```

Then update once gofulmen syncs to v0.2.20.

### "How do I know what paths are available?"

**Problem**: You don't know the exact path to a schema or doc.

**Solution**: Use `ListSchemas()` or browse the [Crucible repository](https://github.com/fulmenhq/crucible):

```go
// List all schemas in devsecops/secrets/v1.0.0/
schemas, err := crucible.ListSchemas("devsecops/secrets/v1.0.0")
if err != nil {
    log.Fatal(err)
}

for _, name := range schemas {
    fmt.Println(name)
}
// Output:
// secrets.schema.json
// defaults.yaml
```

Or check gofulmen's embedded files:
```bash
# In gofulmen repository
ls schemas/crucible-go/devsecops/secrets/v1.0.0/
ls docs/crucible-go/standards/devsecops/
ls config/crucible-go/devsecops/secrets/v1.0.0/
```

---

## 5. Frequently Asked Questions

### Can we bypass helper libraries entirely?

Not recommended. Helper libraries encode language-specific behaviors (validation, code generation) and establish provenance. Dual-hosting is intended as a **cache**, not a replacement.

### How do we know when Crucible updates a catalog?

- Watch Crucible releases: new tags update the `VERSION` file.
- Helper modules expose version constants (`EXIT_CODES_VERSION`, `signals.Version()`).
- Consider subscribing to repository notifications or automation that checks `schemas/**` directories for changes.

### What about downstream forks/customizations?

If you alter the schema locally, you must:

- Rename your schema (e.g., `x-fulmen/…`) to avoid conflicting with SSOT paths.
- Document the divergence in your repo (README, changelog).
- Avoid upstreaming unless the change is intended for all consumers—submit a PR to Crucible instead.

---

## 6. Next Steps

- Helper-library owners: audit modules to ensure they expose provenance (version/hash) and export commands.
- Templates/apps: integrate this guide into your onboarding docs so contributors know how to dual-host responsibly.
- Crucible maintainers: keep this guide updated as new helper commands or validation tooling ship.

For questions or suggestions, open an issue in Crucible and cc @schema-cartographer.
