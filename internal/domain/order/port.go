package order

import "context"

// SessionRepository describes how session data is managed.
type SessionRepository interface {
	// Save saves a new session.
	Save(ctx context.Context, request *AddSessionRequest) (*Session, error)
}
