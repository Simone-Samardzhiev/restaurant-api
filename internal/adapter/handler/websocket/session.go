package websocket

import (
	"sync"

	"github.com/google/uuid"
)

// Session represents a order session with connected clients.
type Session struct {
	Id      uuid.UUID
	mu      sync.RWMutex
	clients map[uuid.UUID]*Client
}

// NewSession allocates and creates a new [Session].
func NewSession(id uuid.UUID) *Session {
	return &Session{
		Id:      id,
		clients: make(map[uuid.UUID]*Client),
	}
}

// AddClient adds a new client.
func (s *Session) AddClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.Id] = client
}

// DeleteClient deletes a new client.
func (s *Session) DeleteClient(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[id]; ok {
		delete(s.clients, id)
		client.conn.Close()
		close(client.send)
	}
}

// Broadcast sends a message to all clients in the session.
func (s *Session) Broadcast(message []byte) {
	s.mu.Lock()
	defer s.mu.Lock()

	for _, client := range s.clients {
		select {
		case client.send <- message:
		default:
			delete(s.clients, client.Id)
			client.conn.Close()
			close(client.send)
		}
	}
}
