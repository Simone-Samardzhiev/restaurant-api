package rest

import (
	"context"
	"log/slog"
	"menu/internal/logger"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// loggerMiddleware is a middleware used for logging request information.
func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		now := time.Now()
		err := next(c)

		c.Logger().LogAttrs(
			c.Request().Context(),
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

// requestIdExtractor is a middleware that will extract the request id
// from the header "X-Request-Id" and insert it as value in the request context
// with key [logger.RequestIdKey].
func requestIdExtractor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		requestId := ctx.Request().Header.Get(echo.HeaderXRequestID)
		if requestId != "" {
			ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), logger.RequestIdKey, requestId)))
		}
		return next(ctx)
	}
}

// notFoundHandler returns [ErrorResponse] when the api endpoint is not found.
func notFoundHandler(ctx *echo.Context) error {
	return ctx.JSON(http.StatusNotFound, ErrorResponse{
		Code:       "ENDPOINT_NOT_FOUND",
		HttpStatus: http.StatusNotFound,
		Message:    "Endpoint not found.",
		RequestID:  ctx.Request().Header.Get(echo.HeaderXRequestID),
	})
}

// rateLimitMiddleware returns rate limiter returning [ErrorResponse]
// when the request goes over the limit.
func rateLimitMiddleware(store middleware.RateLimiterStore) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		Store: store,
		ErrorHandler: func(c *echo.Context, err error) error {
			return err
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Code:       ErrorCodeTooManyRequests,
				HttpStatus: http.StatusTooManyRequests,
				Message:    "Too many requests. Please try again later.",
				RequestID:  c.Request().Header.Get(echo.HeaderXRequestID),
			})
		},
	})
}
