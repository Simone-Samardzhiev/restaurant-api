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
func NewRouter(container *config.Container, productHandler *ProductHandler, orderHandler *OrderHandler) *Router {
	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(middleware.ZapLogger())

	v1 := app.Group("/api/v1")
	{
		admin := v1.Group("/admin")
		admin.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				container.AuthConfig.Username: container.AuthConfig.Password,
			},
			Realm: "admin",
		}))
		{
			admin.Post("/login", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			menu := admin.Group("/menu")
			{
				menu.Post("/categories", productHandler.AddProductCategory)
				menu.Patch("/categories/:id", productHandler.UpdateCategory)
				menu.Delete("/categories/:id", productHandler.DeleteCategory)

				menu.Post("/products", productHandler.AddProduct)
				menu.Patch("/products/:id", productHandler.UpdateProduct)
				menu.Put("/products/:id/image", productHandler.ReplaceProductImage)
				menu.Delete("/products", productHandler.DeleteProduct)
			}

			order := admin.Group("/orders")
			{
				order.Post("/sessions", orderHandler.AddOrder)
				order.Delete("/sessions/:id", orderHandler.DeleteOrder)
			}
		}

		public := v1.Group("/public")
		{
			public.Get("/product-categories", productHandler.GetProductCategories)
			public.Get("/products", productHandler.GetProducts)
		}

	}
	app.Use(middleware.NotFoundHandler())

	return &Router{
		app:    app,
		config: &container.AppConfig,
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
