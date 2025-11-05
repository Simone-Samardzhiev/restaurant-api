package request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AddProductRequest represent product request body.
type AddProductRequest struct {
	Name        string          `json:"name" validate:"required,min=3,max=100"`
	Description string          `json:"description" validate:"required,min=15"`
	Category    uuid.UUID       `json:"category" validate:"required,uuid"`
	Price       decimal.Decimal `json:"price" validate:"required,gtZero"`
}
