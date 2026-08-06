FROM rust:latest AS builder

WORKDIR /app

COPY . .
RUN cargo build --release


FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY --from=builder /app/target/release/api ./api
COPY ./migrations ./migrations

EXPOSE 8080
CMD ["./api"]
