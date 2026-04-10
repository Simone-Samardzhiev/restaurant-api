package websocket

import (
	"restaurant/internal/domain/order"
	"sync"

	"github.com/google/uuid"
)

// Hub manages all websockets connections.
type Hub struct {
	mu       sync.RWMutex
	admins   map[uuid.UUID]*Admin
	sessions map[uuid.UUID]*Session
}

// NewHub allocates and creates a new [Hub] with sessions.
func NewHub(sessions ...order.Session) *Hub {
	sessionsMap := make(map[uuid.UUID]*Session, len(sessions))
	for _, session := range sessions {
		s := NewSession(session.Id, session.Status)
		sessionsMap[s.Id] = s
	}

	return &Hub{
		admins:   make(map[uuid.UUID]*Admin),
		sessions: sessionsMap,
	}
}

// AddAdmin adds a new admin.
func (h *Hub) AddAdmin(admin *Admin) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.admins[admin.Id] = admin
}

// RemoveAdmin removes an admin by id.
func (h *Hub) RemoveAdmin(id uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if admin, ok := h.admins[id]; ok {
		delete(h.admins, id)
		admin.conn.Close()
		close(admin.send)
	}
}

// AddSession adds a session.
func (h *Hub) AddSession(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[session.Id] = session
}

// RemoveSession removes a session by id.
func (h *Hub) RemoveSession(id uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, id)
}

// Broadcast sends a message to all admins and all client by the specified session.
// If the session is [uuid.Nil] the message will be only send
func (h *Hub) Broadcast(sessionId uuid.UUID, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sessionId != uuid.Nil {
		if session, ok := h.sessions[sessionId]; ok {
			session.Broadcast(message)
		}
	}

	for _, admin := range h.admins {
		select {
		case admin.send <- message:
		default:
			delete(h.admins, admin.Id)
			admin.conn.Close()
			close(admin.send)
		}
	}
}
