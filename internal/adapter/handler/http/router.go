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
		admin := v1.Group("/admin")
		admin.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authConfig.Username: authConfig.Password,
			},
			Realm: "admin",
		}))
		{
			admin.Get("/login", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})
			admin.Post("/categories", productHandler.AddProductCategory)
			admin.Patch("/categories/:id", productHandler.UpdateCategory)
			admin.Delete("/categories/:id", productHandler.DeleteCategory)

			admin.Post("/admin", productHandler.AddProduct)
			admin.Patch("/admin/:id", productHandler.UpdateProduct)
			admin.Delete("/admin", productHandler.DeleteProduct)
			admin.Put("/admin/:id/image", productHandler.AddImage)
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
