package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles websocket connections.
type Handler struct {
	hub          *Hub
	orderService port.OrderService
	validator    *validator.Validate
}

// NewHandler creates a new Handler instance.
func NewHandler(hub *Hub, orderService port.OrderService, validator *validator.Validate) *Handler {
	return &Handler{
		hub:          hub,
		orderService: orderService,
		validator:    validator,
	}
}

func writeMessage(msg []byte, conn *websocket.Conn) {
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		zap.L().Error("websocket write message failed", zap.Error(err))
	}
}

func writeStringMessage(msg string, conn *websocket.Conn) {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		zap.L().Error("websocket write message failed", zap.Error(err))
	}
}

func isUnexpectedCloseError(err error) bool {
	return websocket.IsUnexpectedCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
		websocket.CloseNoStatusReceived,
	)
}

func (h *Handler) ConnectAsAdmin(conn *websocket.Conn) {
	client := NewClient(uuid.New(), true, uuid.Nil, conn)
	h.hub.register <- client
	defer func() {
		h.hub.unregister <- client
	}()

	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			if isUnexpectedCloseError(err) {
				zap.L().Error("websocket connection closed", zap.Error(err))
			}

			break
		}
	}
}

func (h *Handler) ValidateClientConnection(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	sessionId, err := uuid.Parse(c.Params("session_id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}

	if err = h.orderService.ValidateSession(c.Context(), sessionId); err != nil {
		return err
	}

	c.Locals("session_id", sessionId)
	return c.Next()
}

func (h *Handler) handleClientOrder(ctx context.Context, message *Message, client *Client) {
	var orderMessageData OrderMessage
	if err := json.Unmarshal(message.Data, &orderMessageData); err != nil {
		writeStringMessage("Invalid json format", client.Connection)
		return
	}

	if err := h.validator.Struct(orderMessageData); err != nil {
		writeStringMessage("Invalid json format", client.Connection)
		return
	}

	orderedProductId, err := h.orderService.OrderProduct(
		ctx,
		orderMessageData.ProductId,
		client.SessionID,
	)

	switch {
	case err == nil:
		dataPayload, encodeErr := json.Marshal(SuccessfulOrderMessage{
			ProductId:        orderMessageData.ProductId,
			OrderedProductId: orderedProductId,
			SessionId:        client.SessionID,
		})
		if encodeErr != nil {
			zap.L().Error("json encode error", zap.Error(encodeErr))
			return
		}

		resMessage := Message{
			Type: SuccessfulOrder,
			Data: dataPayload,
		}

		resBytes, encodeErr := json.Marshal(resMessage)
		if encodeErr != nil {
			zap.L().Error("json encode error", zap.Error(encodeErr))
			return
		}

		writeMessage(resBytes, client.Connection)

		h.hub.broadcast <- NewBroadcast(client.Id, client.SessionID, resBytes)
	case errors.Is(err, domain.ErrOrderSessionIsNotOpen):
		writeStringMessage("Session is not open!", client.Connection)
	case errors.Is(err, domain.ErrProductNotFound):
		writeStringMessage("Product not found", client.Connection)
	default:
		zap.L().Error("error ordering a product", zap.Error(err))
		writeStringMessage("Internal server error", client.Connection)
	}
}

func (h *Handler) ConnectAsClient(conn *websocket.Conn) {
	sessionIdValue := conn.Locals("session_id")
	sessionId, ok := sessionIdValue.(uuid.UUID)
	if !ok {
		zap.L().Error("Missing session_id in locals")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient(uuid.New(), false, sessionId, conn)
	h.hub.register <- client
	defer func() {
		cancel()
		h.hub.unregister <- client
	}()

	for {
		msgType, msg, err := conn.ReadMessage()

		if err != nil {
			if isUnexpectedCloseError(err) {
				zap.L().Error("reading message error", zap.Error(err))
			}
			break
		}

		if msgType != websocket.TextMessage {
			writeStringMessage("Unexpected message type, only text allowed", conn)
			continue
		}

		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			writeStringMessage("Invalid json format", conn)
			continue
		}

		switch message.Type {
		case Order:
			h.handleClientOrder(ctx, &message, client)
		}
	}
}
