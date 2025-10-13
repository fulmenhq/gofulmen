# Gofulmen Development Operations

> **Location**: `docs/development/operations.md` (standardized across all Fulmen helper libraries)

## 🎯 Mission

Enable developers to build enterprise-grade Go applications using Gofulmen library with comprehensive support, clear documentation, and reliable tooling.

This document provides operational guidance for Gofulmen maintainers and contributors.

## 🛠️ Development Workflow

### Getting Started

```bash
# Clone and setup
git clone https://github.com/fulmenhq/gofulmen
cd gofulmen
make bootstrap

# Start development
make sync
make test
```

### Daily Development

```bash
# Standard development cycle
make fmt             # Format code with goneat
make lint            # Run Go vet
make test            # Run all tests
make test-coverage   # With coverage report
```

### Quality Assurance

```bash
# Pre-commit checks
make fmt lint test

# Full quality suite
make check-all       # All quality checks (sync + build + fmt + lint + test)
make precommit       # Pre-commit hooks
make prepush         # Pre-push hooks
```

## 🚀 Release Process

### Version Management

- **Semantic Versioning**: Follow MAJOR.MINOR.PATCH for API changes
- **Changelog Maintenance**: Document all changes in CHANGELOG.md
- **Tagging**: Use Git tags with signed releases
- **GitHub Releases**: Automated with comprehensive release notes

### Release Checklist

```bash
# Complete release preparation
make release-check   # Verify all requirements
make release-prepare # Update docs and sync
make release-build   # Build distribution
```

### Version Bumping

```bash
# Bump patch version (0.1.0 → 0.1.1)
make version-bump-patch

# Bump minor version (0.1.0 → 0.2.0)
make version-bump-minor

# Bump major version (0.1.0 → 1.0.0)
make version-bump-major

# Set specific version
make version-set VERSION=1.2.3
```

## 🔧 Tooling and Commands

### Development Tools

- **Bootstrap**: `make bootstrap` - Install goneat and verify tools
- **Sync**: `make sync` - Sync assets from Crucible SSOT
- **Testing**: `make test` - Run test suite
- **Quality**: `make lint` - Code quality checks
- **Formatting**: `make fmt` - Format code with goneat

### Make Targets Reference

Run `make help` to see all available targets:

- **`make help`**: Show all available targets
- **`make bootstrap`**: Install external tools
- **`make tools`**: Verify tools are available
- **`make sync`**: Sync Crucible assets
- **`make test`**: Run all tests
- **`make fmt`**: Format code
- **`make lint`**: Run linting
- **`make clean`**: Remove build artifacts
- **`make version`**: Print current version
- **`make check-all`**: Run all quality checks

## 🧪 Testing Strategy

### Test Coverage

- **Unit Tests**: 80%+ coverage on public API
- **Integration Tests**: Cross-module functionality
- **Package Tests**: Each package has comprehensive test suite
- **Compatibility Tests**: Go version matrix testing (1.21+)

### Quality Gates

- **Code Style**: Go fmt and Go vet
- **Type Safety**: Go compiler type checking
- **Security**: License audit and dependency scanning
- **Documentation**: Godoc coverage for all exported APIs

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test ./logging/... -v

# Run specific test
go test ./logging/... -run TestNew -v
```

## 📊 Monitoring and Analytics

### Development Metrics

- **Test Coverage**: Track coverage trends over time
- **Performance**: Monitor benchmark performance
- **Quality**: Track lint and vet results
- **Dependencies**: Monitor for security updates

### Release Analytics

- **Go Module Stats**: Track module downloads
- **Usage Analytics**: Error reports and telemetry (opt-in)
- **Community Engagement**: GitHub stars, issues, PRs

## 🤝 Community Guidelines

### Contribution Process

1. **Fork Repository**: Create personal fork for development
2. **Create Branch**: Use descriptive branch names (feature/_, fix/_, docs/\*)
3. **Make Changes**: Implement with tests and documentation
4. **Submit PR**: Pull request with comprehensive description
5. **Code Review**: Address feedback from maintainers
6. **Merge**: Maintainers merge after approval

### Code Standards

- **Go Coding Standards**: Follow [docs/crucible-go/standards/coding/go.md](../crucible-go/standards/coding/go.md)
- **Godoc Comments**: Comprehensive documentation for all exported APIs
- **Error Handling**: Proper error wrapping and context
- **Testing**: Unit tests for all functionality
- **Backward Compatibility**: Maintain API stability for library consumers

### Support Channels

- **GitHub Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Mattermost**: `#agents-gofulmen` for real-time discussion (provisioning in progress)
- **Email**: dave.thompson@3leaps.net for private issues

## 🔐 Security

### Security Process

1. **Vulnerability Reporting**: Private disclosure to maintainers
2. **Security Reviews**: Regular dependency scanning
3. **Patch Management**: Prioritized security updates
4. **Security Documentation**: Security considerations and best practices

### Dependency Management

```bash
# Audit dependencies
make license-audit

# Update dependencies
go get -u ./...
go mod tidy

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

### Security Best Practices

- **Minimal Dependencies**: Prefer standard library when possible
- **Vetted Dependencies**: All new dependencies require approval
- **License Compliance**: Verify all dependency licenses
- **Regular Updates**: Keep dependencies current

## 🏗️ Architecture Guidelines

### Package Organization

- **One Concern Per Package**: Each package has a single, well-defined purpose
- **Minimal Coupling**: Packages should be independently usable
- **Clear Interfaces**: Export minimal, well-documented APIs
- **Internal Packages**: Use internal/ for implementation details

### API Design Principles

- **Progressive Complexity**: Simple use cases require minimal configuration
- **Sensible Defaults**: Provide good defaults for common scenarios
- **Explicit Configuration**: Allow full control when needed
- **Backward Compatibility**: Maintain API stability across minor versions

### Testing Philosophy

- **Test Public APIs**: Focus on exported functionality
- **Table-Driven Tests**: Use table-driven tests for comprehensive coverage
- **Example Tests**: Provide runnable examples in godoc
- **Benchmark Tests**: Include benchmarks for performance-critical code

## 📖 Documentation Standards

### Godoc Requirements

- **Package Comments**: Every package must have a package-level comment
- **Function Comments**: All exported functions must have godoc comments
- **Example Code**: Provide runnable examples for key functionality
- **Links**: Reference related packages and Crucible standards

### Documentation Updates

- **API Changes**: Update godoc comments when changing APIs
- **README Updates**: Keep README.md current with new features
- **Changelog**: Document all changes in CHANGELOG.md
- **Migration Guides**: Provide migration guides for breaking changes

## 🔄 Continuous Integration

### CI/CD Pipeline

- **GitHub Actions**: Automated testing on push and PR
- **Quality Gates**: All checks must pass before merge
- **Multi-Platform**: Test on Linux, macOS, Windows
- **Go Versions**: Test on Go 1.21, 1.22, 1.23

### CI Commands

```bash
# What CI runs
make bootstrap
make sync
make check-all
make test-coverage
make license-audit
```

---

_This documentation supports Gofulmen's mission to enable enterprise-grade Go development._
