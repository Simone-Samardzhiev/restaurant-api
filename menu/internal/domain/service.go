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

// ProductService describes how product business logic is accessed.
type ProductService interface {
	// Add creates a new product draft.
	Add(ctx context.Context, request *AddProductRequest) (*ProductUploadInfo, error)

	// GetUploadInfo returns the upload info for a product by id.
	GetUploadInfo(ctx context.Context, id uuid.UUID) (*ProductUploadInfo, error)

	// ConfirmImageUpload confirms an image is uploaded for a product.
	ConfirmImageUpload(ctx context.Context, productID uuid.UUID) error

	// GetProduct fetches a product by id.
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)

	// GetAllWithImage fetches all products with image.
	GetAllWithImage(ctx context.Context) ([]Product, error)

	// GetImage fetches an image by key.
	GetImage(ctx context.Context, key string) (*Image, error)

	// UpdateProduct updates the data of a product.
	UpdateProduct(ctx context.Context, request *UpdateProductRequest) error

	// MarkProductForImageUpdate marks a product for image update.
	MarkProductForImageUpdate(ctx context.Context, id uuid.UUID, contentType ImageContentType) error

	// Delete deletes a product by id.
	Delete(ctx context.Context, id uuid.UUID) error
}
