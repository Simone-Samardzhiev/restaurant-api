package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductStatus represents the statuses of [Product].
type ProductStatus string

const (
	// ProductStatusReady product is ready and can be displayed.
	ProductStatusReady ProductStatus = "product_ready"

	// ProductStatusAwaitingImage is not ready and waiting for image.
	ProductStatusAwaitingImage ProductStatus = "product_awaiting_image"
)

// Product represents product in the menu.
type Product struct {
	Id          uuid.UUID
	Name        string
	Description string
	Price       decimal.Decimal
	CategoryId  uuid.UUID

	ImageKey string
	Status   ProductStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}
