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

	// UpdateProduct applies update to a product.
	UpdateProduct(ctx context.Context, request *UpdateProductRequest) error

	// UpdateImagePath updates image path of a product and returns the old one.
	UpdateImagePath(ctx context.Context, id uuid.UUID, path string) (string, error)

	// DeleteProduct deletes a product by id and return the image path for cleanup.
	DeleteProduct(ctx context.Context, id uuid.UUID) (string, error)

	// GetProducts fetches products by applying a filter.
	GetProducts(ctx context.Context, filter *ProductFilter) ([]Product, error)
}

// ProductService describes how product business logic is accessed.
type ProductService interface {
	// AddProduct adds a new product with image.
	AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error)

	// UpdateProduct applies update to a product.
	UpdateProduct(ctx context.Context, request *UpdateProductRequest) error

	// UpdateImage updates the image of a product and returns the new image path.
	UpdateImage(ctx context.Context, request *UpdateImageRequest) (string, error)

	// DeleteProduct deletes a product by id.
	DeleteProduct(ctx context.Context, id uuid.UUID) error

	// GetProducts fetches products by applying a filter.
	GetProducts(ctx context.Context, filter *ProductFilter) ([]Product, error)
}
