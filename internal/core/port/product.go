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
}
