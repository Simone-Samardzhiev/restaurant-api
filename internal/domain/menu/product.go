package menu

import (
	"fmt"
	"io"
	"restaurant/internal/domain"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductName represents a valid product name.
type ProductName struct {
	raw string
}

const (
	minProductNameLen = 3
	maxProductNameLen = 100
)

// NewProductName parses ProductName from string.
func NewProductName(name string) (ProductName, error) {
	name = strings.TrimSpace(name)
	count := utf8.RuneCountInString(name)

	if count < minProductNameLen {
		return ProductName{}, fmt.Errorf("product name must be at least %d characters", minProductNameLen)
	}
	if count > maxProductNameLen {
		return ProductName{}, fmt.Errorf("product name must be at most %d characters", maxProductNameLen)
	}

	return ProductName{raw: name}, nil
}

func (n *ProductName) String() string {
	return n.raw
}

// ProductDescription represents a valid product description.
type ProductDescription struct {
	raw string
}

const (
	minProductDescriptionLen = 15
)

// NewProductDescription parses a ProductDescription from string.
func NewProductDescription(description string) (ProductDescription, error) {
	description = strings.TrimSpace(description)
	count := utf8.RuneCountInString(description)
	if count < minProductDescriptionLen {
		return ProductDescription{}, fmt.Errorf("product description must be at least %d characters", minProductDescriptionLen)
	}

	return ProductDescription{raw: description}, nil
}

func (d *ProductDescription) String() string {
	return d.raw
}

// ProductPrice represents a valid product price.
type ProductPrice struct {
	raw decimal.Decimal
}

// NewProductPrice parses a ProductPrice from decimal.Decimal.
func NewProductPrice(price decimal.Decimal) (ProductPrice, error) {
	if price.LessThanOrEqual(decimal.Zero) {
		return ProductPrice{}, fmt.Errorf("price must be greater than zero")
	}
	return ProductPrice{raw: price}, nil
}

func (p *ProductPrice) Value() decimal.Decimal {
	return p.raw
}

func (p *ProductPrice) String() string {
	return p.raw.String()
}

// AddProductRequest represents a request for adding a new product.
type AddProductRequest struct {
	Name        string
	Description string
	CategoryId  uuid.UUID
	Price       decimal.Decimal
	ImageData   io.Reader
	ImageType   string
}

// NewAddProductRequest creates a new AddProductRequest.
func NewAddProductRequest(
	name, description string,
	categoryId uuid.UUID,
	price decimal.Decimal,
	imageData io.Reader,
	imageType string,
) *AddProductRequest {
	return &AddProductRequest{
		Name:        name,
		Description: description,
		CategoryId:  categoryId,
		Price:       price,
		ImageData:   imageData,
		ImageType:   imageType,
	}
}

// Product represent a product entity.
type Product struct {
	Id          uuid.UUID
	Name        ProductName
	Description ProductDescription
	CategoryId  uuid.UUID
	Price       ProductPrice
	ImagePath   string
}

// NewProduct creates a new Product by passing all the fields.
func NewProduct(
	id uuid.UUID,
	name, description string,
	categoryId uuid.UUID,
	price decimal.Decimal,
	imagePath string,
) (*Product, error) {
	validationErrors := domain.NewValidationErrors("invalid product")

	parsedName, err := NewProductName(name)
	if err != nil {
		validationErrors.Add("name", err)
	}

	parsedDescription, err := NewProductDescription(description)
	if err != nil {
		validationErrors.Add("description", err)
	}

	parsedPrice, err := NewProductPrice(price)
	if err != nil {
		validationErrors.Add("price", err)
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	return &Product{
		Id:          id,
		Name:        parsedName,
		Description: parsedDescription,
		CategoryId:  categoryId,
		Price:       parsedPrice,
		ImagePath:   imagePath,
	}, nil
}

// NewProductWithValidFields creates a new Product with already parsed fields.
func NewProductWithValidFields(
	id uuid.UUID,
	name ProductName,
	description ProductDescription,
	categoryId uuid.UUID,
	price ProductPrice,
	imagePath string,
) *Product {
	return &Product{
		Id:          id,
		Name:        name,
		Description: description,
		CategoryId:  categoryId,
		Price:       price,
		ImagePath:   imagePath,
	}
}

// UpdateProductRequest represents a request for updating an existing product..
type UpdateProductRequest struct {
	Id             uuid.UUID
	NewName        *string
	NewDescription *string
	NewCategoryId  *uuid.UUID
	NewPrice       *decimal.Decimal
}

// NewUpdateProductRequest creates a new UpdateProductRequest.
func NewUpdateProductRequest(
	id uuid.UUID,
	newName, newDescription *string,
	newCategoryId *uuid.UUID,
	newPrice *decimal.Decimal,
) *UpdateProductRequest {
	return &UpdateProductRequest{
		Id:             id,
		NewName:        newName,
		NewDescription: newDescription,
		NewCategoryId:  newCategoryId,
		NewPrice:       newPrice,
	}
}

// ProductUpdate represents a product update.
type ProductUpdate struct {
	Id             uuid.UUID
	NewName        *ProductName
	NewDescription *ProductDescription
	NewCategoryId  *uuid.UUID
	NewPrice       *ProductPrice
}

// NewProductUpdate creates a new ProductUpdate by parsing all the fields and validate at least one field is provided.
func NewProductUpdate(
	id uuid.UUID,
	newName, newDescription *string,
	newCategoryId *uuid.UUID,
	newPrice *decimal.Decimal,
) (*ProductUpdate, error) {
	validationErrors := domain.NewValidationErrors("invalid product update")
	hasData := false

	var parsedName *ProductName
	if newName != nil {
		val, err := NewProductName(*newName)

		if err != nil {
			validationErrors.Add("name", err)
		} else {
			parsedName = &val
			hasData = true
		}
	}

	var parsedDescription *ProductDescription
	if newDescription != nil {
		val, err := NewProductDescription(*newDescription)
		if err != nil {
			validationErrors.Add("newDescription", err)
		} else {
			parsedDescription = &val
			hasData = true
		}
	}

	var parsedPrice *ProductPrice
	if newPrice != nil {
		val, err := NewProductPrice(*newPrice)
		if err != nil {
			validationErrors.Add("newPrice", err)
		} else {
			parsedPrice = &val
			hasData = true
		}
	}

	if !hasData {
		return nil, domain.NewBadRequestError("product update does not have data")
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	return &ProductUpdate{
		Id:             id,
		NewName:        parsedName,
		NewDescription: parsedDescription,
		NewCategoryId:  newCategoryId,
		NewPrice:       parsedPrice,
	}, nil
}
