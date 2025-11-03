package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ZapLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		now := time.Now()
		path := c.Path()
		method := c.Method()
		clientIP := c.IP()

		err := c.Next()

		zap.L().Debug(
			"incoming request",
			zap.String("latency", time.Since(now).String()),
			zap.String("clientIP", clientIP),
			zap.String("path", path),
			zap.String("method", method),
		)
		return err
	}
}
