FROM rust:1.97-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libssl-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY Cargo.toml Cargo.lock ./
COPY crates ./crates
RUN cargo build --release -p envoy-server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/envoy-server .

# Single source of truth for persisted state (SQLite + auto-generated
# .jwt_secret). Mount a volume here or both are lost on redeploy.
ENV TURSO_DATABASE_URL=/data/envoy.db
ENV ENVOY_STATE_DIR=/data
ENV ADDR=0.0.0.0:8080

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD sh -c 'curl -f http://localhost:${PORT:-8080}/health || exit 1'
ENTRYPOINT ["./envoy-server"]
