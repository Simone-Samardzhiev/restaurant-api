package websocket

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles order related requests.
type Handler struct {
	upgrader              websocket.Upgrader
	sessionService        order.SessionService
	orderedProductService order.OrderedProductService
	hub                   *Hub
}

// NewHandler allocates and creates a new [Handler].
func NewHandler(sessionService order.SessionService, orderedProductService order.OrderedProductService, hub *Hub) *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		sessionService:        sessionService,
		orderedProductService: orderedProductService,
		hub:                   hub,
	}
}

// ConnectAsAdmin handles connecting as admin.
func (h *Handler) ConnectAsAdmin(ctx *gin.Context) {
	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	admin := NewAdmin(uuid.New(), conn, h.sessionService, h.orderedProductService, h.hub)
	h.hub.AddAdmin(admin)

	go func() {
		admin.WritePump()
		h.hub.DeleteAdmin(admin.Id)
	}()

	go func() {
		admin.ReadPump()
		h.hub.DeleteAdmin(admin.Id)
	}()
}

// ConnectAsClient handles connecting as client.
func (h *Handler) ConnectAsClient(ctx *gin.Context) {
	sessionId, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidUUID, err)).
			SetType(gin.ErrorTypeBind)
		return
	}

	if !h.hub.IsSessionOpen(sessionId) {
		ctx.Error(
			domain.NewBadRequestError(
				"cannot connect to non open session",
				domain.ErrorCodeSessionNotOpened,
				nil),
		).SetType(gin.ErrorTypeBind)
		return
	}

	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(uuid.New(), sessionId, conn, h.orderedProductService, h.hub)
	h.hub.AddClientToSession(sessionId, client)

	go func() {
		client.ReadPump()
		h.hub.DeleteClient(sessionId, client.Id)
	}()
	go func() {
		client.WritePump()
		h.hub.DeleteClient(sessionId, client.Id)
	}()
}
