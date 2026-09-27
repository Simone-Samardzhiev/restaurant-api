FROM rust:latest AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    protobuf-compiler \
    libprotobuf-dev \
    && rm -rf /var/lib/apt/lists/

COPY ./auth .
COPY ./proto ../proto
RUN cargo build --release


FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY --from=builder /app/target/release/api ./api

COPY ./auth/migrations ./migrations

EXPOSE 8080
CMD ["./api"]
