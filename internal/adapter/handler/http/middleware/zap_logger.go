package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		zap.L().Debug(
			"incoming request",
			zap.String("latency", time.Since(now).String()),
			zap.String("clientIP", clientIP),
			zap.String("path", path),
			zap.String("method", method),
		)
	}
}
