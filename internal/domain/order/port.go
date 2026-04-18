package order

import (
	"context"

	"github.com/google/uuid"
)

// SessionRepository describes how session data is managed.
type SessionRepository interface {
	// Save saves a new session.
	Save(ctx context.Context, request *AddSessionRequest) (*Session, error)

	// Get fetches all sessions
	Get(ctx context.Context) ([]Session, error)

	// GetById fetches session by id.
	GetById(ctx context.Context, id uuid.UUID) (*Session, error)

	// Update updates a session by id.
	Update(ctx context.Context, request *UpdateSessionRequest) error
}

// SessionService describes how session business logic is accessed.
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

// OrderedProductService describes how ordered products business logic is accessed.
type OrderedProductService interface {
	// PlaceOrder places a new order to a session.
	PlaceOrder(ctx context.Context, request *AddOrderedProductRequest) (*OrderedProduct, error)
}
