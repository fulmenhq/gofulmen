#!/usr/bin/env bash
set -euo pipefail

cat <<'EOF'
Gofulmen - Available Make Targets

Required targets (Makefile Standard):
  help            - Show this help message
  bootstrap       - Install external tools (sfetch, goneat)
  bootstrap-force - Force reinstall external tools
  tools           - Verify external tools are available
  lint            - Run lint/format/style checks
  test            - Run all tests
  build           - Build distributable artifacts (no-op for libraries)
  build-all       - Build multi-platform binaries (no-op for libraries)
  clean           - Remove build artifacts and caches
  fmt             - Format code
  version         - Print current version
  version-set     - Set version to specific value
  version-bump-major - Bump major version
  version-bump-minor - Bump minor version
  version-bump-patch - Bump patch version
  release-check   - Run release checklist validation
  release-prepare - Prepare for release
  release-build   - Build release artifacts
  release-clean   - Remove local release artifacts
  release-provenance-check - Verify SSOT provenance files
  release-guard-tag-version - Guard tag matches VERSION
  release-tag     - Create signed git tag for VERSION
  release-verify-tag - Verify signed git tag for VERSION
  check-all       - Run all quality checks (sync, fmt, lint, test)
  precommit       - Run pre-commit hooks (check-all)
  prepush         - Run pre-push hooks (check-all)

Goneat targets:
  sync            - Sync assets from Crucible SSOT
  version-bump    - Bump version (usage: make version-bump TYPE=patch|minor|major|calver)

Additional targets:
  test-coverage   - Run tests with coverage report
  assess          - Run goneat assessment (requires bootstrap)
  license-audit   - Audit dependency licenses
EOF
