package rest

import (
	"menu/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

// Router binds handler to API endpoints.
type Router struct {
	server http.Server
}

// NewRouter creates and allocates new [Router].
func NewRouter(c *config.App, categoryHandler *CategoryHandler) *Router {
	switch c.Env {
	case config.Development:
		gin.SetMode(gin.DebugMode)
	case config.Production:
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	api := engine.Group("/api/v1")
	{
		{
			categories := api.Group("/categories")
			categories.POST("/", categoryHandler.AddCategory)
		}
	}

	return &Router{
		server: http.Server{
			Addr:    c.Port,
			Handler: engine,
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
