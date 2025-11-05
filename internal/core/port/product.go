package port

import (
	"context"
	"restaurant/internal/core/domain"
)

// ProductRepository is an interface for interacting with product data.
type ProductRepository interface {
	// AddProduct saves a new product.
	AddProduct(ctx context.Context, product *domain.Product) error
}

// ProductService is an interface for interacting with product business logic.
type ProductService interface {
	// AddProduct saves a new product with linked image.
	AddProduct(ctx context.Context, dto *domain.AddProductDTO) error
}
