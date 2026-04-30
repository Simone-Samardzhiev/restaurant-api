package rest

import (
	"context"
	"log/slog"
	"menu/internal/config"
	"net/http"

	"github.com/labstack/echo/v5"
)

// Router binds handler to API endpoints.
type Router struct {
	server http.Server
}

type RouterConfig struct {
	App             *config.App
	Logger          *slog.Logger
	HeathHandler    *HealthHandler
	CategoryHandler *CategoryHandler
}

// NewRouter creates and allocates new [Router].
func NewRouter(c *RouterConfig) *Router {
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
		Logger:           c.Logger,
		IPExtractor:      echo.ExtractIPFromRealIPHeader(),
	})

	e.Use(loggerMiddleware)

	e.RouteNotFound("/*", handleEndpointNotFound)
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
