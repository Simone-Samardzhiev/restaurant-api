package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles order related requests.
type Handler struct {
	upgrader websocket.Upgrader
	hub      *Hub
}

// NewHandler allocates and creates a new [Handler].
func NewHandler() *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hub: NewHub(),
	}
}

// ConnectAsAdmin handles connecting as admin.
func (h *Handler) ConnectAsAdmin(ctx *gin.Context) {
	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	admin := NewAdmin(uuid.New(), conn)
	h.hub.AddAdmin(admin)

	go func() {
		admin.WritePump()
		h.hub.RemoveAdmin(admin.Id)
	}()

	go func() {
		admin.ReadPump()
		h.hub.RemoveAdmin(admin.Id)
	}()
}
