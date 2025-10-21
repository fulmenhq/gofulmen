---
title: "Gofulmen Library Overview"
description: "Comprehensive overview of the Go foundation library for FulmenHQ ecosystem"
author: "Foundation Forge"
date: "2025-10-11"
last_updated: "2025-10-11"
status: "active"
tags: ["overview", "library", "go", "foundation"]
---

# Gofulmen Library Overview

## Purpose & Scope

Gofulmen is the Go foundation library for the FulmenHQ ecosystem, providing enterprise-grade packages for configuration management, structured logging, schema validation, and developer tooling. It serves as the canonical implementation of Crucible standards for Go applications, ensuring consistency, reliability, and scalability across the ecosystem.

**Target Audience**: Go developers building CLI tools, API services, background workers, and enterprise applications within the FulmenHQ ecosystem.

**Design Philosophy**: Progressive complexity with enterprise-grade defaults. Simple use cases require minimal configuration, while complex applications have access to full enterprise features.

## Crucible Overview

**What is Crucible?**

Crucible is the FulmenHQ single source of truth (SSOT) for schemas, standards, and configuration templates. It ensures consistent APIs, documentation structures, and behavioral contracts across all language foundations (gofulmen, pyfulmen, tsfulmen, etc.).

**Why the Shim & Docscribe Module?**

Rather than copying Crucible assets into every project, helper libraries provide idiomatic access through shim APIs. This keeps your application lightweight, versioned correctly, and aligned with ecosystem-wide standards. The docscribe module lets you discover, parse, and validate Crucible content programmatically without manual file management.

**Where to Learn More:**

- [Crucible Repository](https://github.com/fulmenhq/crucible) - SSOT schemas, docs, and configs
- [Fulmen Technical Manifesto](crucible-go/architecture/fulmen-technical-manifesto.md) - Philosophy and design principles
- [SSOT Sync Standard](crucible-go/standards/library/modules/ssot-sync.md) - How libraries stay synchronized

## Module Catalog

| Module          | Status    | Specification                                                                   | Purpose                                                            |
| --------------- | --------- | ------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| **config/**     | ✅ Stable | [Config Path API](crucible-go/standards/library/modules/config-path-api.md)     | XDG-compliant configuration path discovery and three-layer loading |
| **logging/**    | ✅ Stable | [Logging Standard](crucible-go/standards/observability/logging.md)              | Structured logging with progressive profiles (SIMPLE → ENTERPRISE) |
| **schema/**     | ✅ Stable | [Schema Validation](crucible-go/standards/library/modules/schema-validation.md) | JSON Schema validation with catalog and composition support        |
| **crucible/**   | ✅ Stable | [Crucible Shim](crucible-go/standards/library/modules/crucible-shim.md)         | Access to embedded Crucible schemas, docs, and standards           |
| **docscribe/**  | ✅ Stable | [Docscribe Module](crucible-go/standards/library/modules/docscribe.md)          | Frontmatter parsing, header extraction, and document processing    |
| **bootstrap/**  | ✅ Stable | [Bootstrap Pattern](crucible-go/standards/library/modules/fuldx-bootstrap.md)   | Dependency-free tool installation for Go repositories              |
| **pathfinder/** | ✅ Stable | [Pathfinder Extension](crucible-go/standards/library/extensions/pathfinder.md)  | Safe filesystem discovery with path traversal protection           |
| **ascii/**      | ✅ Stable | [ASCII Helpers](crucible-go/standards/library/extensions/ascii-helpers.md)      | Terminal utilities, Unicode width calculation, box drawing         |
| **foundry/**    | ✅ Stable | [Foundry Interfaces](crucible-go/standards/library/foundry/interfaces.md)       | Time, correlation IDs, patterns, MIME, HTTP status, country codes  |

**Legend**: ✅ Stable | 🚧 Planned | ⚠️ Experimental | 🔄 Refactoring

## Observability Highlights

### Progressive Logging Profiles

Gofulmen implements the [Fulmen Logging Standard](crucible-go/standards/observability/logging.md) with four progressive profiles:

| Profile        | Use Case             | Features                                      | Configuration                |
| -------------- | -------------------- | --------------------------------------------- | ---------------------------- |
| **SIMPLE**     | CLI tools, scripts   | Console output, basic severity                | Minimal (service name only)  |
| **STRUCTURED** | API services, jobs   | JSON output, correlation IDs, file sinks      | Service + sinks              |
| **ENTERPRISE** | Production workloads | Full envelope, middleware, throttling, policy | Service + sinks + middleware |
| **CUSTOM**     | Specialized adapters | Full control via custom config                | Service + customConfig       |

**Example Usage**:

```go
import "github.com/fulmenhq/gofulmen/logging"

// SIMPLE profile for CLI tools
logger := logging.New("mycli", logging.WithProfile(logging.ProfileSimple))
logger.Info("Starting operation")

// ENTERPRISE profile with policy enforcement
logger := logging.New("datawhirl",
    logging.WithProfile(logging.ProfileEnterprise),
    logging.WithPolicyFile("/org/logging-policy.yaml"))
logger.Info("Processing batch", logging.WithCorrelationID(correlationID))
```

### Policy Enforcement

Organizations can enforce logging standards via YAML policy files:

```yaml
# /org/logging-policy.yaml
allowedProfiles: [SIMPLE, STRUCTURED, ENTERPRISE]
requiredProfiles:
  workhorse: [ENTERPRISE]
  service: [STRUCTURED, ENTERPRISE]
  cli: [SIMPLE, STRUCTURED]
environmentRules:
  production: [STRUCTURED, ENTERPRISE]
  development: [SIMPLE, STRUCTURED, ENTERPRISE]
```

Policy files are resolved in order:

1. `.goneat/logging-policy.yaml` (repository-local)
2. `/etc/fulmen/logging-policy.yaml` (system-wide)
3. `/org/logging-policy.yaml` (organization-managed)

## Dependency Map

| Gofulmen Package | External Dependencies  | Crucible Assets                                                  | Notes                                        |
| ---------------- | ---------------------- | ---------------------------------------------------------------- | -------------------------------------------- |
| **config/**      | None (stdlib only)     | `schemas/config/fulmen-ecosystem/v1.0.0/`                        | XDG Base Directory compliant                 |
| **logging/**     | `uber-go/zap`          | `schemas/observability/logging/v1.0.0/`                          | Progressive profiles with policy enforcement |
| **schema/**      | `xeipuuv/gojsonschema` | `schemas/meta/draft-2020-12/`                                    | JSON Schema draft 2020-12 support            |
| **crucible/**    | None (embedded assets) | All synced assets in `docs/crucible-go/`, `schemas/crucible-go/` | Provides access to Crucible SSOT             |
| **bootstrap/**   | None (stdlib only)     | `schemas/tooling/external-tools/v1.0.0/`                         | Dependency-free tool installation            |
| **pathfinder/**  | None (stdlib only)     | `schemas/pathfinder/v1.0.0/`                                     | Safe filesystem discovery                    |
| **ascii/**       | None (stdlib only)     | `schemas/ascii/v1.0.0/`, `config/terminal/v1.0.0/`               | Terminal utilities and Unicode handling      |
| **foundry/**     | Cloud SDKs (optional)  | `schemas/library/foundry/v1.0.0/`, `config/library/foundry/`     | Enterprise data utilities (planned)          |

**Dependency Philosophy**: Minimize external dependencies; prefer standard library when possible. Cloud provider SDKs are optional and loaded only when needed.

## Roadmap & Gaps

### Current Version: 0.1.2

**Completed**:

- ✅ Core library modules (config, logging, schema, crucible, bootstrap)
- ✅ Extension modules (pathfinder, ascii, foundry)
- ✅ Goneat bootstrap integration
- ✅ Crucible SSOT synchronization
- ✅ Progressive logging with profiles and middleware pipeline
- ✅ Policy enforcement framework with YAML governance
- ✅ Schema validation with catalog and composition helpers
- ✅ Three-layer configuration loading
- ✅ Pathfinder security with path traversal protection
- ✅ Foundry utilities (time, correlation IDs, patterns, MIME, HTTP, country codes)

**Planned** (v0.1.3+):

- 📋 File checksums with xxHash128 (pathfinder enhancement)
- 📋 Additional coverage improvements for bootstrap package
- 📋 Performance optimizations
- 📋 Additional middleware (sampling, batching)

**Planned** (v0.2.0):

- 📋 Metrics integration (following logging pattern)
- 📋 Tracing integration (OpenTelemetry support)
- 📋 Cloud storage evaluation (pending cross-library discussion)
- 📋 Cosmography shim (when SSOT expands)
- 📋 Registry API clients (if SSOT repos expose HTTP endpoints)

### Migration Path

Applications currently using gofulmen v0.1.0 will have a clear upgrade path:

- **v0.1.0 → v0.2.0**: Additive changes only (new modules, enhanced features)
- **v0.2.0 → v1.0.0**: Potential breaking changes with migration guide
- **v1.0.0+**: Semantic versioning with backward compatibility guarantees

## Integration Examples

### Quick Start (CLI Tool)

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
    "github.com/fulmenhq/gofulmen/config"
)

func main() {
    // Simple logging for CLI
    logger := logging.New("mycli", logging.WithProfile(logging.ProfileSimple))
    logger.Info("Starting CLI tool")

    // XDG-compliant config paths
    configDir := config.GetAppConfigDir("mycli")
    logger.Info("Config directory", logging.WithField("path", configDir))
}
```

### Enterprise Service

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
    "github.com/fulmenhq/gofulmen/schema"
    "github.com/fulmenhq/gofulmen/crucible"
)

func main() {
    // Enterprise logging with policy enforcement
    logger := logging.New("datawhirl",
        logging.WithProfile(logging.ProfileEnterprise),
        logging.WithPolicyFile("/org/logging-policy.yaml"),
        logging.WithMiddleware(
            logging.CorrelationMiddleware(),
            logging.RedactSecretsMiddleware(),
            logging.ThrottlingMiddleware(1000, 100),
        ))

    // Schema validation using Crucible assets
    schemaData := crucible.GetPathfinderFindQuerySchema()
    validator, _ := schema.NewValidator(schemaData)

    logger.Info("Service started",
        logging.WithCorrelationID(correlationID),
        logging.WithContext(map[string]interface{}{
            "version": "2025.10.2",
            "region": "us-east-1",
        }))
}
```

## Testing & Quality

### Test Coverage

- **Target**: 80% minimum coverage across all packages
- **Current**: 85% average coverage
- **CI/CD**: All tests must pass before merge

### Quality Gates

```bash
make check-all  # Runs: sync + build + fmt + lint + test
make test       # Run all tests
make fmt        # Format code with goneat
make lint       # Run Go vet
```

### Continuous Integration

- GitHub Actions runs `make check-all` on every PR
- Release builds require `make release-check` to pass
- Version bumps validated via `make version-bump-{patch|minor|major}`

## Contributing

See [MAINTAINERS.md](../MAINTAINERS.md) for governance structure and [REPOSITORY_SAFETY_PROTOCOLS.md](../REPOSITORY_SAFETY_PROTOCOLS.md) for safety guidelines.

**Key Guidelines**:

- Follow [Go Coding Standards](crucible-go/standards/coding/go.md)
- Maintain backward compatibility (breaking changes require major version bump)
- Add tests for all new functionality
- Run `make check-all` before commits
- Document all exported APIs with godoc comments

## Resources

### Documentation

- [Integration Guide](INTEGRATION.md) - How to integrate gofulmen into your application
- [Goneat Guide](GONEAT.md) - Development tooling and workflows
- [Bootstrap Strategy](../ops/bootstrap-strategy.md) - Bootstrap architecture

### Standards & Specifications

- [Fulmen Helper Library Standard](crucible-go/architecture/fulmen-helper-library-standard.md)
- [Logging Standard](crucible-go/standards/observability/logging.md)
- [Config Path Standard](crucible-go/standards/config/fulmen-config-paths.md)
- [Makefile Standard](crucible-go/standards/makefile-standard.md)

### Package Documentation

- [Logging](../logging/README.md) - Structured logging with progressive profiles
- [Config](../config/README.md) - Configuration management, XDG paths, and three-layer loader (preview)
- [Schema](../schema/README.md) - JSON Schema validation, catalog discovery, composition/diff helpers, CLI shim (`gofulmen-schema`)
- [Crucible](../crucible/README.md) - Access to Crucible SSOT assets
- [Bootstrap](../bootstrap/README.md) - Tool installation and management
- [Pathfinder](../pathfinder/README.md) - Safe filesystem discovery
- [ASCII](../ascii/README.md) - Terminal utilities and Unicode handling

## Version Information

- **Current Version**: 0.1.2
- **Crucible Version**: 2025.10.2
- **Go Version**: 1.21+
- **License**: MIT

## Contact & Support

- **Maintainer**: @3leapsdave (Dave Thompson)
- **AI Co-Maintainer**: 🔧 Foundation Forge (@foundation-forge)
- **Issues**: [GitHub Issues](https://github.com/fulmenhq/gofulmen/issues)
- **Mattermost**: `#agents-gofulmen` (provisioning in progress)
