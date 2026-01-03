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
GONEAT_VERSION ?= v0.4.0
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

version-bump: ## Bump version (usage: make version-bump TYPE=patch|minor|major|calver)
	@if [ -z "$(TYPE)" ]; then \
		echo "❌ TYPE not specified. Usage: make version-bump TYPE=patch|minor|major|calver"; \
		exit 1; \
	fi
	@echo "Bumping version ($(TYPE))..."
	@$(GONEAT_RESOLVE); $$GONEAT version bump $(TYPE)
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
check-all: build fmt lint verify-hooks-compat test ## Run all quality checks (ensures sync, fmt, lint, hooks, test)
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
	@if [ -z "$(SCHEMA_ID)" ]; then \
		echo "❌ SCHEMA_ID not specified. Usage: make export-schema SCHEMA_ID=observability/logging/v1.0.0/log-event.schema.json OUT=output.json"; \
		exit 1; \
	fi
	@if [ -z "$(OUT)" ]; then \
		echo "❌ OUT not specified. Usage: make export-schema SCHEMA_ID=... OUT=output.json"; \
		exit 1; \
	fi
	@echo "Exporting schema $(SCHEMA_ID) to $(OUT)..."
	@go run ./cmd/gofulmen-export-schema --schema-id="$(SCHEMA_ID)" --out="$(OUT)" --no-validate
	@echo "✅ Schema exported successfully"

export-schema-example: ## Export example logging schema
	@echo "Exporting example logging schema..."
	@mkdir -p vendor/crucible/schemas
	@go run ./cmd/gofulmen-export-schema \
		--schema-id=observability/logging/v1.0.0/log-event.schema.json \
		--out=vendor/crucible/schemas/logging-event.schema.json \
		--no-validate
	@echo "✅ Example schema exported to vendor/crucible/schemas/logging-event.schema.json"
