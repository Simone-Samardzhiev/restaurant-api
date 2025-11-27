package websocket

import (
	"encoding/json"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

type MessageType string

const (
	Order  MessageType = "order"
	Update MessageType = "update"
	Delete MessageType = "delete"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OrderData struct {
	ProductId uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}

type UpdateData struct {
	ProductId uuid.UUID                 `json:"productId"`
	Status    domain.OrderProductStatus `json:"status"`
}

type DeleteData struct {
	ProductId uuid.UUID `json:"productId"`
}
