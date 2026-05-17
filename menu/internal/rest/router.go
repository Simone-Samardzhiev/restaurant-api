package rest

import (
	"context"
	"log/slog"
	"menu/internal/config"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Router binds handler to API endpoints.
type Router struct {
	server http.Server
}

// RouterConfig is configuration for [NewRouter].
type RouterConfig struct {
	App    *config.App
	Logger *slog.Logger
	Store  middleware.RateLimiterStore

	HeathHandler    *HealthHandler
	CategoryHandler *CategoryHandler
	ProductHandler  *ProductHandler
}

// NewRouter creates and allocates new [Router].
func NewRouter(c *RouterConfig) *Router {
	e := echo.NewWithConfig(echo.Config{
		Logger:           c.Logger,
		HTTPErrorHandler: ErrorHandler,
		IPExtractor:      echo.ExtractIPFromRealIPHeader(),
	})

	e.Use(loggerMiddleware)
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		Store: c.Store,
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
	}))
	e.RouteNotFound("/*", func(ctx *echo.Context) error {
		return ctx.JSON(http.StatusNotFound, ErrorResponse{
			Code:       "ENDPOINT_NOT_FOUND",
			HttpStatus: http.StatusNotFound,
			Message:    "Endpoint not found.",
			RequestID:  ctx.Request().Header.Get(echo.HeaderXRequestID),
		})
	})
	e.GET("/health", c.HeathHandler.IsHealthy)

	api := e.Group("/api/v1")
	{
		{
			categories := api.Group("/categories")
			categories.POST("", c.CategoryHandler.AddCategory)
			categories.GET("", c.CategoryHandler.GetCategories)
			categories.PATCH("/:id", c.CategoryHandler.UpdateCategory)
			categories.DELETE("/:id", c.CategoryHandler.DeleteCategory)
		}
		{
			products := api.Group("/products")
			products.POST("", c.ProductHandler.Add)
		}
	}

	return &Router{
		server: http.Server{
			Addr:    c.App.Addr,
			Handler: e,
		},
	}
}

// Start starts the underlying [http.Server].
//
// Start always returns a non-nil error.
// After Stop is called router returns [http.ErrServerClosed].
func (r *Router) Start() error {
	return r.server.ListenAndServe()
}

// Stop stops the underlying [http.Server].
func (r *Router) Stop(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
