package domain

import "github.com/google/uuid"

// OrderSessionStatus is an enum for order status.
type OrderSessionStatus string

// Enum values
const (
	Closed OrderSessionStatus = "closed"
	Open   OrderSessionStatus = "open"
	Paid   OrderSessionStatus = "paid"
)

// OrderSession represents an order session  entity.
type OrderSession struct {
	Id          uuid.UUID
	TableNumber int
	Status      OrderSessionStatus
}

// NewSession creates a new OrderSession instance.
func NewSession(id uuid.UUID, tableNumber int, status OrderSessionStatus) *OrderSession {
	return &OrderSession{
		Id:          id,
		TableNumber: tableNumber,
		Status:      status,
	}
}

// OrderedProductStatus is an enum for ordered product status.
type OrderedProductStatus string

const (
	Pending   OrderedProductStatus = "pending"
	Preparing OrderedProductStatus = "preparing"
	Done      OrderedProductStatus = "done"
)

// OrderedProduct represents an ordered product entity.
type OrderedProduct struct {
	Id             uuid.UUID
	ProductId      uuid.UUID
	OrderSessionID uuid.UUID
	Status         OrderedProductStatus
}

// NewOrderedProduct creates a new OrderedProduct instance.
func NewOrderedProduct(id, productId, orderSessionID uuid.UUID, status OrderedProductStatus) *OrderedProduct {
	return &OrderedProduct{
		Id:             id,
		ProductId:      productId,
		OrderSessionID: orderSessionID,
		Status:         status,
	}
}

// UpdateOrderSessionDTO is a DTO for updating a order session.
type UpdateOrderSessionDTO struct {
	Id             uuid.UUID
	NewTableNumber *int
	NewStatus      *OrderSessionStatus
}

// NewUpdateOrderSessionDTO creates a new UpdateOrderSessionDTO instance.
func NewUpdateOrderSessionDTO(id uuid.UUID, newTableNumber *int, newStatus *OrderSessionStatus) *UpdateOrderSessionDTO {
	return &UpdateOrderSessionDTO{
		Id:             id,
		NewTableNumber: newTableNumber,
		NewStatus:      newStatus,
	}
}
