package response

import (
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductCategoryResponse represents a product category response.
type ProductCategoryResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// NewProductCategoryResponse creates a new ProductCategoryResponse instance.
func NewProductCategoryResponse(id uuid.UUID, name string) ProductCategoryResponse {
	return ProductCategoryResponse{
		Id:   id,
		Name: name,
	}
}

// ProductResponse represents a product category response.
type ProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ImageUrl    *string         `json:"imageUrl"`
	Category    uuid.UUID       `json:"category"`
	Price       decimal.Decimal `json:"price"`
}

// NewProductResponse creates a new ProductResponse instance.
func NewProductResponse(product *domain.Product) ProductResponse {
	return ProductResponse{
		Id:          product.Id,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Price:       product.Price,
		ImageUrl:    product.ImageUrl,
	}
}

// UpdateImageResponse represents a response when updating the image of a product.
type UpdateImageResponse struct {
	Url string `json:"url"`
}

// NewUpdateImageResponse creates a new UpdateImageResponse instance.
func NewUpdateImageResponse(url string) UpdateImageResponse {
	return UpdateImageResponse{
		Url: url,
	}
}
