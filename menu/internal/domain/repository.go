package domain

import (
	"context"
	"time"

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

	// GetAllWithImage fetches all products with status different from [ProductStatusMissingImage].
	GetAllWithImage(ctx context.Context) ([]Product, error)

	// Update updates the data of a product.
	Update(ctx context.Context, request *UpdateProductRequest) error

	// Delete deletes a product by id.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus updates the status of a product by id.
	UpdateStatus(ctx context.Context, id uuid.UUID, status ProductStatus) error

	// MarkForImageUpdate updates the status of the product to [ProductStatusAwaitingImageUpdate] and adds
	// the new image key and content type.
	MarkForImageUpdate(ctx context.Context, id uuid.UUID, imageKey string, contentType ImageContentType) error

	// DeleteExpiredByStatus deletes all products whose status is [ProductStatusMissingImage] and a
	// set duration has passed since it was created.
	//
	// Returns the image keys for all delete products, so the images can be deleted
	// if the client has forgotten to confirm upload.
	DeleteExpiredByStatus(ctx context.Context, olderThan time.Duration) ([]string, error)
}
