package handler

import (
	"net/http"
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type Router struct {
	server *http.Server
}

func NewRouter(container *config.Container, productHandler *rest.ProductHandler) *Router {
	switch container.AppConfig.Environment {
	case config.Production:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(middleware.ZapLogger())
	g.Use(middleware.ErrorMiddleware())

	v1 := g.Group("/api/v1")
	{
		admin := v1.Group("/admin")
		admin.Use(gin.BasicAuth(gin.Accounts{
			container.AuthConfig.Username: container.AuthConfig.Password,
		}))
		admin.POST("/login", func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})

		{
			menu := admin.Group("/menu")
			menu.POST("/categories", productHandler.AddCategory)
			menu.PATCH("/categories/:id", productHandler.UpdateCategory)
			menu.DELETE("/categories/:id", productHandler.DeleteCategory)
		}
	}
	{
		public := v1.Group("/public")
		public.GET("/categories", productHandler.GetCategories)
	}

	return &Router{
		server: &http.Server{
			Handler: g,
			Addr:    container.AppConfig.Port,
		},
	}
}

func (r *Router) Start() error {
	return r.server.ListenAndServe()
}

func (r *Router) Stop(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
