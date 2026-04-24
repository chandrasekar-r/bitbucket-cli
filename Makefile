.PHONY: help build test lint install clean completions release-dry-run

# Default target — show available commands
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Build ─────────────────────────────────────────────────────────────────────

build: ## Build bb binary (output: ./bin/bb)
	@mkdir -p bin
	CGO_ENABLED=0 go build \
		-ldflags "-s -w \
			-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Version=$(shell git describe --tags --always 2>/dev/null || echo dev) \
			-X github.com/chandrasekar-r/bitbucket-cli/internal/version.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) \
			-X github.com/chandrasekar-r/bitbucket-cli/internal/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
			-X github.com/chandrasekar-r/bitbucket-cli/internal/version.OAuthClientID=$(BB_OAUTH_CLIENT_ID) \
			-X github.com/chandrasekar-r/bitbucket-cli/internal/version.OAuthClientSecret=$(BB_OAUTH_CLIENT_SECRET)" \
		-o bin/bb ./cmd/bb
	@echo "Built: bin/bb"

install: build ## Install bb to /usr/local/bin
	install -m 755 bin/bb /usr/local/bin/bb
	@echo "Installed: /usr/local/bin/bb"

# ── Test ──────────────────────────────────────────────────────────────────────

test: ## Run all tests
	go test -race -count=1 ./...

test-verbose: ## Run tests with verbose output
	go test -race -count=1 -v ./...

# ── Code quality ──────────────────────────────────────────────────────────────

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

# ── Shell completions ─────────────────────────────────────────────────────────

completions: build ## Generate shell completion scripts into ./completions/
	@bash scripts/completions.sh

# ── Release ───────────────────────────────────────────────────────────────────

release-dry-run: ## Build all release artifacts locally (no publish)
	goreleaser release --snapshot --clean

# ── Cleanup ───────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf bin/ dist/ completions/
	@echo "Cleaned"
