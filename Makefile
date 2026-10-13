# Makefile (Rust)

SERVER_BIN := envoy-server
CLI_BIN := envoy

CARGO := cargo

.PHONY: dev dev-server dev-cli test test-ci build build-release lint fmt clean help

# Run the server (expects JWT_SECRET + TURSO_DATABASE_URL in env/.env)
dev-server dev:
	$(CARGO) run -p envoy-server

# Run the CLI (expects ENVOY_SERVER_URL for local dev)
dev-cli:
	ENVOY_SERVER_URL=http://localhost:8080 $(CARGO) run -p envoy-cli -- $(ARGS)

# Run the full test suite
test:
	$(CARGO) test --workspace

# CI-friendly test with warnings denied
test-ci:
	$(CARGO) test --workspace --locked
	$(CARGO) clippy --workspace --no-deps -- -D warnings

# Debug builds
build:
	$(CARGO) build --workspace

# Optimised builds (binaries in target/release/)
build-release:
	$(CARGO) build --workspace --release

lint:
	$(CARGO) clippy --workspace --no-deps

fmt:
	$(CARGO) fmt --all

clean:
	$(CARGO) clean

help:
	@echo "Makefile commands:"
	@echo "  make dev/dev-server - run the API server"
	@echo "  make dev-cli        - run the CLI (pass ARGS=\"... =\")"
	@echo "  make test           - run all tests (cargo test)"
	@echo "  make test-ci        - tests + clippy with warnings denied"
	@echo "  make build          - debug build all crates"
	@echo "  make build-release  - release build all crates"
	@echo "  make lint           - clippy"
	@echo "  make fmt            - rustfmt"
	@echo "  make clean          - cargo clean"
