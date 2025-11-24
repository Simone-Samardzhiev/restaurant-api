package response

import (
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// OrderResponse represents an order response.
type OrderResponse struct {
	Id          uuid.UUID                 `json:"id"`
	TableNumber int                       `json:"tableNumber"`
	Status      domain.OrderSessionStatus `json:"status"`
}

// NewOrderResponse creates a new OrderResponse instance.
func NewOrderResponse(order *domain.OrderSession) OrderResponse {
	return OrderResponse{
		Id:          order.Id,
		TableNumber: order.TableNumber,
		Status:      order.Status,
	}
}
