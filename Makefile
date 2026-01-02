# Makefile

# Binary names
BINARY := envoy-server
CLI_BINARY := envoy-cli

# Go tooling
GORUN	:= go run
GOTEST := go test
GOBUILD := go build

# Directories
PKGS := ./...
CMD_DIR := ./cmd
CLI_DIR := ./cmd

# Air (live reload dev tool). Install: https://github.com/cosmtrek/air
AIR := air

.PHONY: dev dev-server dev-cli test test-ci build-dev build generate bump-patch bump-minor bump-major help

# Development run with air (live reload). Requires 'air' installed.
# Sets ENVOY_SERVER_URL to localhost for local development
dev-server:
	@which $(AIR) >/dev/null 2>&1 || { echo "air not found"; exit 1; }
	@export ENVOY_SERVER_URL=http://localhost:8080 && \
	$(AIR)

# Alias for dev-server
dev: dev-server

# Run the CLI (development - uses localhost)
dev-cli:
	@export ENVOY_SERVER_URL=http://localhost:8080 && \
	$(GORUN) -tags=cli $(CMD_DIR)/cli.go

# Run the full test suite
test:
	$(GOTEST) -v $(PKGS)

# CI-friendly test with race detector and coverage
test-ci:
	$(GOTEST) -race -coverprofile=coverage.out $(PKGS)

# Build binary (development URL is overridden by export for this build only)
build-dev:
	@export ENVOY_SERVER_URL=http://localhost:8080 && \
	$(GOBUILD) -o $(BINARY) $(CMD_DIR)/server.go && \
	$(GOBUILD) -tags=cli -o $(CLI_BINARY) $(CMD_DIR)/cli.go

# Build binary (production URL is hardcoded in code)
build:
	$(GOBUILD) -o $(BINARY) $(CMD_DIR)/server.go
	$(GOBUILD) -tags=cli -o $(CLI_BINARY) $(CMD_DIR)/cli.go

# Generate SQL Queries
generate:
	sqlc generate

# Version bumping scripts
bump-patch:
	@./version.sh patch

bump-minor:
	@./version.sh minor

bump-major:
	@./version.sh major

# Help
help:
	@echo "Makefile commands:"
	@echo "  make dev/dev-server - run with air (live reload)"
	@echo "  make dev-cli       - run CLI (development)"
	@echo "  make test          - run all tests (go test ./...)"
	@echo "  make test-ci       - run all tests with race detector and coverage"
	@echo "  make build-dev     - build binaries for development"
	@echo "  make build         - build binaries for production"
	@echo "  make generate      - generate SQL queries"
	@echo "  make bump-patch    - bump patch version (v0.0.0 -> v0.0.1)"
	@echo "  make bump-minor    - bump minor version (v0.0.0 -> v0.1.0)"
	@echo "  make bump-major    - bump major version (v0.0.0 -> v1.0.0)"
