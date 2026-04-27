package rest

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)

		c.Logger().Debug(
			"Request passed",
			slog.String("method", c.Request().Method),
			slog.String("host", c.Request().Host),
			slog.String("path", c.Request().URL.Path),
			slog.String("ip", c.RealIP()),
		)

		return err
	}
}
