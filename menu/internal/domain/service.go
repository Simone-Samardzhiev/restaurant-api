package domain

import (
	"context"

	"github.com/google/uuid"
)

// CategoryService describes how category business logic is accessed.
type CategoryService interface {

	// Add creates a new category with specific name.
	Add(ctx context.Context, name string) (*Category, error)

	// GetAll fetches all categories.
	GetAll(ctx context.Context) ([]Category, error)

	Update(ctx context.Context, id uuid.UUID, name string) error
}
