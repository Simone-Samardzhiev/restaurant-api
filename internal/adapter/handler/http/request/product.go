package request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AddProductCategoryRequest represents add category request body.
type AddProductCategoryRequest struct {
	Name string `json:"name" validate:"required,min=4,max=100"`
}

// UpdateCategoryRequest represents update category request body.
type UpdateCategoryRequest struct {
	NewName *string `json:"newName" validate:"omitempty,min=4,max=100"`
}

// AddProductRequest represents add product request body.
type AddProductRequest struct {
	Name        string          `json:"name" validate:"required,min=3,max=100"`
	Description string          `json:"description" validate:"required,min=15"`
	Category    uuid.UUID       `json:"category" validate:"required"`
	Price       decimal.Decimal `json:"price" validate:"required,gtZero"`
}

// UpdateProductRequest represents update product request body.
type UpdateProductRequest struct {
	NewName        *string          `json:"newName" validate:"omitempty,min=3,max=100"`
	NewDescription *string          `json:"newDescription" validate:"omitempty,min=15"`
	NewCategory    *uuid.UUID       `json:"newCategory" validate:"omitempty"`
	NewPrice       *decimal.Decimal `json:"newPrice" validate:"omitempty,gtZero"`
}
