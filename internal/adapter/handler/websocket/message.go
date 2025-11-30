package websocket

import (
	"encoding/json"

	"github.com/google/uuid"
)

// MessageType is an enum for message types
type MessageType string

const (
	Order           MessageType = "order"
	SuccessfulOrder MessageType = "successful_order"
)

// Message represent a websocket message.
type Message struct {
	Type MessageType     `json:"type" validate:"required,messageType"`
	Data json.RawMessage `json:"data" validate:"required"`
}

// OrderMessage represents the data needed for ordering a new product.
type OrderMessage struct {
	ProductId uuid.UUID `json:"productId" validate:"required"`
}

// SuccessfulOrderMessage represents the data send when order is ordering a product is completed.
type SuccessfulOrderMessage struct {
	ProductId        uuid.UUID `json:"productId"`
	OrderedProductId uuid.UUID `json:"orderedProductId"`
}
