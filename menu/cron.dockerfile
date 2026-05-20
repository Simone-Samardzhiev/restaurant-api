FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o /app/bin/cron ./cmd/cron/main.go
CMD ["./bin/cron"]