package port

import (
	"context"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// OrderRepository is an interface for interacting with orders data.
type OrderRepository interface {
	// AddSession inserts a new order session.
	AddSession(ctx context.Context, session *domain.OrderSession) error
	// DeleteSession deletes a session by specific id.
	DeleteSession(ctx context.Context, id uuid.UUID) error
}

// OrderService is an interface for interacting with orders business login
type OrderService interface {
	// CreateSession creates a new order session.
	CreateSession(ctx context.Context) (*domain.OrderSession, error)
	// DeleteSession deletes a session by specific id.
	DeleteSession(ctx context.Context, id uuid.UUID) error
}
