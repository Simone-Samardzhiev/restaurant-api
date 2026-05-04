package rest

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/net/context"
)

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		now := time.Now()
		err := next(c)

		c.Logger().LogAttrs(
			context.Background(),
			slog.LevelDebug,
			"REQUEST",
			slog.String("method", c.Request().Method),
			slog.String("path", c.Request().URL.Path),
			slog.String("ip", c.RealIP()),
			slog.String("latency", time.Since(now).String()),
		)

		return err
	}
}
