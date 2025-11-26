package request

import "restaurant/internal/core/domain"

// UpdateOrderSessionRequest represent a request to update order session.
type UpdateOrderSessionRequest struct {
	NewTableNumber *int                       `json:"newTableNumber" validate:"omitempty,min=1"`
	NewStatus      *domain.OrderSessionStatus `json:"newStatus" validate:"omitempty,orderStatus"`
}
