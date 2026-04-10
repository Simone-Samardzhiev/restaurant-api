package websocket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Admin represents a connected admin.
type Admin struct {
	Id   uuid.UUID
	conn *websocket.Conn
	send chan []byte
}

// NewAdmin allocates and creates a new [Admin].
func NewAdmin(id uuid.UUID, conn *websocket.Conn) *Admin {
	return &Admin{
		Id:   id,
		conn: conn,
		send: make(chan []byte, 256),
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

// ReadPump reads events from websocket connection.
func (a *Admin) ReadPump() {
	for {
		_, _, err := a.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
