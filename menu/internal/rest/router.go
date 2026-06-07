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

	e.Use(requestIdExtractor)
	e.Use(loggerMiddleware)
	e.RouteNotFound("/*", notFoundHandler)
	e.GET("/health", c.HeathHandler.IsHealthy)

	api := e.Group("/api/v1")
	api.Use(rateLimitMiddleware(c.Store))
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
			products.POST("", c.ProductHandler.AddProduct)
			products.GET("/:id/upload-info", c.ProductHandler.GetUploadInfo)
			products.POST("/:id/image", c.ProductHandler.ConfirmImageUpload)
			products.GET("/:id", c.ProductHandler.GetProduct)
			products.GET("", c.ProductHandler.GetAllProductsWithImage)
			products.GET("/image/:key", c.ProductHandler.GetImage)
			products.PATCH("/:id", c.ProductHandler.UpdateProduct)
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
