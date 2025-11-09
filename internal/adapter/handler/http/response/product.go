package response

import "github.com/google/uuid"

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
