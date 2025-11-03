package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DishCategory is an entity representing a dish.
type DishCategory struct {
	Id   uuid.UUID
	Name string
}

// NewDishCategory creates a new DishCategory instance.
func NewDishCategory(name string) *DishCategory {
	return &DishCategory{
		Id:   uuid.New(),
		Name: name,
	}
}

// Dish is an entity representing a dish.
type Dish struct {
	Id          uuid.UUID
	Name        string
	Description string
	ImageUrl    string
	Category    string
	Price       decimal.Decimal
}

// NewDish creates a new Dish instance.
func NewDish(name, description, imageUrl string, category string, price decimal.Decimal) *Dish {
	return &Dish{
		Id:          uuid.New(),
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
		Category:    category,
		Price:       price,
	}
}
