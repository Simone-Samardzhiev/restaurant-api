package domain

import "github.com/google/uuid"

// OrderStatus is an enum for order status.
type OrderStatus string

// Enum values
const (
	Closed OrderStatus = "closed"
	Open   OrderStatus = "open"
	Paid   OrderStatus = "paid"
)

// Order represents an order entity.
type Order struct {
	Id          uuid.UUID
	TableNumber int
	Status      OrderStatus
}

// NewOrder creates a new Product instance.
func NewOrder(id uuid.UUID, tableNumber int, status OrderStatus) *Order {
	return &Order{
		Id:          id,
		TableNumber: tableNumber,
		Status:      status,
	}
}

// OrderProductStatus is an enum for ordered products status.
type OrderProductStatus string

const (
	Pending   OrderProductStatus = "pending"
	Preparing OrderProductStatus = "preparing"
	Done      OrderProductStatus = "Done"
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
