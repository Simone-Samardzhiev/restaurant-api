package websocket

import "github.com/google/uuid"

type Broadcast struct {
	ClientId  uuid.UUID
	SessionId uuid.UUID
	Message   []byte
}

func NewBroadcast(clientId, sessionId uuid.UUID, message []byte) *Broadcast {
	return &Broadcast{
		ClientId:  clientId,
		SessionId: sessionId,
		Message:   message,
	}
}

// Hub keeps all connected clients and manages message between them.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Broadcast
	clients    map[uuid.UUID]*Client
	admins     map[uuid.UUID]*Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Broadcast),
		clients:    make(map[uuid.UUID]*Client),
		admins:     make(map[uuid.UUID]*Client),
	}
}

// Run start an infinity for loop listening on the channels.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if client.IsAdmin {
				h.admins[client.Id] = client
			} else {
				h.clients[client.Id] = client
			}

		case client := <-h.unregister:
			if client.IsAdmin {
				delete(h.admins, client.Id)
			} else {
				delete(h.clients, client.Id)
			}

		case broadcast := <-h.broadcast:
			for _, client := range h.clients {
				if client.SessionID == broadcast.SessionId && client.Id != broadcast.ClientId {
					writeMessage(broadcast.Message, client.Connection)
				}
			}

			for _, admin := range h.admins {
				if admin.Id != broadcast.ClientId {
					writeMessage(broadcast.Message, admin.Connection)
				}
			}
		}
	}
}
