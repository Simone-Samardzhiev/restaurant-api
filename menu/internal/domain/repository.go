package domain

import (
	"context"

	"github.com/google/uuid"
)

// CategoryRepository describes how category data is accessed.
type CategoryRepository interface {

	// Save saves a new category.
	Save(ctx context.Context, category *Category) error

	// GetAll fetches all categories.
	GetAll(ctx context.Context) ([]Category, error)

	// Update updates the category name by id.
	Update(ctx context.Context, id uuid.UUID, name string) error
}
