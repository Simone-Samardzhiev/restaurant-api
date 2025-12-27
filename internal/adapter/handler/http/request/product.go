package request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AddProductCategoryRequest represents add category request body.
type AddProductCategoryRequest struct {
	Name string `json:"name" binding:"required,min=4,max=100"`
}

// UpdateCategoryRequest represents update category request body.
type UpdateCategoryRequest struct {
	NewName *string `json:"newName" binding:"omitempty,min=4,max=100"`
}

// AddProductRequest represents add product request body.
type AddProductRequest struct {
	Name        string          `json:"name" binding:"required,min=3,max=100"`
	Description string          `json:"description" binding:"required,min=15"`
	Category    uuid.UUID       `json:"category" binding:"required"`
	Price       decimal.Decimal `json:"price" binding:"required,gtZero"`
}

// UpdateProductRequest represents update product request body.
type UpdateProductRequest struct {
	NewName        *string          `json:"newName" binding:"omitempty,min=3,max=100"`
	NewDescription *string          `json:"newDescription" binding:"omitempty,min=15"`
	NewCategory    *uuid.UUID       `json:"newCategory" binding:"omitempty"`
	NewPrice       *decimal.Decimal `json:"newPrice" binding:"omitempty,gtZero"`
}
