package websocket

import (
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Hub keeps all connected clients and manages message between them.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[uuid.UUID]*Client
	admins     map[uuid.UUID]*Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		register: make(chan *Client),
		clients:  make(map[uuid.UUID]*Client),
		admins:   make(map[uuid.UUID]*Client),
	}
}

// Run start an infinity for loop listening on the channels.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if client.IsAdmin {
				h.admins[client.Id] = client
			} else {
				h.clients[client.Id] = client
			}
		case client := <-h.unregister:
			if client.IsAdmin {
				delete(h.admins, client.Id)
			} else {
				delete(h.clients, client.Id)
			}
		}
	}
}

// Handler handles websocket connections.
type Handler struct {
	hub          *Hub
	orderService port.OrderService
}

// NewHandler creates a new Handler instance.
func NewHandler(hub *Hub, orderService port.OrderService) *Handler {
	return &Handler{
		hub:          hub,
		orderService: orderService,
	}
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
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNoStatusReceived,
			) {
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

func (h *Handler) ConnectAsClient(conn *websocket.Conn) {
	sessionIdValue := conn.Locals("session_id")
	sessionId, ok := sessionIdValue.(uuid.UUID)
	if !ok {
		zap.L().Error("Missing session_id in locals")
	}

	client := NewClient(uuid.New(), false, sessionId, conn)
	h.hub.register <- client

	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNoStatusReceived,
			) {
				zap.L().Error("websocket connection closed", zap.Error(err))
			}

			break
		}
	}
}
