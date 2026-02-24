package menu

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// CategoryRepository describes how categories data is managed.
type CategoryRepository interface {
	// SaveCategory saves a category.
	SaveCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error

	// DeleteCategory deletes a category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying a filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)
}

// CategoryService describes how category business logic is accessed.
type CategoryService interface {
	// AddCategory adds a category.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error

	// DeleteCategory deletes a category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying a filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)
}

// ImageRepository describes how image data is managed.
type ImageRepository interface {
	// SaveImage saves an image and returns the path to it.
	SaveImage(ctx context.Context, data io.Reader, imageType ImageType) (string, error)

	// DeleteImage deletes an image by path.
	DeleteImage(ctx context.Context, path string) error
}

// ProductRepository described how product data is managed.
type ProductRepository interface {
	// SaveProduct saves a product.
	SaveProduct(ctx context.Context, request *SaveProductRequest) (*Product, error)
}

// ProductService describes how product business logic is accessed.
type ProductService interface {
	// AddProduct adds a new product with image.
	AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error)
}
