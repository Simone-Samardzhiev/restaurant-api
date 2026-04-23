FROM golang:latest AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o /app/bin/api ./cmd/api/main.go

FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/bin/api ./bin/api
COPY ./migrations ./migrations

EXPOSE 8080
CMD ["./bin/api"]