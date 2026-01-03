# Gofulmen – Maintainers

**Project**: gofulmen  
**Purpose**: Go foundation library for FulmenHQ ecosystem – Enterprise-grade packages for configuration, logging, schema validation, and developer tooling  
**Governance Model**: 3leaps Initiative

## Human Maintainers

### @3leapsdave (Dave Thompson)

- **Role**: Project Lead & Primary Maintainer
- **Responsibilities**: Architecture oversight, release management, API governance, library quality assurance
- **Contact**: dave.thompson@3leaps.net | GitHub [@3leapsdave](https://github.com/3leapsdave) | X [@3leapsdave](https://x.com/3leapsdave)
- **Supervision**: All AI agent contributions

## Agentic Roles (v0.3.0)

Gofulmen uses role-based prompts and attribution per Crucible v0.3.0.

- Role catalog: `config/crucible-go/agentic/roles/`
- Attribution baseline: `docs/crucible-go/standards/agentic-attribution.md`

### Recommended roles for gofulmen

- `devlead` (implementation)
- `devrev` (review)
- `infoarch` (documentation/standards)
- `secrev` (security review for secrets/crypto/supply chain)

## Agent Attribution Guidelines

- Follow the [Agentic Attribution Standard](docs/crucible-go/standards/agentic-attribution.md).
- Commits should use role-based attribution (model + interface + role).
- Always use the canonical `noreply@3leaps.net` email for `Co-Authored-By:`.
- Supervised mode requires `Committer-of-Record:` (human accountability).

## Governance Structure

- Human maintainers approve architecture, releases, and supervise AI agents.
- AI agents execute tasks, maintain code quality, and uphold library standards under supervision.
- See `REPOSITORY_SAFETY_PROTOCOLS.md` for guardrails and escalation paths.

## Communication Channels

- **Primary**: GitHub Issues and Pull Requests
- **Real-time**: Mattermost `#agents-gofulmen` (provisioning in progress)
- **Escalation**: Direct contact with @3leapsdave for critical issues

## Contribution Guidelines

All contributors (human and AI) must:

- Follow Go coding standards from `docs/crucible-go/standards/coding/go.md`
- Maintain test coverage above 80%
- Run `make check-all` before commits
- Document all exported APIs with godoc comments
- Maintain backward compatibility for library consumers
- Coordinate breaking changes with @3leapsdave
