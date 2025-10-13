# Gofulmen Makefile
# Compliant with FulmenHQ Makefile Standard
# Quick Start Commands:
#   make help           - Show all available commands
#   make bootstrap      - Install external tools (goneat)
#   make test           - Run tests
#   make fmt            - Format code
#   make check-all      - Full quality check (fmt, test, coverage, license)

# Variables
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")
GONEAT := ./bin/goneat

# Go related variables
GOCMD := go
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod

.PHONY: help bootstrap bootstrap-force tools sync version-bump lint test build build-all clean fmt version check-all precommit prepush
.PHONY: version-set version-bump-major version-bump-minor version-bump-patch release-check release-prepare release-build
.PHONY: test-coverage assess license-inventory license-save license-audit update-licenses dev

# Default target
all: fmt test

# Bootstrap targets
bootstrap: ## Install external tools (goneat)
	@echo "Installing external tools..."
	@if [ "$(FORCE)" = "1" ] || [ "$(FORCE)" = "true" ]; then \
		go run ./cmd/bootstrap --install --verbose --force; \
	else \
		go run ./cmd/bootstrap --install --verbose; \
	fi
	@echo "✅ Bootstrap completed. Use './bin/goneat' or add ./bin to PATH"

bootstrap-force: ## Force reinstall external tools
	@$(MAKE) bootstrap FORCE=1

tools: ## Verify external tools are available
	@go run ./cmd/bootstrap --verify --verbose

sync: ## Sync assets from Crucible SSOT
	@if [ ! -f $(GONEAT) ]; then \
		echo "❌ goneat not found. Run 'make bootstrap' first."; \
		exit 1; \
	fi
	@echo "Syncing assets from Crucible..."
	@$(GONEAT) ssot sync
	@$(MAKE) sync-foundry-assets
	@echo "✅ Sync completed"

sync-foundry-assets: ## Copy foundry YAML assets to embedded location (post-sync hook)
	@echo "Copying foundry assets for embed..."
	@mkdir -p foundry/assets
	@cp config/crucible-go/library/foundry/*.yaml foundry/assets/
	@echo "✅ Foundry assets synchronized"

version-bump: ## Bump version (usage: make version-bump TYPE=patch|minor|major|calver)
	@if [ ! -f $(GONEAT) ]; then \
		echo "❌ goneat not found. Run 'make bootstrap' first."; \
		exit 1; \
	fi
	@if [ -z "$(TYPE)" ]; then \
		echo "❌ TYPE not specified. Usage: make version-bump TYPE=patch|minor|major|calver"; \
		exit 1; \
	fi
	@echo "Bumping version ($(TYPE))..."
	@$(GONEAT) version bump --type $(TYPE)
	@echo "✅ Version bumped to $$(cat VERSION)"

version-set: ## Set version to specific value (usage: make version-set VERSION=x.y.z)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION not specified. Usage: make version:set VERSION=x.y.z"; \
		exit 1; \
	fi
	@echo "$(VERSION)" > VERSION
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

release-prepare: ## Prepare for release (sync, tests, version bump)
	@echo "Preparing release..."
	@$(MAKE) sync
	@$(MAKE) check-all
	@echo "✅ Release preparation complete"

release-build: build-all ## Build release artifacts (binaries + checksums)
	@echo "✅ Release build complete"

# Help target
help: ## Show this help message
	@echo "Gofulmen - Available Make Targets"
	@echo ""
	@echo "Required targets (Makefile Standard):"
	@echo "  help            - Show this help message"
	@echo "Required targets (Makefile Standard):"
	@echo "  help            - Show this help message"
	@echo "  bootstrap       - Install external tools from .goneat/tools.yaml"
	@echo "  bootstrap-force - Force reinstall external tools"
	@echo "  tools           - Verify external tools are available"
	@echo "  lint            - Run lint/format/style checks"
	@echo "  test            - Run all tests"
	@echo "  build           - Build distributable artifacts (no-op for libraries)"
	@echo "  build-all       - Build multi-platform binaries (no-op for libraries)"
	@echo "  clean           - Remove build artifacts and caches"
	@echo "  fmt             - Format code"
	@echo "  version         - Print current version"
	@echo "  version-set     - Set version to specific value"
	@echo "  version-bump-major - Bump major version"
	@echo "  version-bump-minor - Bump minor version"
	@echo "  version-bump-patch - Bump patch version"
	@echo "  release-check   - Run release checklist validation"
	@echo "  release-prepare - Prepare for release"
	@echo "  release-build   - Build release artifacts"
	@echo "  check-all       - Run all quality checks (sync, fmt, lint, test)"
	@echo "  precommit       - Run pre-commit hooks (check-all)"
	@echo "  prepush         - Run pre-push hooks (check-all)"
	@echo ""
	@echo "Goneat targets:"
	@echo "  sync            - Sync assets from Crucible SSOT"
	@echo "  version-bump    - Bump version (usage: make version-bump TYPE=patch|minor|major|calver)"
	@echo ""
	@echo "Additional targets:"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  assess          - Run goneat assessment (requires bootstrap)"
	@echo "  license-audit   - Audit dependency licenses"
	@echo ""

# Lint target (required by standard)
lint: ## Run lint/format/style checks
	@echo "Running Go vet..."
	@$(GOCMD) vet ./...
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
check-all: build fmt lint test ## Run all quality checks (ensures sync, fmt, lint, test)
	@echo "✅ All quality checks passed"

# Hook targets (required by standard)
precommit: check-all ## Run pre-commit hooks (check-all includes sync, lint, test)
	@echo "✅ Pre-commit checks passed"

prepush: check-all ## Run pre-push hooks (check-all includes sync, lint, test)
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
	@if [ ! -f ./bin/goneat ]; then \
		echo "❌ goneat not found. Run 'make bootstrap' first."; \
		exit 1; \
	fi
	@echo "Formatting with goneat..."
	@./bin/goneat format
	@echo "✅ Formatting completed"

assess: ## Run goneat assess (requires bootstrap)
	@if [ ! -f ./bin/goneat ]; then \
		echo "goneat not found. Run 'make bootstrap' first."; \
		exit 1; \
	fi
	@echo "Running goneat assess..."
	@./bin/goneat assess

# License compliance
license-inventory: ## Generate CSV inventory of dependency licenses
	@echo "🔎 Generating license inventory (CSV)..."
	@mkdir -p docs/licenses dist/reports
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		echo "Installing go-licenses..."; \
		go install github.com/google/go-licenses@latest; \
	fi
	go-licenses csv ./... > docs/licenses/inventory.csv
	@echo "✅ Wrote docs/licenses/inventory.csv"

license-save: ## Save third-party license texts
	@echo "📄 Saving third-party license texts..."
	@rm -rf docs/licenses/third-party
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		echo "Installing go-licenses..."; \
		go install github.com/google/go-licenses@latest; \
	fi
	go-licenses save ./... --save_path=docs/licenses/third-party
	@echo "✅ Saved third-party licenses to docs/licenses/third-party"

license-audit: ## Audit for forbidden licenses
	@echo "🧪 Auditing dependency licenses..."
	@mkdir -p dist/reports
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		echo "Installing go-licenses..."; \
		go install github.com/google/go-licenses@latest; \
	fi
	forbidden='GPL|LGPL|AGPL|MPL|CDDL'; \
	out=$$(go-licenses csv ./...); \
	echo "$$out" > dist/reports/license-inventory.csv; \
	if echo "$$out" | grep -E "$$forbidden" >/dev/null; then \
		echo "❌ Forbidden license detected. See dist/reports/license-inventory.csv"; \
		exit 1; \
	else \
		echo "✅ No forbidden licenses detected"; \
	fi

update-licenses: license-inventory license-save ## Update license inventory and texts

# Clean targets
clean: ## Clean build artifacts and reports
	@echo "Cleaning artifacts..."
	rm -rf dist coverage.out coverage.html
	@echo "✅ Clean completed"

# Development setup
dev: ## Set up development environment
	@echo "Setting up development environment..."
	$(MAKE) fmt
	$(MAKE) test
	@echo "✅ Development environment ready"