# Envoy
A secure environment variable management tool built using Rust. Self-host the server for your team and manage it with the CLI.

## Core Features:

- Variable Creation: Securely store environment variables.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.

## Installation

Build from source (binaries in `target/release/`):
```bash
cargo build --release
```

Run the server (self-hosted — defaults to `./data/envoy.db`, JWT secret auto-generated and persisted to `./data/.jwt_secret`):
```bash
./target/release/envoy-server
```

With explicit config:
```bash
JWT_SECRET=... TURSO_DATABASE_URL=./data/envoy.db ./target/release/envoy-server
```

Run the CLI:
```bash
./target/release/envoy --help
```

Point the CLI at your server, then set it up:
```bash
envoy setup connect --url http://localhost:8080
envoy setup init        # first user becomes instance admin (fresh server only)
envoy setup login
envoy admin invite someone@example.com   # admin creates a one-time invite
envoy setup accept --token <token>       # invitee joins with their own password
```
