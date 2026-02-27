package menu

import (
	"errors"
	"io"
	"restaurant/internal/domain"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Constants for [ProductName] length requirements.
const (
	minProductNameLength = 3
	maxProductNameLength = 100
)

// ProductName represents a valid product name.
type ProductName struct {
	raw string
}

func (p *ProductName) String() string {
	return p.raw
}

// ParseProductName parses a [ProductName] from string.
//
// If the name is invalid the returned error will be of type [domain.ErrorDetail].
func ParseProductName(name string) (ProductName, error) {
	name = strings.TrimSpace(name)
	length := utf8.RuneCountInString(name)

	if length < minProductNameLength {
		return ProductName{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeProductNameTooShort,
			Message: "product name too short",
			Metadata: map[string]any{
				"actual": length,
				"min":    minProductNameLength,
				"max":    maxProductNameLength,
			},
		}
	}

	if length > maxProductNameLength {
		return ProductName{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeProductNameTooLong,
			Message: "product name is too long.",
			Metadata: map[string]any{
				"actual": length,
				"min":    minProductNameLength,
				"max":    maxProductNameLength,
			},
		}
	}

	return ProductName{raw: name}, nil
}

// Constants for [ProductDescription] length requirements.
const (
	minProductDescriptionLength = 15
)

// ProductDescription represents a valid product description.
type ProductDescription struct {
	raw string
}

func (p *ProductDescription) String() string {
	return p.raw
}

// ParseProductDescription parses a [ProductDescription] from string.
//
// If the description is invalid the returned error will be of type [domain.ErrorDetail].
func ParseProductDescription(description string) (ProductDescription, error) {
	description = strings.TrimSpace(description)
	length := utf8.RuneCountInString(description)

	if length < minProductDescriptionLength {
		return ProductDescription{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeProductDescriptionTooShort,
			Message: "product description is too short",
			Metadata: map[string]any{
				"actual": length,
				"min":    minProductDescriptionLength,
			},
		}
	}
	return ProductDescription{raw: description}, nil
}

// ProductPrice represents a valid product price.
type ProductPrice struct {
	raw decimal.Decimal
}

func (p *ProductPrice) Value() decimal.Decimal {
	return p.raw
}

// ParseProductPrice parses a [ProductPrice] from [decimal.Decimal].
//
// If the price is invalid the returned error will be of type [domain.ErrorDetail].
func ParseProductPrice(price decimal.Decimal) (ProductPrice, error) {
	if price.LessThanOrEqual(decimal.Zero) {
		return ProductPrice{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeProductPriceLessThanZero,
			Message: "product price cannot be less than zero",
			Metadata: map[string]any{
				"actual": price,
				"min":    "0",
			},
		}
	}

	return ProductPrice{raw: price}, nil
}

// Product represent a valid product of the menu.
type Product struct {
	Id          uuid.UUID
	Name        ProductName
	Description ProductDescription
	Price       ProductPrice
	CategoryId  uuid.UUID
	ImagePath   string
}

// ParseProduct parses a [Product] from id, name, description, price, category id, imagePath.
//
// If the product is invalid the returned error will be of type [domain.Error].
func ParseProduct(
	id uuid.UUID,
	name,
	description string,
	price decimal.Decimal,
	categoryId uuid.UUID,
	imagePath string,
) (*Product, error) {
	errs := make([]domain.ErrorDetail, 0)

	parsedName, err := ParseProductName(name)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	parsedDescription, err := ParseProductDescription(description)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	parsedPrice, err := ParseProductPrice(price)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid product", domain.ErrorCodeInvalidProduct, errs...)
	}

	return &Product{
		Id:          id,
		Name:        parsedName,
		Description: parsedDescription,
		Price:       parsedPrice,
		CategoryId:  categoryId,
		ImagePath:   imagePath,
	}, nil
}

// SaveProductRequest represent a request for saving a product.
type SaveProductRequest struct {
	Name        ProductName
	Description ProductDescription
	Price       ProductPrice
	CategoryId  uuid.UUID
	ImagePath   string
}

// NewSaveProductRequest allocates and creates a new [SaveProductRequest].
func NewSaveProductRequest(
	name ProductName,
	description ProductDescription,
	price ProductPrice,
	categoryId uuid.UUID,
	imagePath string,
) *SaveProductRequest {
	return &SaveProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
		CategoryId:  categoryId,
		ImagePath:   imagePath,
	}
}

// AddProductRequest represents a request for adding a new product with image.
type AddProductRequest struct {
	Name        ProductName
	Description ProductDescription
	Price       ProductPrice
	CategoryId  uuid.UUID
	ImageData   io.Reader
	ImageType   ImageType
}

// ParseAddProductRequest parses [AddProductRequest] from name, description, price, category id, image data and image type.
//
// If the fields are invalid the returned error will be of type [domain.Error].
func ParseAddProductRequest(
	name,
	description string,
	price decimal.Decimal,
	categoryId uuid.UUID,
	imageData io.Reader,
	imageType string,
) (*AddProductRequest, error) {
	errs := make([]domain.ErrorDetail, 0)
	parsedName, err := ParseProductName(name)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	parsedDescription, err := ParseProductDescription(description)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	parsedPrice, err := ParseProductPrice(price)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	parsedImageType, err := ParseImageType(imageType)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid product", domain.ErrorCodeInvalidProduct, errs...)
	}

	return &AddProductRequest{
		Name:        parsedName,
		Description: parsedDescription,
		Price:       parsedPrice,
		CategoryId:  categoryId,
		ImageData:   imageData,
		ImageType:   parsedImageType,
	}, nil
}

// MustParseAddProductRequest is like [ParseAddProductRequest], but instead
// of returning the error it panics.
func MustParseAddProductRequest(
	name,
	description string,
	price decimal.Decimal,
	categoryId uuid.UUID,
	imageData io.Reader,
	imageType string,
) *AddProductRequest {
	request, err := ParseAddProductRequest(name, description, price, categoryId, imageData, imageType)
	if err != nil {
		panic(err)
	}
	return request
}

// UpdateProductRequest represents a request for updating a product.
type UpdateProductRequest struct {
	Id          uuid.UUID
	Name        *ProductName
	Description *ProductDescription
	Price       *ProductPrice
	CategoryId  *uuid.UUID
	ImagePath   *string
}

// ParseUpdateProductRequest parses an [UpdateProductRequest] from id, name, description, price, category id.
//
// If fields are invalid the returned error will be of type
func ParseUpdateProductRequest(
	id uuid.UUID,
	name,
	description *string,
	price *decimal.Decimal,
	categoryId *uuid.UUID,
	imagePath *string,
) (*UpdateProductRequest, error) {
	if name == nil && description == nil && price == nil && categoryId == nil && imagePath == nil {
		return nil, domain.NewBadRequestError(
			"update does not have data",
			domain.ErrorCodeNoData,
			nil,
		)
	}

	errs := make([]domain.ErrorDetail, 0)
	update := &UpdateProductRequest{
		Id:         id,
		CategoryId: categoryId,
		ImagePath:  imagePath,
	}

	if name != nil {
		parsedName, err := ParseProductName(*name)
		if err != nil {
			if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
				errs = append(errs, *detailErr)
			} else {
				return nil, err
			}
		}
		update.Name = &parsedName
	}

	if description != nil {
		parsedDescription, err := ParseProductDescription(*description)
		if err != nil {
			if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
				errs = append(errs, *detailErr)
			} else {
				return nil, err
			}
		}
		update.Description = &parsedDescription
	}

	if price != nil {
		parsedPrice, err := ParseProductPrice(*price)
		if err != nil {
			if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
				errs = append(errs, *detailErr)
			} else {
				return nil, err
			}
		}
		update.Price = &parsedPrice
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid product", domain.ErrorCodeInvalidProductUpdate, errs...)
	}

	return update, nil
}

// MustParseProductUpdateRequest is like [ParseUpdateProductRequest], but insted of
// returning the error it panics.
func MustParseProductUpdateRequest(
	id uuid.UUID,
	name,
	description *string,
	price *decimal.Decimal,
	categoryId *uuid.UUID,
	imagePath *string,
) *UpdateProductRequest {
	request, err := ParseUpdateProductRequest(id, name, description, price, categoryId, imagePath)
	if err != nil {
		panic(err)
	}
	return request
}
