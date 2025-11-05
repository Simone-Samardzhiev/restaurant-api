package middleware

import (
	"restaurant/internal/adapter/handler/http/response"

	"github.com/gofiber/fiber/v2"
)

func NotFoundHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
			StatusCode: fiber.StatusNotFound,
			Code:       "endpoint_not_found",
			Messages: []string{
				"Endpoint not found",
			},
		})
	}
}
