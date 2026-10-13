# Envoy
A secure environment variable management tool built using Rust. Project consists of a REST API and a CLI.

## Core Features:

- Variable Creation: Securely store environment variables.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.

## Installation

Build from source (binaries in `target/release/`):
```bash
cargo build --release
```

Run the server:
```bash
JWT_SECRET=... TURSO_DATABASE_URL=... TURSO_AUTH_TOKEN=... ./target/release/envoy-server
```

Run the CLI:
```bash
./target/release/envoy --help
```
