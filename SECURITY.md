# Security Policy

## Overview

3 Leaps, LLC is committed to ensuring the security of our open-source projects and supported ecosystems (e.g., fulmenhq, mdmeld, docemist). We appreciate the community's help in responsibly disclosing vulnerabilities to protect users.

All reports and handling must align with our [Code of Conduct](https://github.com/3leaps/oss-policies/blob/main/CODE-OF-CONDUCT.md).

## Supported Versions

Security updates are provided for:

- **Latest stable release**: Current production-ready version
- **Alpha releases**: Best-effort support during active development

**Current Status**: gofulmen v0.3.x is the reference implementation. We provide security patches for the latest v0.3.x release.

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :x:                |
| < 0.2   | :x:                |

## Reporting a Vulnerability

If you discover a potential security vulnerability, please report it privately—do not disclose it publicly (e.g., via issues or forums) until we've had a chance to address it.

### How to Report

- **Preferred Method**: Email security@3leaps.net with details, including:
  - Description of the vulnerability
  - Steps to reproduce (e.g., affected version, configuration)
  - Potential impact (e.g., data exposure, denial of service, privilege escalation)
  - Any proposed fixes or patches
- **Alternative**: Use GitHub Security Advisories in this repository (if enabled) for private reporting
- **Encryption**: If sensitive, encrypt your report using our public PGP key (available upon request)

We prioritize confidentiality and will acknowledge your report within 3 business days.

## Vulnerability Handling Process

1. **Acknowledgment**: We'll confirm receipt and provide an initial assessment within 3 business days.
2. **Triage and Validation**: Our team will investigate and validate the issue, typically within 7 days.
3. **Fix Development**: If confirmed, we'll develop a fix. Timeline depends on severity but aims for resolution within 30 days for critical issues.
4. **Coordinated Disclosure**: We'll work with you on a disclosure plan. Vulnerabilities are publicly disclosed after a fix is released, or no later than 90 days from report (whichever comes first), unless mutually agreed otherwise.
5. **Credit**: Reporters are credited in advisories (with your permission) for responsible disclosures.

## Scope

This policy applies to:

- gofulmen library code (`github.com/fulmenhq/gofulmen` module)
- CLI tools bundled with gofulmen
- Documentation examples that could lead to insecure implementations

Out of scope:

- Theoretical vulnerabilities without practical exploit path
- Vulnerabilities in dependencies (report to upstream, but notify us if affecting gofulmen)
- Issues requiring physical access to user systems

## Safe Harbor

If you follow this policy in good faith (e.g., no exploitation beyond proof-of-concept), we will not pursue legal action against you. We consider this ethical security research.

## Security Best Practices for gofulmen Users

When using gofulmen in your applications:

- **Input Validation**: Always validate user input before passing to schema validation or file operations
- **Path Traversal**: Use Pathfinder's boundary enforcement for security-sensitive filesystem operations
- **Secrets Management**: Never log secrets with gofulmen's logging module - use redaction middleware
- **Dependencies**: Keep gofulmen and its dependencies up to date via `go get -u`
- **Configuration**: Validate configuration files with schema validation before use
- **Context Usage**: Always pass contexts with appropriate timeouts to prevent resource exhaustion

## Dependency Audit

To audit gofulmen's dependencies:

```bash
# View dependency graph
go mod graph

# Check for known vulnerabilities
govulncheck ./...

# Generate SBOM (requires cyclonedx-gomod)
cyclonedx-gomod mod -output sbom.json
```

## Questions

For questions about this policy, contact security@3leaps.net or open a non-security issue in this repository.

For additional governance details and contributor obligations, see the [3 Leaps Open Source Policies](https://github.com/3leaps/oss-policies).

---

_This policy is subject to change. Last updated: 2026-01-08._
