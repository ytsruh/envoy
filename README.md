# Envoy
A secure environment variable management tool built using Rust. Self-host the server for your team and manage it with the CLI.

## Core Features:

- Variable Creation: Securely store environment variables.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.

## Installation

### CLI

No Rust toolchain needed once releases exist. For now, build from source:

```bash
make build-release # binaries in `target/release/`: `envoy`, `envoy-server`
./target/release/envoy --help
```

The examples below assume `envoy` is on your `PATH` — substitute `./target/release/envoy` if you built from source.

### Server (Docker)

```bash
docker build -t envoy-server .
docker run -p 8080:8080 -v envoy-data:/data envoy-server
curl http://localhost:8080/health
```

Data (SQLite + JWT secret) lives in the `envoy-data` volume (`/data` in the container). Don't skip the volume — without it redeploys wipe the DB and log everyone out.

## Self-hosting

Deploy however you like — bare binary, Docker, or any host that can run a container with a persistent volume. The server is stateless; all state is two files on disk:

| What | File | Configured by |
| ---- | ---- | ------------- |
| Database (SQLite) | `envoy.db` | `TURSO_DATABASE_URL` (a file path, or `libsql://...` for remote Turso) |
| JWT secret (auto-generated unless you set one) | `.jwt_secret` | `JWT_SECRET`, else `ENVOY_STATE_DIR/.jwt_secret` |

How the secret directory is resolved: `ENVOY_STATE_DIR` if set, else the database file's parent directory, else `./data`.

1. **Persist the state directory.** Mount a persistent volume at whatever path you choose (the image defaults to `/data`). Without this, redeploys wipe the DB and log everyone out.
2. **Set `JWT_SECRET`** to a long random string via your host's secret management. If unset, the server generates one and stores it as `.jwt_secret` in the state directory — it survives restarts as long as the volume exists, but losing it invalidates all sessions.
3. **Put TLS in front** (reverse proxy / load balancer / your host's HTTPS). The server itself speaks plain HTTP — never expose it bare over the internet.
4. **Bind:** `ADDR` wins, else `$PORT` (for hosts that inject it), else `0.0.0.0:8080`.
5. **Healthcheck:** point your host at `/health` (returns `{"health":"ok"}`; also wired to the Docker `HEALTHCHECK`).

Custom mount path example (any path works — `/data` is just the image default):

```bash
docker build -t envoy-server .
docker run -p 8080:8080 -v envoy-data:/mydata \
  -e TURSO_DATABASE_URL=/mydata/envoy.db \
  -e ENVOY_STATE_DIR=/mydata \
  envoy-server
```

Bare binary example:

```bash
TURSO_DATABASE_URL=/var/lib/envoy/envoy.db \
ENVOY_STATE_DIR=/var/lib/envoy \
JWT_SECRET=... \
ADDR=0.0.0.0:8080 \
./envoy-server
```

See `.env.example` for all variables.

Point the CLI at it:

```bash
envoy manage connect --url https://<your-server>
envoy manage init        # first user becomes instance admin (fresh server only)
envoy manage login
```

### Production checklist

- **Persistent volume**: required. Holds `envoy.db` and `.jwt_secret` (defaults to `/data` in the image; override with `TURSO_DATABASE_URL` / `ENVOY_STATE_DIR`). Back it up.
- **`JWT_SECRET`**: set it via your host's secret management. Auto-generation is fine for trying out, but losing `.jwt_secret` invalidates all sessions.
- **HTTPS**: terminate in front of the server. Never expose it bare over the internet.
- **Health**: `/health` returns `{"health":"ok"}` — wired to the Docker `HEALTHCHECK`; point your host's healthcheck at it too.
- **Logs**: `ENVOY_LOG_LEVEL=info` (default warn).

## Local dev

The examples below assume a release build — substitute `target/debug/` if you ran `make build`.

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
envoy manage connect --url http://localhost:8080
envoy manage init        # first user becomes instance admin (fresh server only)
envoy manage login
envoy manage invites create someone@example.com   # admin creates a one-time invite
envoy manage accept --token <token>       # invitee joins with their own password
```
