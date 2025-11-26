package response

import (
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// OrderSessionResponse represents an order response.
type OrderSessionResponse struct {
	Id          uuid.UUID                 `json:"id"`
	TableNumber int                       `json:"tableNumber"`
	Status      domain.OrderSessionStatus `json:"status"`
}

// NewOrderSessionResponse creates a new OrderSessionResponse instance.
func NewOrderSessionResponse(order *domain.OrderSession) OrderSessionResponse {
	return OrderSessionResponse{
		Id:          order.Id,
		TableNumber: order.TableNumber,
		Status:      order.Status,
	}
}
