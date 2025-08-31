# Makefile

# Binary name
BINARY := envoy

# Go tooling
GORUN	:= go run
GOTEST := go test
GOBUILD := go build

# Directories
PKGS := ./...
CMD_DIR := ./cmd

# Air (live reload dev tool). Install: https://github.com/cosmtrek/air
AIR := air

.PHONY: run dev test test-ci build start help

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

# Start the compiled binary (production)
start: build
		@echo "Starting $(BINARY)..."
		./$(BINARY)

# Help
help:
	@echo "Makefile targets:"
	@echo "  make run          - run the program"
	@echo "  make dev          - run with air (live reload)"
	@echo "  make test         - run all tests (go test ./...)"
	@echo "  make test-ci      - run all tests with race detector and coverage"
	@echo "  make start        - build and run the binary"
