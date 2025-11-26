package http

import (
	"net/http"
	"restaurant/internal/adapter/handler/http/request"
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/domain"
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

func (h *OrderHandler) GetSessions(c *fiber.Ctx) error {
	sessions, err := h.orderService.GetSessions(c.Context())
	if err != nil {
		return err
	}

	res := make([]response.OrderSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, response.NewOrderSessionResponse(&session))
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *OrderHandler) CreateSession(c *fiber.Ctx) error {
	order, err := h.orderService.CreateSession(c.Context())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewOrderSessionResponse(order))
}

func (h *OrderHandler) UpdateSession(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}

	var req request.UpdateOrderSessionRequest
	if err = c.BodyParser(&req); err != nil {
		return err
	}

	if err = h.validator.Struct(req); err != nil {
		return err
	}

	if err = h.orderService.UpdateSession(
		c.Context(),
		domain.NewUpdateOrderSessionDTO(id, req.NewTableNumber, req.NewStatus),
	); err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (h *OrderHandler) DeleteSession(c *fiber.Ctx) error {
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
