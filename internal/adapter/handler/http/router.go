package http

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http/middleware"
	"restaurant/internal/adapter/handler/http/response"

	"github.com/gofiber/fiber/v2/middleware/basicauth"

	"github.com/gofiber/fiber/v2"
)

// Router represents an HTTP router for requests.
type Router struct {
	app    *fiber.App
	config *config.AppConfig
}

// NewRouter creates a new Router instance.
func NewRouter(appConfig *config.AppConfig, authConfig *config.AuthConfig, productHandler *ProductHandler) *Router {
	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(middleware.ZapLogger())

	v1 := app.Group("/api/v1")
	{
		products := v1.Group("/products")
		products.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authConfig.Username: authConfig.Password,
			},
			Realm: "admin",
		}))
		{
			products.Post("", productHandler.AddProduct)
		}
	}
	app.Use(middleware.NotFoundHandler())

	return &Router{
		app:    app,
		config: appConfig,
	}
}

// Listen server HTTP at the provided port
func (r *Router) Listen() error {
	return r.app.Listen(r.config.Port)
}

// Shutdown gracefully shutdowns the app.
func (r *Router) Shutdown() error {
	return r.app.Shutdown()
}
