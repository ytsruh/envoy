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

.PHONY: run dev test test-ci build start build-cli run-cli help

# Run the program
run:
	$(GORUN) $(CMD_DIR)

# Development run with air (live reload). Requires 'air' installed.
dev:
	@command -v $(AIR) >/dev/null 2>&1 || (echo "air not found;"; exit 1)
	$(AIR)

# Run the full test suite
test:
	$(GOTEST) -v $(PKGS)

# CI-friendly test with race detector and coverage
test-ci:
	$(GOTEST) -race -coverprofile=coverage.out $(PKGS)

# Build binary
build:
	$(GOBUILD) -o $(BINARY) $(CMD_DIR)
	$(GOBUILD) -tags=cli -o $(CLI_BINARY) $(CMD_DIR)/cli.go

# Start the compiled binary (production)
start: build
		@echo "Starting $(BINARY)..."
		./$(BINARY)

# Build CLI binary
build-cli:
	$(GOBUILD) -tags=cli -o $(CLI_BINARY) $(CMD_DIR)/cli.go

# Run the CLI (development)
run-cli:
	$(GORUN) -tags=cli $(CMD_DIR)/cli.go

# Generate SQL Queries
generate:
	@command sqlc generate

# Help
help:
	@echo "Makefile commands:"
	@echo "  make run          - run the program"
	@echo "  make dev          - run with air (live reload)"
	@echo "  make test         - run all tests (go test ./...)"
	@echo "  make test-ci      - run all tests with race detector and coverage"
	@echo "  make start        - build and run the binary"
	@echo "  make build-cli    - build the CLI binary"
	@echo "  make run-cli      - run the CLI (development)"
	@echo "  make generate     - generate SQL queries"
