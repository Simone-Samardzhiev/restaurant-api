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

	// Update updates a category name by id.
	Update(ctx context.Context, id uuid.UUID, name string) error

	// Delete deletes a category by id.
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProductRepository describes how product data is accessed.
type ProductRepository interface {

	// Save saves a new product.
	Save(ctx context.Context, product *Product) error

	// Get fetches a product by id.
	Get(ctx context.Context, id uuid.UUID) (*Product, error)
}
