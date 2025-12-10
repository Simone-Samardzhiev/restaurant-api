package websocket

import (
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Hub struct {
	clients map[uuid.UUID]Client
	admins  map[uuid.UUID]Admin

	registerClient   chan *Client
	unregisterClient chan *Client

	registerAdmin   chan *Admin
	unregisterAdmin chan *Admin

	broadcast chan *Broadcast
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]Client),
		admins:  make(map[uuid.UUID]Admin),

		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),

		registerAdmin:   make(chan *Admin),
		unregisterAdmin: make(chan *Admin),

		broadcast: make(chan *Broadcast),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerClient:
			h.clients[client.Id] = *client
		case client := <-h.unregisterClient:
			delete(h.clients, client.Id)

		case admin := <-h.registerAdmin:
			h.admins[admin.Id] = *admin
		case admin := <-h.unregisterAdmin:
			delete(h.admins, admin.Id)

		case broadcast := <-h.broadcast:
			messageData, err := json.Marshal(broadcast.Message)
			if err != nil {
				zap.L().Error("error encoding broadcast message", zap.Error(err))
			}

			for _, client := range h.clients {
				if client.SessionId == broadcast.SessionId {
					writeMessage(messageData, client.Conn)
				}
			}

			for _, admin := range h.admins {
				writeMessage(messageData, admin.Conn)
			}
		}

	}
}
