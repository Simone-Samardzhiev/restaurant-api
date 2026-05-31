package domain

import (
	"context"
	"io"

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

// ProductService describes how product business logic is accessed.
type ProductService interface {
	// Add creates a new product draft.
	Add(ctx context.Context, request *AddProductRequest) (*ProductDraft, error)

	// GetDraft returns a draft for a product by id.
	GetDraft(ctx context.Context, id uuid.UUID) (*ProductDraft, error)

	// ConfirmImageUpload confirms an image is uploaded for a product.
	ConfirmImageUpload(ctx context.Context, productID uuid.UUID) error

	// GetProduct fetches a product by id.
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)

	// GetImage fetches an image by key.
	GetImage(ctx context.Context, key string) (io.ReadCloser, error)
}
