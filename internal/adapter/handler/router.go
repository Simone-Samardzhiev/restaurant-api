package handler

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http"
	"restaurant/internal/adapter/handler/http/middleware"
	"restaurant/internal/adapter/handler/http/response"

	"restaurant/internal/adapter/handler/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	fiberWebsocket "github.com/gofiber/websocket/v2"
)

// Router represents an HTTP router for requests.
type Router struct {
	app    *fiber.App
	config *config.AppConfig
}

// NewRouter creates a new Router instance.
func NewRouter(
	container *config.Container,
	productHandler *http.ProductHandler,
	orderHandler *http.OrderHandler,
	websocketHandler *websocket.Handler,
) *Router {
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
			Unauthorized: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponse{
					StatusCode: fiber.StatusUnauthorized,
					Code:       "unauthorized",
					Messages:   []string{"You are not authorized to access this resource"},
				})
			},
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
				order.Get("/sessions", orderHandler.GetSessions)
				order.Post("/sessions", orderHandler.CreateSession)
				order.Delete("/sessions/:id", orderHandler.DeleteSession)
				order.Get("/connect", fiberWebsocket.New(websocketHandler.Admin))
			}
		}

		public := v1.Group("/public")
		{
			public.Get("/product-categories", productHandler.GetProductCategories)
			public.Get("/products", productHandler.GetProducts)
			public.Get("/connect/:session", fiberWebsocket.New(websocketHandler.Client))
			public.Get("/bill/:id", orderHandler.GetBill)
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
