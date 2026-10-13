FROM rust:1.97-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libssl-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY Cargo.toml Cargo.lock ./
COPY crates ./crates
RUN cargo build --release -p envoy-server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/envoy-server .

VOLUME /data

EXPOSE 8080
ENTRYPOINT ["./envoy-server"]
