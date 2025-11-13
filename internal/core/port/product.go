package port

import (
	"context"
	"io"
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

	// GetProductCategories fetches all product categories.
	GetProductCategories(ctx context.Context) ([]domain.ProductCategory, error)

	// AddProduct saves a new product.
	AddProduct(ctx context.Context, product *domain.Product) error

	// UpdateProduct updates an existing product.
	UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error

	// UpdateProductImage replaces the image data of a product.
	UpdateProductImage(ctx context.Context, productId uuid.UUID, image *domain.Image) error

	// DeleteProductById deletes a product by specified id and return its data.
	DeleteProductById(ctx context.Context, id uuid.UUID) (*domain.Product, error)

	// DeleteProductsByCategory deletes all products by specified category id and return their data.
	DeleteProductsByCategory(ctx context.Context, categoryId uuid.UUID) ([]domain.Product, error)

	// GetProductById fetches a single product by id.
	GetProductById(ctx context.Context, id uuid.UUID) (*domain.Product, error)

	// GetProducts fetches all products.
	GetProducts(ctx context.Context) ([]domain.Product, error)

	// GetProductsByCategory fetches products by category id.
	GetProductsByCategory(ctx context.Context, categoryId uuid.UUID) ([]domain.Product, error)
}

// ProductService is an interface for interacting with product business logic.
type ProductService interface {
	// AddCategory saves a new product category.
	AddCategory(ctx context.Context, name string) (*domain.ProductCategory, error)

	// UpdateCategory updates an existing category.
	UpdateCategory(ctx context.Context, dto *domain.UpdateCategoryProductDTO) error

	// DeleteCategory deletes a category by specified id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetProductCategories fetches all product categories.
	GetProductCategories(ctx context.Context) ([]domain.ProductCategory, error)

	// AddProduct saves a new product with linked image.
	AddProduct(ctx context.Context, dto *domain.AddProductDTO) (*domain.Product, error)

	// UpdateProduct updates an existing product.
	UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error

	// ReplaceProductImage sets a new image to a product.
	ReplaceProductImage(ctx context.Context, productId uuid.UUID, data io.Reader) (string, error)

	// DeleteProduct deletes a product with filters.
	DeleteProduct(ctx context.Context, dto *domain.DeleteProductDTO) error

	// GetProducts fetches products.
	GetProducts(ctx context.Context, dto *domain.GetProductsDTO) ([]domain.Product, error)
}
