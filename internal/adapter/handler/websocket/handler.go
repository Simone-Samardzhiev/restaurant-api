package websocket

import (
	"encoding/json"
	"restaurant/internal/core/port"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Handler represent a handler for websocket connections.
type Handler struct {
	orderService port.OrderService
	hub          *Hub
	validator    *validator.Validate
}

// NewHandler creates a new Handler instance.
func NewHandler(orderService port.OrderService, hub *Hub, validator *validator.Validate) *Handler {
	return &Handler{
		orderService: orderService,
		hub:          hub,
		validator:    validator,
	}
}

// handleOrderedProductDeletion handles ordered product deletion.
func (h *Handler) handleOrderedProductDeletion(ctx context.Context, message *Message, isPrivilegedCall bool, conn *websocket.Conn) {
	var deletionData DeleteOrderedProductData
	if err := json.Unmarshal(message.Data, &deletionData); err != nil {
		writeString("Invalid json data", conn)
		return
	}

	if err := h.validator.Struct(deletionData); err != nil {
		writeString("Invalid json data", conn)
		return
	}

	deletedProduct, err := h.orderService.DeleteOrderedProduct(ctx, deletionData.Id, isPrivilegedCall)
	if err != nil {
		handleDomainError(conn, err)
		return
	}

	data, encodeErr := json.Marshal(NewSuccessfulDeletionOfOrderedProductData(deletionData.Id))
	if encodeErr != nil {
		zap.L().Error("Error encoding message", zap.Error(encodeErr))
		writeString("Internal server error", conn)
		return
	}

	h.hub.broadcast <- NewBroadcast(NewMessage(SuccessfulDeletionOfOrderedProduct, data), deletedProduct.OrderSessionID)
}

// handleUpdatingOrderedProductStatus handles updating product statuses
func (h *Handler) handleUpdatingOrderedProductStatus(ctx context.Context, message *Message, conn *websocket.Conn) {
	var updatingData UpdateOrderedProductStatusData
	if err := json.Unmarshal(message.Data, &updatingData); err != nil {
		writeString("Invalid json data", conn)
		return
	}

	if err := h.validator.Struct(updatingData); err != nil {
		writeString("Invalid json data", conn)
	}

	updatedProduct, err := h.orderService.UpdateOrderedProductStatus(ctx, updatingData.Id, updatingData.Status)
	if err != nil {
		handleDomainError(conn, err)
		return
	}

	h.hub.broadcast <- NewBroadcast(NewMessage(SuccessfulUpdateOrderedProductStatus, message.Data), updatedProduct.OrderSessionID)
}

// Admin handles admin websocket session.
func (h *Handler) Admin(conn *websocket.Conn) {
	admin := NewAdmin(conn)
	ctx, cancel := context.WithCancel(context.Background())
	h.hub.registerAdmin <- admin

	defer func() {
		cancel()
		h.hub.unregisterAdmin <- admin
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if !isExpectedCloseError(err) {
				zap.L().Error("error reading message", zap.Error(err))
			}
			return
		}

		if msgType != websocket.TextMessage {
			writeString("Unexpected message type, only text messages are supported", conn)
			return
		}

		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			writeString("Unexpected json", conn)
			return
		}

		if err = h.validator.Struct(&message); err != nil {
			writeString("Unexpected json", conn)
		}

		switch message.Type {
		case DeleteOrderedProduct:
			h.handleOrderedProductDeletion(ctx, &message, true, conn)
		case UpdateOrderedProductStatus:
			h.handleUpdatingOrderedProductStatus(ctx, &message, conn)
		default:
			writeString("Unexpected message type", conn)
		}
	}

}

// validateClientSession validates the client session and return the session id
// and bool variable representing if the session is valid.
func (h *Handler) validateClientSession(ctx context.Context, conn *websocket.Conn) (uuid.UUID, bool) {
	sessionId, err := uuid.Parse(conn.Params("session"))
	if err != nil {
		writeString("Invalid session id format", conn)
		return uuid.Nil, false
	}

	if err = h.orderService.ValidateSession(ctx, sessionId); err != nil {
		writeString("The session is not open. Please connect the waiter.", conn)
		return uuid.Nil, false
	}
	return sessionId, true
}

// handleOrder handles order message from clients.
func (h *Handler) handleOrder(ctx context.Context, message *Message, sessionId uuid.UUID, conn *websocket.Conn) {
	var orderData OrderData
	if err := json.Unmarshal(message.Data, &orderData); err != nil {
		writeString("Invalid json data", conn)
		return
	}
	if err := h.validator.Struct(orderData); err != nil {
		writeString("Invalid json data", conn)
		return
	}

	orderedProduct, err := h.orderService.OrderProduct(ctx, orderData.ProductID, sessionId)
	if err != nil {
		handleDomainError(conn, err)
		return
	}

	data, encodeErr := json.Marshal(
		NewSuccessfulOrderData(
			orderedProduct.Id,
			orderedProduct.ProductId,
			orderedProduct.OrderSessionID,
			orderedProduct.Status,
		),
	)
	if encodeErr != nil {
		zap.L().Error("Error encoding message", zap.Error(encodeErr))
		writeString("Internal server error", conn)
		return
	}

	h.hub.broadcast <- NewBroadcast(NewMessage(SuccessfulOrder, data), sessionId)
}

// Client handles client websocket session.
func (h *Handler) Client(conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	sessionId, ok := h.validateClientSession(ctx, conn)
	if !ok {
		return
	}

	client := NewClient(sessionId, conn)
	h.hub.registerClient <- client

	defer func() {
		cancel()
		h.hub.unregisterClient <- client
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if !isExpectedCloseError(err) {
				zap.L().Error("error reading message", zap.Error(err))
			}
			return
		}

		if msgType != websocket.TextMessage {
			writeString("Unexpected message type, only text messages are supported", conn)
			return
		}

		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			writeString("Unexpected json", conn)
			return
		}

		if err = h.validator.Struct(&message); err != nil {
			writeString("Unexpected json", conn)
			return
		}

		switch message.Type {
		case Order:
			h.handleOrder(ctx, &message, sessionId, conn)
		case DeleteOrderedProduct:
			h.handleOrderedProductDeletion(ctx, &message, false, conn)
		default:
			writeString("Unexpected message type", conn)
		}
	}
}
