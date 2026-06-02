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
	ProductStatusReady ProductStatus = "ready"

	// ProductStatusAwaitingImage is not ready and waiting for image.
	ProductStatusAwaitingImage ProductStatus = "awaiting_image"
)

// Product represents product in the menu.
type Product struct {
	Id          uuid.UUID
	Name        string
	Description string
	Price       decimal.Decimal
	CategoryId  uuid.UUID

	ImageKey         string
	ImageContentType ImageContentType
	Status           ProductStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}

// AddProductRequest represents a request for adding a new product.
type AddProductRequest struct {
	Name             string
	Description      string
	Price            decimal.Decimal
	CategoryId       uuid.UUID
	ImageContentType ImageContentType
}

// ProductDraft represent a product that has been saved, but image is needed to finalize it.
type ProductDraft struct {
	Id             uuid.UUID
	ImageUploadUrl string
}

// UpdateProductRequest represents a request for updating products data.
type UpdateProductRequest struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	Price       *decimal.Decimal
	CategoryId  *uuid.UUID
}
