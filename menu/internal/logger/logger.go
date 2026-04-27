package logger

import (
	"log/slog"
	"menu/internal/config"
	"os"
)

// New creates new [slog.Logger] based on the app configuration.
func New(c *config.App) *slog.Logger {
	switch c.Env {
	case config.Development:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}))
	case config.Production:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}))
	}

	return slog.Default()
}
