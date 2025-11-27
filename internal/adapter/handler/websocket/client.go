package websocket

import (
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// Client represents a connected user to the websocket.
type Client struct {
	Id         uuid.UUID
	IsAdmin    bool
	SessionID  uuid.UUID
	Connection *websocket.Conn
}

// NewClient creates a new Client instance.
func NewClient(id uuid.UUID, isAdmin bool, sessionID uuid.UUID, connection *websocket.Conn) *Client {
	return &Client{
		Id:         id,
		IsAdmin:    isAdmin,
		SessionID:  sessionID,
		Connection: connection,
	}
}
