package logger

import (
	"context"
	"log/slog"
)

// silentHandler implements [slog.Handler] discarding any messages.
type silentHandler struct{}

func (s *silentHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (s *silentHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (s *silentHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s
}

func (s *silentHandler) WithGroup(name string) slog.Handler {
	return s
}

// NewSilentLogger return a logger that discard any logs.
func NewSilentLogger() *slog.Logger {
	return slog.New(&silentHandler{})
}
