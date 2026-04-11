package websocket

import (
	"encoding/json"
	"restaurant/internal/domain/order"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Admin represents a connected admin.
type Admin struct {
	Id             uuid.UUID
	conn           *websocket.Conn
	sessionService order.SessionService
	hub            *Hub
	send           chan []byte
}

// NewAdmin allocates and creates a new [Admin].
func NewAdmin(id uuid.UUID, conn *websocket.Conn, sessionService order.SessionService, hub *Hub) *Admin {
	return &Admin{
		Id:             id,
		conn:           conn,
		sessionService: sessionService,
		hub:            hub,
		send:           make(chan []byte, 256),
	}
}

// WritePump listens for messages over the send channel and sends them over the
// websocket connection.
func (a *Admin) WritePump() {
	for message := range a.send {
		if err := a.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// AddSessionRequest represents the request data for adding a new session.
type AddSessionRequest struct {
	Table  int    `json:"table"`
	Status string `json:"status"`
}

// AddSessionResponse represents the response data of an order session.
type AddSessionResponse struct {
	Id     uuid.UUID `json:"id"`
	Table  int       `json:"table"`
	Status string    `json:"status"`
}

// addSession handles [AddSessionEvent].
func (a *Admin) addSession(message *Message) {
	var req AddSessionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		a.send <- handleInvalidJSON(err)
		return
	}

	domainRequest, err := order.ParseAddSessionRequest(req.Table, req.Status)
	if err != nil {
		a.send <- handleDomainError(err)
		return
	}

	result, err := a.sessionService.AddSession(context.Background(), domainRequest)
	if err != nil {
		a.send <- handleDomainError(err)
		return
	}
	a.hub.AddSession(NewSession(result.Id, result.Status))

	data, err := json.Marshal(AddSessionResponse{
		Id:     result.Id,
		Table:  result.Table.Number(),
		Status: result.Status.String(),
	})
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}

	response := Message{
		Event: SessionAddedEvent,
		Data:  data,
	}

	body, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}
	a.hub.Broadcast(uuid.Nil, body)
}

// UpdateSessionRequest represents the request data for updating a session.
type UpdateSessionRequest struct {
	Id     uuid.UUID `json:"id"`
	Table  *int      `json:"table"`
	Status *string   `json:"status"`
}

// UpdateSessionResponse represents the response data of successfully updated session.
type UpdateSessionResponse struct {
	Id     uuid.UUID `json:"id"`
	Table  *int      `json:"table,omitempty"`
	Status *string   `json:"status,omitempty"`
}

// updateSession handles [UpdateSessionEvent].
func (a *Admin) updateSession(message *Message) {
	var req UpdateSessionRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		a.send <- handleInvalidJSON(err)
		return
	}

	domainRequest, err := order.ParseUpdateSessionRequest(req.Id, req.Table, req.Status)
	if err != nil {
		a.send <- handleDomainError(err)
		return
	}

	if err = a.sessionService.UpdateSession(context.Background(), domainRequest); err != nil {
		a.send <- handleDomainError(err)
		return
	}

	data, err := json.Marshal(UpdateSessionResponse{
		Id:     domainRequest.Id,
		Table:  req.Table,
		Status: req.Status,
	})
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}

	response := Message{
		Event: SessionUpdatedEvent,
		Data:  data,
	}
	body, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}

	a.hub.Broadcast(domainRequest.Id, body)
}

// ReadPump reads events from websocket connection.
func (a *Admin) ReadPump() {
	for {
		_, msg, err := a.conn.ReadMessage()
		if err != nil {
			return
		}

		var message Message

		if err = json.Unmarshal(msg, &message); err != nil {
			a.send <- handleInvalidJSON(err)
			continue
		}

		switch message.Event {
		case AddSessionEvent:
			a.addSession(&message)
		case UpdateSessionEvent:
			a.updateSession(&message)
		default:
			a.send <- handleInvalidEvent()
		}
	}
}
