FROM rust:1.88-bookworm AS build

WORKDIR /src
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /src/target/release/nexgou-benchmark-actix-web /app/actix-web-benchmark
RUN mkdir -p /app/data
ENV PORT=3004 DB_PATH=/app/data/db.sqlite SERVICE_NAME=actix-web SERVICE_VERSION=2.0.0 LOG_LEVEL=silent
EXPOSE 3004
CMD ["/app/actix-web-benchmark"]
