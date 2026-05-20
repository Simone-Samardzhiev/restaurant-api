package logger

import (
	"context"
	"log/slog"
	"menu/internal/config"
	"os"
)

// handlerWithRequestId extracts the request id from with key [RequestIdKey] when providing [context.Context]
// to the logger.
type handlerWithRequestId struct {
	inner slog.Handler
}

type requestIdKey string

const RequestIdKey requestIdKey = "requestId"

func (h *handlerWithRequestId) Handle(ctx context.Context, record slog.Record) error {
	if requestId, ok := ctx.Value(RequestIdKey).(string); ok {
		record.AddAttrs(slog.String("requestId", requestId))
	}
	return h.inner.Handle(ctx, record)
}

func (h *handlerWithRequestId) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *handlerWithRequestId) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.inner.WithAttrs(attrs)
}

func (h *handlerWithRequestId) WithGroup(name string) slog.Handler {
	return h.inner.WithGroup(name)
}

func newHandlerWithRequestId(h slog.Handler) *handlerWithRequestId {
	return &handlerWithRequestId{h}
}

// New creates new [slog.Logger] based on the app configuration.
func New(c *config.App) *slog.Logger {
	var inner slog.Handler
	switch c.Env {
	case config.Development:
		inner = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case config.Production:
		inner = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})
	}

	return slog.New(newHandlerWithRequestId(inner))
}
