package port

import (
	"context"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// ProductRepository is an interface for interacting with product data.
type ProductRepository interface {
	// AddCategory saves a new product category.
	AddCategory(ctx context.Context, category *domain.ProductCategory) error
	// UpdateCategory updates an existing category.
	UpdateCategory(ctx context.Context, dto *domain.UpdateCategoryProductDTO) error
	// DeleteCategory deletes a category by specified id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	// AddProduct saves a new product.
	AddProduct(ctx context.Context, product *domain.Product) error
	// UpdateProduct updates an existing product.
	UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error
	// UpdateProductImagePath updates the image path of specif product by id.
	UpdateProductImagePath(ctx context.Context, id uuid.UUID, path *string) error
	// DeleteProductById deletes a product by specified id.
	DeleteProductById(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	// DeleteProductsByCategory deletes all products by specified category id.
	DeleteProductsByCategory(ctx context.Context, categoryId uuid.UUID) ([]domain.Product, error)
}

// ProductService is an interface for interacting with product business logic.
type ProductService interface {
	// AddCategory saves a new product category.
	AddCategory(ctx context.Context, name string) error
	// UpdateCategory updates an existing category.
	UpdateCategory(ctx context.Context, dto *domain.UpdateCategoryProductDTO) error
	// DeleteCategory deletes a category by specified id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	// AddProduct saves a new product with linked image.
	AddProduct(ctx context.Context, dto *domain.AddProductDTO) error
	// UpdateProduct updates an existing product.
	UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error
	// AddImage adds a new image to a product.
	AddImage(ctx context.Context, image *domain.Image, productId uuid.UUID) error
	// DeleteProduct deletes a product with filters.
	DeleteProduct(ctx context.Context, dto *domain.DeleteProductDTO) error
}
