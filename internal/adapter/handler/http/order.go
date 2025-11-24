package http

import (
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/port"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// OrderHandler handler order related http requests.
type OrderHandler struct {
	orderService port.OrderService
	validator    *validator.Validate
}

// NewOrderHandler creates a new OrderHandler instance
func NewOrderHandler(orderService port.OrderService, validator *validator.Validate) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		validator:    validator,
	}
}

func (h *OrderHandler) AddOrder(c *fiber.Ctx) error {
	order, err := h.orderService.CreateSession(c.Context())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewOrderResponse(order))
}

func (h *OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return err
	}

	err = h.orderService.DeleteSession(c.Context(), id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}
