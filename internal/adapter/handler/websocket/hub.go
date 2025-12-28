package websocket

import "github.com/google/uuid"

// hub holds the connection to the clients.
type hub struct {
	clients          map[uuid.UUID]Client
	registerClient   chan Client
	unregisterClient chan Client

	admins          map[uuid.UUID]Admin
	registerAdmin   chan Admin
	unregisterAdmin chan Admin
}

func newHub() *hub {
	return &hub{
		clients:          make(map[uuid.UUID]Client),
		registerClient:   make(chan Client),
		unregisterClient: make(chan Client),

		admins:        make(map[uuid.UUID]Admin),
		registerAdmin: make(chan Admin),
	}
}

func (h *hub) Listen() {
	for {
		select {
		case client := <-h.registerClient:
			h.clients[client.Id] = client
		case client := <-h.unregisterClient:
			delete(h.clients, client.Id)

		case admin := <-h.registerAdmin:
			h.admins[admin.Id] = admin
		case admin := <-h.unregisterAdmin:
			delete(h.admins, admin.Id)
		}
	}
}
