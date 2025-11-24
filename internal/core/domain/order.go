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

// OrderProductStatus is an enum for ordered product status.
type OrderProductStatus string

const (
	Pending   OrderProductStatus = "pending"
	Preparing OrderProductStatus = "preparing"
	Done      OrderProductStatus = "done"
)

// OrderedProduct represents an ordered product entity.
type OrderedProduct struct {
	Id        uuid.UUID
	ProductId uuid.UUID
	OrderId   uuid.UUID
}

// NewOrderedProduct creates a new OrderedProduct instance.
func NewOrderedProduct(id, productId uuid.UUID, orderId uuid.UUID) *OrderedProduct {
	return &OrderedProduct{
		Id:        id,
		ProductId: productId,
		OrderId:   orderId,
	}
}
