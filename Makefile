# Gofulmen Makefile
# Compliant with FulmenHQ Makefile Standard
# Quick Start Commands:
#   make help           - Show all available commands
#   make bootstrap      - Install external tools (sfetch, goneat)
#   make test           - Run tests
#   make fmt            - Format code
#   make check-all      - Full quality check (fmt, test, coverage, license)

# Variables
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")

# Go related variables
GOCMD := go
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod

# Tool installation (user-space bin dir; overridable with BINDIR=...)
#
# Defaults:
# - macOS/Linux: $HOME/.local/bin
# - Windows (Git Bash / MSYS / MINGW / Cygwin): %USERPROFILE%\\bin (or $HOME/bin)
BINDIR ?=
BINDIR_RESOLVE = \
	BINDIR="$(BINDIR)"; \
	if [ -z "$$BINDIR" ]; then \
		OS_RAW="$$(uname -s 2>/dev/null || echo unknown)"; \
		case "$$OS_RAW" in \
			MINGW*|MSYS*|CYGWIN*) \
				if [ -n "$$USERPROFILE" ]; then \
					if command -v cygpath >/dev/null 2>&1; then \
						BINDIR="$$(cygpath -u "$$USERPROFILE")/bin"; \
					else \
						BINDIR="$$USERPROFILE/bin"; \
					fi; \
				elif [ -n "$$HOME" ]; then \
					BINDIR="$$HOME/bin"; \
				else \
					BINDIR="./bin"; \
				fi ;; \
			*) \
				if [ -n "$$HOME" ]; then \
					BINDIR="$$HOME/.local/bin"; \
				else \
					BINDIR="./bin"; \
				fi ;; \
		esac; \
	fi

# Tooling
GONEAT_VERSION ?= v0.5.1
SFETCH_INSTALL_URL ?= https://github.com/3leaps/sfetch/releases/latest/download/install-sfetch.sh

SFETCH_RESOLVE = \
	$(BINDIR_RESOLVE); \
	SFETCH=""; \
	if [ -x "$$BINDIR/sfetch" ]; then SFETCH="$$BINDIR/sfetch"; fi; \
	if [ -z "$$SFETCH" ]; then SFETCH="$$(command -v sfetch 2>/dev/null || true)"; fi

GONEAT_RESOLVE = \
	$(BINDIR_RESOLVE); \
	GONEAT=""; \
	if [ -x "$$BINDIR/goneat" ]; then GONEAT="$$BINDIR/goneat"; fi; \
	if [ -z "$$GONEAT" ]; then GONEAT="$$(command -v goneat 2>/dev/null || true)"; fi; \
	if [ -z "$$GONEAT" ]; then echo "❌ goneat not found. Run 'make bootstrap' first."; exit 1; fi

.PHONY: all help bootstrap bootstrap-force tools sync crucible-update version-bump lint test build build-all clean fmt version check-all precommit prepush
.PHONY: version-set version-bump-major version-bump-minor version-bump-patch release-check release-prepare release-build
.PHONY: release-tag release-verify-tag release-provenance-check release-guard-tag-version
.PHONY: test-coverage assess license-inventory license-save license-audit update-licenses dev export-schema export-schema-example verify-hooks-compat

# Default target
all: fmt test

# Bootstrap targets
bootstrap: ## Install external tools (sfetch, goneat + foundation tools)
	@echo "Installing external tools..."
	@$(BINDIR_RESOLVE); BINDIR="$$BINDIR" GONEAT_VERSION="$(GONEAT_VERSION)" SFETCH_INSTALL_URL="$(SFETCH_INSTALL_URL)" ./scripts/make-bootstrap.sh

bootstrap-force: ## Force reinstall external tools
	@$(MAKE) bootstrap FORCE=1

tools: ## Verify external tools are available
	@echo "Verifying external tools..."
	@$(SFETCH_RESOLVE); if [ -z "$$SFETCH" ]; then echo "❌ sfetch not found. Run 'make bootstrap' first."; exit 1; fi
	@$(GONEAT_RESOLVE); echo "✅ goneat: $$($$GONEAT --version 2>&1 | head -n1 || true)"
	@echo "✅ All tools verified"

sync: ## Sync assets from Crucible SSOT
	@echo "Syncing assets from Crucible..."
	@$(GONEAT_RESOLVE); $$GONEAT ssot sync
	@./scripts/sync-appidentity-schema.sh
	@echo "✅ Sync completed"

crucible-update: ## Update Crucible dependency to specific version (usage: make crucible-update VERSION=v0.2.19)
	@VERSION="$(VERSION)" ./scripts/crucible-update.sh

version-bump: ## Bump version (usage: make version-bump TYPE=patch|minor|major|calver)
	@test -n "$(TYPE)" || { echo "❌ TYPE not specified. Usage: make version-bump TYPE=patch|minor|major|calver"; exit 1; }
	@echo "Bumping version ($(TYPE))..."
	@$(GONEAT_RESOLVE); $$GONEAT version bump $(TYPE)
	@echo "✅ Version bumped to $$(cat VERSION)"

version-set: ## Set version to specific value (usage: make version-set VERSION=x.y.z)
	@test -n "$(VERSION)" || { echo "❌ VERSION not specified. Usage: make version-set VERSION=x.y.z"; exit 1; }
	@printf "%s\n" "$(VERSION)" > VERSION
	@echo "✅ Version set to $(VERSION)"

version-bump-major: ## Bump major version
	@$(MAKE) version-bump TYPE=major

version-bump-minor: ## Bump minor version
	@$(MAKE) version-bump TYPE=minor

version-bump-patch: ## Bump patch version
	@$(MAKE) version-bump TYPE=patch

release-check: ## Run release checklist validation
	@echo "Running release checklist..."
	@$(MAKE) check-all
	@echo "✅ Release check passed"

release-provenance-check: ## Verify Crucible SSOT provenance files exist
	@./scripts/release-provenance-check.sh

release-guard-tag-version: ## Guard: ensure tag matches VERSION (CI-friendly)
	@./scripts/release-guard-tag-version.sh

release-tag: ## Create and verify a signed git tag for VERSION
	@./scripts/release-tag.sh

release-verify-tag: ## Verify the signed git tag for VERSION
	@./scripts/release-verify-tag.sh

release-prepare: ## Prepare for release (sync, tests, version bump)
	@echo "Preparing release..."
	@$(MAKE) sync
	@$(MAKE) check-all
	@echo "✅ Release preparation complete"

release-build: build-all ## Build release artifacts (binaries + checksums)
	@echo "✅ Release build complete"

release-clean: ## Remove local release artifacts (dist/release)
	@echo "Cleaning release artifacts..."
	@rm -rf dist/release
	@echo "✅ Release artifacts removed"

# Help target
help: ## Show this help message
	@./scripts/make-help.sh

# Lint target (required by standard)
lint: ## Run lint checks
	@echo "Running Go vet..."
	@$(GOCMD) vet ./...
	@echo "Running golangci-lint..."
	@$(GONEAT_RESOLVE); $$GONEAT assess --categories lint
	@echo "✅ Lint checks passed"

# Build targets (required by standard)
build: sync ## Build distributable artifacts (ensures sync first)
	@echo "⚠️  Gofulmen is a library - no build artifacts to produce"
	@echo "✅ Build target satisfied (no-op, sync completed)"

build-all: ## Build multi-platform binaries and generate checksums
	@echo "⚠️  Gofulmen is a library - no cross-platform binaries to produce"
	@echo "✅ Build-all target satisfied (no-op)"

# Version target (required by standard)
version: ## Print current version
	@echo "$(VERSION)"

# Quality targets
check-all: build fmt lint verify-hooks-compat test license-audit ## Run all quality checks (ensures sync, fmt, lint, hooks, test, license)
	@echo "✅ All quality checks passed"

# Hook targets (required by standard)
verify-hooks-compat: ## Verify generated goneat hooks are compatible
	@./scripts/verify-hooks-compat.sh

precommit: ## Run pre-commit hooks
	@echo "Running pre-commit validation..."
	@$(GONEAT_RESOLVE); $$GONEAT assess --hook pre-commit --hook-manifest .goneat/hooks.yaml --package-mode
	@echo "✅ Pre-commit checks passed"

prepush: ## Run pre-push hooks
	@echo "Running pre-push validation..."
	@$(GONEAT_RESOLVE); $$GONEAT assess --hook pre-push --hook-manifest .goneat/hooks.yaml --package-mode
	@echo "✅ Pre-push checks passed"

# Test targets
test: ## Run all tests
	@echo "Running test suite..."
	$(GOTEST) ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format targets
fmt: ## Format code with goneat (requires bootstrap)
	@echo "Formatting with goneat..."
	@$(GONEAT_RESOLVE); $$GONEAT format
	@echo "✅ Formatting completed"

assess: ## Run goneat assess (requires bootstrap)
	@echo "Running goneat assess..."
	@$(GONEAT_RESOLVE); $$GONEAT assess

# License compliance
license-inventory: ## Generate CSV inventory of dependency licenses
	@./scripts/license.sh inventory

license-save: ## Save third-party license texts
	@./scripts/license.sh save

license-audit: ## Audit for forbidden licenses
	@./scripts/license.sh audit

update-licenses: license-inventory license-save ## Update license inventory and texts

# Clean targets
clean: ## Clean build artifacts and reports
	@echo "Cleaning artifacts..."
	rm -rf bin dist coverage.out coverage.html vendor
	@echo "✅ Clean completed"

# Development setup
dev: ## Set up development environment
	@echo "Setting up development environment..."
	$(MAKE) fmt
	$(MAKE) test
	@echo "✅ Development environment ready"

# Schema export targets
export-schema: ## Export a schema (usage: make export-schema SCHEMA_ID=... OUT=...)
	@SCHEMA_ID="$(SCHEMA_ID)" OUT="$(OUT)" ./scripts/export-schema.sh export

export-schema-example: ## Export example logging schema
	@./scripts/export-schema.sh example
