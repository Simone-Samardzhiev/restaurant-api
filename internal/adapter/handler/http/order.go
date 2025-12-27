package http

import (
	"net/http"
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler handler order related http requests.
type OrderHandler struct {
	orderService port.OrderService
}

// NewOrderHandler creates a new OrderHandler instance
func NewOrderHandler(orderService port.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) GetSessions(c *gin.Context) {
	sessions, err := h.orderService.GetSessions(c)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	res := make([]response.OrderSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, response.NewOrderSessionResponse(&session))
	}

	c.JSON(http.StatusOK, res)
}

func (h *OrderHandler) CreateSession(c *gin.Context) {
	order, err := h.orderService.CreateSession(c)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusOK, response.NewOrderSessionResponse(order))
}

func (h *OrderHandler) DeleteSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	err = h.orderService.DeleteSession(c, id)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.Status(http.StatusOK)
}

func (h *OrderHandler) GetOrderedProducts(c *gin.Context) {
	products, err := h.orderService.GetOrderedProducts(c)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}
	res := make([]response.OrderedProductResponse, 0, len(products))
	for _, product := range products {
		res = append(res, response.NewOrderedProductResponse(&product))
	}

	c.JSON(http.StatusOK, res)
}

func (h *OrderHandler) GetBill(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypePublic)
		return
	}

	bill, err := h.orderService.GetBill(c, id)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusOK, response.NewBillResponse(bill))
}
