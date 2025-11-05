package domain

import (
	"io"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductCategory is an entity representing a product.
type ProductCategory struct {
	Id   uuid.UUID
	Name string
}

// NewProductCategory creates a new ProductCategory instance.
func NewProductCategory(id uuid.UUID, name string) *ProductCategory {
	return &ProductCategory{
		Id:   id,
		Name: name,
	}
}

// Product is an entity representing a product.
type Product struct {
	Id          uuid.UUID
	Name        string
	Description string
	ImagePath   string
	Category    uuid.UUID
	Price       decimal.Decimal
}

// NewProduct creates a new Product instance.
func NewProduct(id uuid.UUID, name, description, imagePath string, category uuid.UUID, price decimal.Decimal) *Product {
	return &Product{
		Id:          id,
		Name:        name,
		Description: description,
		ImagePath:   imagePath,
		Category:    category,
		Price:       price,
	}
}

// AddProductDTO is a DTO for adding a product.
type AddProductDTO struct {
	Name        string
	Description string
	Category    uuid.UUID
	Price       decimal.Decimal
	Image       io.Reader
}

// NewAddProductDTO creates a new AddProductDTO instance.
func NewAddProductDTO(name, description string, category uuid.UUID, price decimal.Decimal, imageFile io.Reader) *AddProductDTO {
	return &AddProductDTO{
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
		Image:       imageFile,
	}
}
