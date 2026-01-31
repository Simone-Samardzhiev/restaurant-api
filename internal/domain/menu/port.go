package menu

import (
	"io"

	"context"

	"github.com/google/uuid"
)

// CategoryRepository describes how category data is stored and managed.
type CategoryRepository interface {
	// AddCategory stores a new Category.
	AddCategory(ctx context.Context, category *Category) error

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, update *CategoryUpdate) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying the filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)
}

// ProductRepository describes how product data is stored and managed.
type ProductRepository interface {
	// AddProduct stores a new product.
	AddProduct(ctx context.Context, product *Product) error

	// UpdateProduct updates an existing Product.
	UpdateProduct(ctx context.Context, update *ProductUpdate) error

	// UpdateProductImagePath updates the image path of an existing Product.
	UpdateProductImagePath(ctx context.Context, id uuid.UUID, path string) error

	// GetProductImagePathById fetches a product image path by id.
	GetProductImagePathById(ctx context.Context, id uuid.UUID) (string, error)

	// GetProductImagePaths fetches paths for all images used by products in a map/set.
	GetProductImagePaths(ctx context.Context) (map[string]struct{}, error)

	// GetProducts fetches all products.
	GetProducts(ctx context.Context) ([]Product, error)

	// DeleteProduct deletes an existing Product by id and returns the image path linked with it.
	DeleteProduct(ctx context.Context, id uuid.UUID) (string, error)
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

// CategoryService describes category-related business logic.
type CategoryService interface {
	// AddCategory creates and stores a new Category from AddCategoryRequest.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategories fetches categories by applying the filter.
	GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error)
}

// ProductService describes product-related business logic.
type ProductService interface {
	// AddProduct creates and stores a new Product and image from AddProductRequest.
	AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error)

	// UpdateProduct updates an existing Product.
	UpdateProduct(ctx context.Context, request *UpdateProductRequest) error

	// ReplaceProductImage replaces the image of an existing Product.
	ReplaceProductImage(ctx context.Context, request *ReplaceProductImageRequest) (string, error)

	// DeleteOrphanImages deletes all orphan images than aren't used by any products.
	DeleteOrphanImages(ctx context.Context)

	// DeleteProduct deletes an existing Product by id.
	DeleteProduct(ctx context.Context, id uuid.UUID) error

	// GetProducts fetches all products.
	GetProducts(ctx context.Context) ([]Product, error)
}
