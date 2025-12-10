package port

import (
	"context"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// OrderRepository is an interface for interacting with orders data.
type OrderRepository interface {
	// GetSessions fetches all sessions.
	GetSessions(ctx context.Context) ([]domain.OrderSession, error)

	// GetSessionByID fetches a session by id.
	GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.OrderSession, error)

	// AddSession inserts a new order session.
	AddSession(ctx context.Context, session *domain.OrderSession) error

	// UpdateSession updates an order session by id.
	UpdateSession(ctx context.Context, session *domain.UpdateOrderSessionDTO) error

	// DeleteSession deletes a session by specific id.
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// AddOrderedProduct inserts an ordered product.
	AddOrderedProduct(ctx context.Context, product *domain.OrderedProduct) error

	// DeletePendingOrderedProduct deletes an ordered product only if the status is pending.
	DeletePendingOrderedProduct(ctx context.Context, orderedProductId uuid.UUID) (*domain.OrderedProduct, error)

	// DeleteOrderedProduct deletes an ordered product.
	DeleteOrderedProduct(ctx context.Context, orderedProductId uuid.UUID) (*domain.OrderedProduct, error)

	// UpdateOrderedProductStatus updates and returns the ordered product with updates status.
	UpdateOrderedProductStatus(ctx context.Context, id uuid.UUID, status domain.OrderedProductStatus) (*domain.OrderedProduct, error)
}

// OrderService is an interface for interacting with orders business login
type OrderService interface {
	// GetSessions fetches all sessions.
	GetSessions(ctx context.Context) ([]domain.OrderSession, error)

	// CreateSession creates a new order session.
	CreateSession(ctx context.Context) (*domain.OrderSession, error)

	// UpdateSession updates an order session by id.
	UpdateSession(ctx context.Context, session *domain.UpdateOrderSessionDTO) error

	// DeleteSession deletes a session by specific id.
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// ValidateSession validates the session exists and its open.
	ValidateSession(ctx context.Context, sessionId uuid.UUID) error

	// OrderProduct validates the session and adds the product.
	OrderProduct(ctx context.Context, productId uuid.UUID, sessionId uuid.UUID) (*domain.OrderedProduct, error)

	// DeleteOrderedProduct deletes the ordered product status.
	DeleteOrderedProduct(ctx context.Context, productId uuid.UUID, isPrivilegedCall bool) (*domain.OrderedProduct, error)

	// UpdateOrderedProductStatus updates and returns the ordered product with updates status.
	UpdateOrderedProductStatus(ctx context.Context, id uuid.UUID, status domain.OrderedProductStatus) (*domain.OrderedProduct, error)
}
