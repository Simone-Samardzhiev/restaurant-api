package http

import (
	"restaurant/internal/adapter/config"

	"github.com/gofiber/fiber/v2"
)

// Router represents an HTTP router for requests.
type Router struct {
	app    *fiber.App
	config *config.AppConfig
}

// NewRouter creates a new Router instance.
func NewRouter(config *config.AppConfig) *Router {
	app := fiber.New()

	return &Router{
		app:    app,
		config: config,
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
