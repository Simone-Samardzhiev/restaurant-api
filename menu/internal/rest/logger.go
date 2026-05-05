package rest

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
)

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		now := time.Now()
		requestId := c.Request().Header.Get(echo.HeaderXRequestID)

		err := next(c)

		c.Logger().LogAttrs(
			c.Request().Context(),
			slog.LevelDebug,
			"REQUEST",
			slog.String("method", c.Request().Method),
			slog.String("path", c.Request().URL.Path),
			slog.String("ip", c.RealIP()),
			slog.String("latency", time.Since(now).String()),
			slog.String("requestId", requestId),
		)

		return err
	}
}
