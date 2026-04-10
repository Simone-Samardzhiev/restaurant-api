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

// DeleteAdmin removes an admin by id.
func (h *Hub) DeleteAdmin(id uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if admin, ok := h.admins[id]; ok {
		delete(h.admins, id)
		admin.conn.Close()
		close(admin.send)
	}
}

func (h *Hub) IsSessionOpen(id uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, ok := h.sessions[id]
	if !ok {
		return false
	}

	return session.Status.Equal("open")
}

func (h *Hub) AddClientToSession(sessionId uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	session, ok := h.sessions[sessionId]
	if !ok {
		return
	}
	session.AddClient(client)
}

// AddSession adds a session.
func (h *Hub) AddSession(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[session.Id] = session
}

// DeleteClient removes a client from session by session id and client id.
func (h *Hub) DeleteClient(sessionId uuid.UUID, clientId uuid.UUID) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, ok := h.sessions[sessionId]
	if !ok {
		return
	}
	session.DeleteClient(clientId)
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
