package handler

import (
	nethttp "net/http"
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http"
	"restaurant/internal/adapter/handler/http/middleware"

	"restaurant/internal/adapter/handler/http/validator"

	"restaurant/internal/adapter/handler/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/net/context"
)

// Router represents an HTTP router for requests.
type Router struct {
	server *nethttp.Server
}

// NewRouter creates a new Router instance.
func NewRouter(
	container *config.Container,
	validator *validator.Validator,
	productHandler *http.ProductHandler,
	orderHandler *http.OrderHandler,
	websocketHandler *websocket.Handler,
) *Router {
	switch container.AppConfig.Environment {
	case config.Development:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	binding.Validator = validator

	app := gin.New()
	app.Use(gin.Recovery())
	app.Use(middleware.ZapLogger())
	app.Use(middleware.Error(validator.Translator))

	{
		v1 := app.Group("/api/v1")
		{
			admin := v1.Group("/admin")
			admin.Use(gin.BasicAuth(gin.Accounts{
				container.AuthConfig.Username: container.AuthConfig.Password,
			}))

			admin.POST("/login", func(c *gin.Context) {
				c.Status(nethttp.StatusOK)
			})

			{
				menu := admin.Group("/menu")
				menu.POST("/categories", productHandler.AddProductCategory)
				menu.PATCH("/categories/:id", productHandler.UpdateCategory)
				menu.DELETE("/categories/:id", productHandler.DeleteCategory)

				menu.POST("/products", productHandler.AddProduct)
				menu.PATCH("/products/:id", productHandler.UpdateProduct)
				menu.PUT("/products/:id/image", productHandler.ReplaceProductImage)
				menu.DELETE("/products/:id", productHandler.DeleteProduct)
			}

			{
				order := admin.Group("/orders")
				order.GET("/sessions", orderHandler.GetSessions)
				order.POST("/sessions", orderHandler.CreateSession)
				order.DELETE("/sessions/:id", orderHandler.DeleteSession)

				order.GET("/ordered-products", orderHandler.GetOrderedProducts)
				order.GET("/connect", websocketHandler.ConnectAsAdmin)
			}
		}
		{
			public := v1.Group("/public")
			public.GET("/product-categories", productHandler.GetProductCategories)
			public.GET("/products", productHandler.GetProducts)
			public.GET("/bill/:id", orderHandler.GetBill)
		}
	}

	return &Router{
		server: &nethttp.Server{
			Addr:    container.AppConfig.Port,
			Handler: app,
		},
	}
}

// Listen server HTTP at the provided port
func (r *Router) Listen() error {
	return r.server.ListenAndServe()
}

// Shutdown gracefully shutdowns the app.
func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
