package websocket

import (
	"net/http"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Handler handles all WebSocket connections.
type Handler struct {
	hub          *hub
	orderService port.OrderService
	upgrader     websocket.Upgrader
}

// NewHandler creates a new Handler instance.
func NewHandler(hub *hub, orderService port.OrderService) *Handler {
	return &Handler{
		hub:          hub,
		orderService: orderService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// / handleWebsocketError logs unexpected WebSocket closing errors.
func handleWebsocketError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		zap.L().Error("websocket close with unexpected error", zap.Error(err))
	}
}

func (h *Handler) ConnectAsAdmin(c *gin.Context) {
	if !c.IsWebsocket() {
		c.JSON(http.StatusBadRequest, notWebSocketUpgradeErrorMessage)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("websocket upgrade failed", zap.Error(err))
		return
	}

	admin := NewAdmin(conn)
	h.hub.registerAdmin <- admin
	defer func() {
		h.hub.unregisterAdmin <- admin
		conn.Close()
	}()

	for {
		messageType, message, readErr := conn.ReadMessage()
		if readErr != nil {
			handleWebsocketError(readErr)
			return
		}

		if writeErr := conn.WriteMessage(messageType, message); writeErr != nil {
			handleWebsocketError(writeErr)
			return
		}
	}
}

func (h *Handler) ConnectAsClient(c *gin.Context) {
	rawID := c.Param("id")
	sessionId, err := uuid.Parse(rawID)
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypePublic)
		return
	}

	if err = h.orderService.ValidateSession(c.Request.Context(), sessionId); err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypePublic)
		return
	}

	if !c.IsWebsocket() {
		c.JSON(http.StatusBadRequest, notWebSocketUpgradeErrorMessage)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := NewClient(sessionId, conn)
	h.hub.registerClient <- client
	defer func() {
		h.hub.unregisterClient <- client
		conn.Close()
	}()

	for {
		messageType, message, readErr := conn.ReadMessage()
		if readErr != nil {
			handleWebsocketError(readErr)
			return
		}

		if writeErr := conn.WriteMessage(messageType, message); writeErr != nil {
			handleWebsocketError(writeErr)
			return
		}
	}
}
