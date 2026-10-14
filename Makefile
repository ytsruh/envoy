# Makefile (Rust)

SERVER_BIN := envoy-server
CLI_BIN := envoy

CARGO := cargo

.PHONY: server test test-ci build build-release lint fmt clean help website-serve website-build

# Run the server (zero-config locally: SQLite at ./data/envoy.db, JWT secret auto-generated)
server:
	$(CARGO) run -p envoy-server

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
	rm -rf website/public

website-serve:
	zola --root website serve

website-build:
	zola --root website build

help:
	@echo "Makefile commands:"
	@echo "  make server         - run the API server (cargo run -p envoy-server)"
	@echo "  make test           - run all tests (cargo test)"
	@echo "  make test-ci        - tests + clippy with warnings denied"
	@echo "  make build          - debug build (binaries in target/debug/: envoy, envoy-server)"
	@echo "  make build-release  - release build (binaries in target/release/: envoy, envoy-server)"
	@echo "  make lint           - clippy"
	@echo "  make fmt            - rustfmt"
	@echo "  make clean          - cargo clean"
	@echo "  make website-serve  - run Zola dev server for website/"
	@echo "  make website-build  - build static site to website/public/"
