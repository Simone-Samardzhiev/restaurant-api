package order

import (
	"context"
)

// SessionRepository describes how session data is managed.
type SessionRepository interface {
	// Save saves a new session.
	Save(ctx context.Context, request *AddSessionRequest) (*Session, error)

	// Get fetches all sessions
	Get(ctx context.Context) ([]Session, error)

	// Update updates a session by id.
	Update(ctx context.Context, request *UpdateSessionRequest) error
}

// SessionService describes how session business logic is access.
type SessionService interface {
	// AddSession adds a new session.
	AddSession(ctx context.Context, request *AddSessionRequest) (*Session, error)

	// UpdateSession updates a session.
	UpdateSession(ctx context.Context, request *UpdateSessionRequest) error
}

// OrderedProductRepository describes how ordered products data is managed.
type OrderedProductRepository interface {
	// Save saves a new ordered product.
	Save(ctx context.Context, request *AddOrderedProductRequest) (*OrderedProduct, error)
}
