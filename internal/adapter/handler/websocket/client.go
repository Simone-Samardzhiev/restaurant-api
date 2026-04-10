package websocket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a connected client.
type Client struct {
	Id        uuid.UUID
	SessionId uuid.UUID
	conn      *websocket.Conn
	send      chan []byte
}

// NewClient allocates and creates a new client with 256 buffered channel.
func NewClient(id, sessionId uuid.UUID, conn *websocket.Conn) *Client {
	return &Client{
		Id:        id,
		SessionId: sessionId,
		conn:      conn,
		send:      make(chan []byte, 256),
	}
}

// WritePump listens for messages over the send channel and sends them over the
// websocket connection.
func (c *Client) WritePump() {
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// ReadPump reads events from websocket connection.
func (c *Client) ReadPump() {
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
