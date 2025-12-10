package websocket

import (
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type Client struct {
	Id        uuid.UUID
	SessionId uuid.UUID
	Conn      *websocket.Conn
}

func NewClient(sessionId uuid.UUID, conn *websocket.Conn) *Client {
	return &Client{
		Id:        uuid.New(),
		SessionId: sessionId,
		Conn:      conn,
	}
}

type Admin struct {
	Id   uuid.UUID
	Conn *websocket.Conn
}

func NewAdmin(conn *websocket.Conn) *Admin {
	return &Admin{
		Id:   uuid.New(),
		Conn: conn,
	}
}
