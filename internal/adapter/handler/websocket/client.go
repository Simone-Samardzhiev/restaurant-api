package websocket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represent a connected client to the WebSocket.
type Client struct {
	Id        uuid.UUID
	SessionId uuid.UUID
	Conn      *websocket.Conn
}

// NewClient creates a new Client instance.
func NewClient(sessionId uuid.UUID, conn *websocket.Conn) Client {
	return Client{
		Id:        uuid.New(),
		SessionId: sessionId,
		Conn:      conn,
	}
}

// Admin represent a connected admin to the WebSocket.
type Admin struct {
	Id   uuid.UUID
	Conn *websocket.Conn
}

// NewAdmin creates a new Admin instance.
func NewAdmin(conn *websocket.Conn) Admin {
	return Admin{
		Id:   uuid.New(),
		Conn: conn,
	}
}
