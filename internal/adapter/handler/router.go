package handler

import (
	"net/http"
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"

	"context"

	"github.com/gin-gonic/gin"
)

// Handlers holds all rest handlers.
type Handlers struct {
	CategoryHandler *rest.CategoryHandler
}

// Router routes all http request to the specific handler function.
type Router struct {
	server *http.Server
}

// NewRouter allocates and creates a new [Router] with set up router.
func NewRouter(container *config.Container, handlers Handlers) *Router {
	switch container.AppConfig.Environment {
	case config.Production:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ZapLogger())
	router.Use(middleware.Error())

	api := router.Group("/api/v1")
	{
		admin := api.Group("/admin")
		admin.Use(gin.BasicAuth(gin.Accounts{
			container.AuthConfig.Username: container.AuthConfig.Password,
		}))

		{
			menu := admin.Group("/menu")
			{
				categories := menu.Group("/categories")
				categories.POST("", handlers.CategoryHandler.AddCategory)
				categories.PATCH("/:id", handlers.CategoryHandler.UpdateCategory)
			}
		}
	}

	server := &http.Server{
		Addr:    container.AppConfig.Port,
		Handler: router,
	}
	return &Router{server: server}
}

// Run listens to the provided port by [NewRouter].
func (r *Router) Run() error {
	return r.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
