package rest

import (
	"context"
	"menu/internal/config"
	"net/http"

	"github.com/labstack/echo/v5"
)

// Router binds handler to API endpoints.
type Router struct {
	server http.Server
}

// NewRouter creates and allocates new [Router].
func NewRouter(c *config.App, heathHandler *HealthHandler, categoryHandler *CategoryHandler) *Router {
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
	})

	e.RouteNotFound("/*", handleEndpointNotFound)
	e.GET("/health", heathHandler.IsHealthy)

	api := e.Group("/api/v1")
	{
		{
			categories := api.Group("/categories")
			categories.POST("", categoryHandler.AddCategory)
		}
	}

	return &Router{
		server: http.Server{
			Addr:    c.Addr,
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
