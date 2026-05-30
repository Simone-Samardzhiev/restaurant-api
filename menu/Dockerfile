FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o /app/bin/api ./cmd/api/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/api ./bin/api
COPY ./migrations ./migrations

EXPOSE 8080
CMD ["./bin/api"]