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

	// Update updates a category name by id.
	Update(ctx context.Context, id uuid.UUID, name string) error

	// Delete deletes a category by id.
	Delete(ctx context.Context, id uuid.UUID) error
}
