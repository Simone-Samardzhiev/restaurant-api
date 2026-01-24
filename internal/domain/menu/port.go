package menu

import (
	"io"

	"context"

	"github.com/google/uuid"
)

// Repository describes how menu data is stored and managed.
type Repository interface {
	// AddCategory stores a new Category.
	AddCategory(ctx context.Context, category *Category) error

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, update *CategoryUpdate) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying the filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)

	// AddProduct stores a new Product.
	AddProduct(ctx context.Context, product *Product) error

	UpdateProduct(ctx context.Context, update *ProductUpdate) error

	// GetProductImagePaths fetches paths for all images used by products in a map/set.
	GetProductImagePaths(ctx context.Context) (map[string]struct{}, error)
}

// ImageRepository describes how image data is stored and managed.
type ImageRepository interface {
	// AddImage stores a new image.
	AddImage(ctx context.Context, data io.Reader, imageType ImageType) (string, error)

	// DeleteImage deletes an image with specific path.
	DeleteImage(ctx context.Context, path string) error

	// GetAllImagePaths fetches paths for all images in a map/set.
	GetAllImagePaths(ctx context.Context) (map[string]struct{}, error)
}

// Service describes menu-related business logic.
type Service interface {
	// AddCategory creates and stores a new Category from AddCategoryRequest.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying the filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)

	// AddProduct creates and stores a new Product and image from AddProductRequest.
	AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error)

	UpdateProduct(ctx context.Context, update *UpdateProductRequest) error

	// DeleteOrphanImages deletes all orphan images than aren't used by any products.
	DeleteOrphanImages(ctx context.Context)
}
