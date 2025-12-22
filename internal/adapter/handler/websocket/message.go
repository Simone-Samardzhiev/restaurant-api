package websocket

import (
	"encoding/json"
	"restaurant/internal/core/domain"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// writeMessage writes a byte message to a connection and logs any errors.
func writeMessage(msg []byte, conn *websocket.Conn) {
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		zap.L().Error("error writing message", zap.Error(err))
	}
}

// writeString writes a string message to a connection and logs any errors.
func writeString(msg string, conn *websocket.Conn) {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		zap.L().Error("error writing message", zap.Error(err))
	}
}

// MessageType is an enum for message types.
type MessageType string

// MessageType enum values.
const (
	Order                                MessageType = "ORDER"
	SuccessfulOrder                      MessageType = "ORDER_OK"
	DeleteOrderedProduct                 MessageType = "DELETE_ORDERED_PRODUCT"
	SuccessfulDeletionOfOrderedProduct   MessageType = "DELETE_ORDERED_PRODUCT_OK"
	UpdateOrderedProductStatus           MessageType = "UPDATE_ORDERED_PRODUCT_STATUS"
	SuccessfulUpdateOrderedProductStatus MessageType = "UPDATE_ORDERED_PRODUCT_STATUS_OK"
	UpdateSession                        MessageType = "UPDATE_SESSION"
	SuccessfulUpdateSession              MessageType = "UPDATE_SESSION_OK"
	Pay                                  MessageType = "PAY"
	SuccessfulPayment                    MessageType = "PAY_OK"
)

// Message represent a websocket message.
type Message struct {
	Type MessageType     `json:"type" binding:"required,messageType"`
	Data json.RawMessage `json:"data" binding:"required"`
}

// NewMessage creates a new Message instance.
func NewMessage(messageType MessageType, data json.RawMessage) Message {
	return Message{
		Type: messageType,
		Data: data,
	}
}

// OrderData represent the message data for ordering a product.
type OrderData struct {
	ProductID uuid.UUID `json:"productId" validate:"required"`
}

// SuccessfulOrderData represent a successful message when order is accepted.
type SuccessfulOrderData struct {
	Id        uuid.UUID                   `json:"id"`
	ProductID uuid.UUID                   `json:"productId"`
	SessionId uuid.UUID                   `json:"sessionId"`
	Status    domain.OrderedProductStatus `json:"status"`
}

// NewSuccessfulOrderData creates a new SuccessfulOrderData instance.
func NewSuccessfulOrderData(id, productID, sessionId uuid.UUID, status domain.OrderedProductStatus) SuccessfulOrderData {
	return SuccessfulOrderData{
		Id:        id,
		ProductID: productID,
		SessionId: sessionId,
		Status:    status,
	}
}

// DeleteOrderedProductData represents the message data for deleting an ordered product.
type DeleteOrderedProductData struct {
	Id uuid.UUID `json:"id" validate:"required"`
}

// SuccessfulDeletionOfOrderedProductData represent a successful message when
// deletion of an ordered product is successful.
type SuccessfulDeletionOfOrderedProductData struct {
	Id uuid.UUID `json:"id" validate:"required"`
}

// NewSuccessfulDeletionOfOrderedProductData creates a new SuccessfulDeletionOfOrderedProductData instance.
func NewSuccessfulDeletionOfOrderedProductData(id uuid.UUID) SuccessfulDeletionOfOrderedProductData {
	return SuccessfulDeletionOfOrderedProductData{
		Id: id,
	}
}

// UpdateOrderedProductStatusData represent an update of a product status.
type UpdateOrderedProductStatusData struct {
	Id     uuid.UUID                   `json:"id" validate:"required"`
	Status domain.OrderedProductStatus `json:"status" validate:"orderedProductStatus"`
}

// UpdateOrderSessionData represents the message data for updating an order session
type UpdateOrderSessionData struct {
	Id          uuid.UUID                  `json:"id" validate:"required"`
	TableNumber *int                       `json:"tableNumber" validate:"omitempty,min=1"`
	Status      *domain.OrderSessionStatus `json:"status" validate:"omitempty,orderStatus"`
}

// SuccessfulUpdateOrderSessionData represent a successful message when order session update is successful.
type SuccessfulUpdateOrderSessionData struct {
	Id          uuid.UUID                 `json:"id"`
	TableNumber int                       `json:"tableNumber"`
	Status      domain.OrderSessionStatus `json:"status"`
}

// NewSuccessfulUpdateOrderSessionData creates a new SuccessfulUpdateOrderSessionData instance.
func NewSuccessfulUpdateOrderSessionData(id uuid.UUID, tableNumber int, status domain.OrderSessionStatus) SuccessfulUpdateOrderSessionData {
	return SuccessfulUpdateOrderSessionData{
		Id:          id,
		TableNumber: tableNumber,
		Status:      status,
	}
}

type PaymentData struct {
	Id uuid.UUID `json:"id" validate:"required"`
}

// Broadcast represent a broadcast to a specific session id.
type Broadcast struct {
	Message   Message
	SessionId uuid.UUID
}

// NewBroadcast creates a new Broadcast instance.
func NewBroadcast(message Message, sessionId uuid.UUID) *Broadcast {
	return &Broadcast{
		Message:   message,
		SessionId: sessionId,
	}
}
