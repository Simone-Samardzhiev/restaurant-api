FROM golang:latest AS builder

WORKDIR /app

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -gcflags="all=-N -l" -o /app/bin/api ./cmd/api/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app

COPY --from=builder /go/bin/dlv /usr/local/bin/dlv
COPY --from=builder /app/bin/api ./bin/api
COPY ./migrations ./migrations

EXPOSE 8080 2345

CMD ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./bin/api"]