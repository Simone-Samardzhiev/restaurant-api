package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductCategory is an entity representing a dish.
type ProductCategory struct {
	Id   uuid.UUID
	Name string
}

// NewProductCategory creates a new ProductCategory instance.
func NewProductCategory(name string) *ProductCategory {
	return &ProductCategory{
		Id:   uuid.New(),
		Name: name,
	}
}

// Product is an entity representing a dish.
type Product struct {
	Id          uuid.UUID
	Name        string
	Description string
	ImageUrl    string
	Category    string
	Price       decimal.Decimal
}

// NewDish creates a new Product instance.
func NewDish(name, description, imageUrl string, category string, price decimal.Decimal) *Product {
	return &Product{
		Id:          uuid.New(),
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
		Category:    category,
		Price:       price,
	}
}
