package rest

import (
	"errors"
	"menu/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
)

// ProductHandler handles HTTP request for products.
type ProductHandler struct {
	baseImageUrl string
	service      domain.ProductService
}

// NewProductHandler creates and allocates new [ProductHandler].
func NewProductHandler(baseImageUlr string, service domain.ProductService) *ProductHandler {
	return &ProductHandler{
		baseImageUrl: baseImageUlr,
		service:      service,
	}
}

const (
	// MinProductNameLength is the minimum length of product name (included).
	MinProductNameLength = 4
	// MaxProductNameLength is the maximum length of the product name (included).
	MaxProductNameLength = 128
	// MinProductDescriptionLength is the minimum length of the product description.
	MinProductDescriptionLength = 16
)

func isValidContentType(imageContentType string) bool {
	switch domain.ImageContentType(imageContentType) {
	case domain.ImageContentTypeJPEG, domain.ImageContentTypePNG, domain.ImageContentTypeWebP:
		return true
	default:
		return false
	}
}

// AddProductRequest represents the JSON request for adding a new product.
type AddProductRequest struct {
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Price            decimal.Decimal `json:"price"`
	CategoryID       uuid.UUID       `json:"categoryId"`
	ImageContentType string          `json:"imageContentType"`
}

// ProductDraftResponse represents the JSON response of a product draft.
type ProductDraftResponse struct {
	Id             uuid.UUID `json:"id"`
	ImageUploadUrl string    `json:"imageUploadUrl"`
}

func (a *AddProductRequest) Validate() map[string][]string {
	fields := make(map[string][]string)

	a.Name = strings.TrimSpace(a.Name)
	length := utf8.RuneCountInString(a.Name)
	if length < MinProductNameLength {
		fields["name"] = append(fields["name"], "Must be at least "+strconv.Itoa(MinProductNameLength)+" characters.")
	}
	if length > MaxProductNameLength {
		fields["name"] = append(fields["name"], "Must be at most "+strconv.Itoa(MaxProductNameLength)+" characters.")
	}

	a.Description = strings.TrimSpace(a.Description)
	length = utf8.RuneCountInString(a.Description)
	if length < MinProductDescriptionLength {
		fields["description"] = append(fields["description"], "Must be at least "+strconv.Itoa(MinProductDescriptionLength)+" characters.")
	}

	if a.Price.LessThan(decimal.Zero) {
		fields["price"] = append(fields["price"], "Price must be greater than zero.")
	}

	a.ImageContentType = strings.TrimSpace(a.ImageContentType)
	a.ImageContentType = strings.ToLower(a.ImageContentType)
	if !isValidContentType(a.ImageContentType) {
		fields["imageContentType"] = append(fields["imageContentType"], "Invalid image content type.")
	}

	if len(fields) > 0 {
		return fields
	}

	return nil
}

func (p *ProductHandler) AddProduct(ctx *echo.Context) error {
	var req AddProductRequest
	if err := ctx.Bind(&req); err != nil {
		return NewInvalidJSONError(err)
	}

	if fields := req.Validate(); fields != nil {
		return NewValidationError(fields)
	}

	draft, err := p.service.Add(ctx.Request().Context(), &domain.AddProductRequest{
		Name:             req.Name,
		Description:      req.Description,
		Price:            req.Price,
		CategoryId:       req.CategoryID,
		ImageContentType: domain.ImageContentType(req.ImageContentType),
	})
	if err != nil {
		return NewError(err)
	}

	return ctx.JSON(http.StatusCreated, ProductDraftResponse{
		Id:             draft.Id,
		ImageUploadUrl: draft.ImageUploadUrl,
	})
}

func (p *ProductHandler) GetUploadInfo(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	draft, err := p.service.GetUploadInfo(ctx.Request().Context(), id)
	if err != nil {
		return NewError(err)
	}

	return ctx.JSON(http.StatusOK, ProductDraftResponse{
		Id:             draft.Id,
		ImageUploadUrl: draft.ImageUploadUrl,
	})
}

func (p *ProductHandler) ConfirmImageUpload(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	if err = p.service.ConfirmImageUpload(ctx.Request().Context(), id); err != nil {
		return NewError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ProductResponse represent the JSON response of a product.
type ProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	ImageUrl    string          `json:"imageUrl"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (p *ProductHandler) GetProduct(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	product, err := p.service.GetProduct(ctx.Request().Context(), id)
	if err != nil {
		return NewError(err)
	}

	return ctx.JSON(http.StatusOK, ProductResponse{
		Id:          product.Id,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CategoryId:  product.CategoryId,
		ImageUrl:    p.baseImageUrl + "/" + product.ImageKey,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	})
}

func (p *ProductHandler) GetAllProductsWithImage(ctx *echo.Context) error {
	products, err := p.service.GetAllWithImage(ctx.Request().Context())
	if err != nil {
		return NewError(err)
	}

	res := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		res = append(res, ProductResponse{
			Id:          product.Id,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			CategoryId:  product.CategoryId,
			ImageUrl:    p.baseImageUrl + "/" + product.ImageKey,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		})
	}

	ctx.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=3600, s-maxage=3600")
	return ctx.JSON(http.StatusOK, res)
}

func (p *ProductHandler) GetImage(ctx *echo.Context) error {
	key := ctx.Param("key")

	image, err := p.service.GetImage(ctx.Request().Context(), key)
	if err != nil {
		return NewError(err)
	}

	defer image.Close()

	ctx.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=604800, s-maxage=604800")
	return ctx.Stream(http.StatusOK, string(image.ContentType), image.Data)
}

// UpdateProductRequest represents the JSON request for updating the data of a product.
type UpdateProductRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Price       *decimal.Decimal `json:"price,omitempty"`
	CategoryID  *uuid.UUID       `json:"categoryId,omitempty"`
}

func (u *UpdateProductRequest) Validate() map[string][]string {
	fields := make(map[string][]string)

	if u.Name != nil {
		u.Name = new(strings.TrimSpace(*u.Name))
		length := utf8.RuneCountInString(*u.Name)
		if length < MinProductNameLength {
			fields["name"] = append(fields["name"], "Must be at least "+strconv.Itoa(MinProductNameLength)+" characters.")
		}
		if length > MaxProductNameLength {
			fields["name"] = append(fields["name"], "Must be at most "+strconv.Itoa(MaxProductNameLength)+" characters.")
		}
	}

	if u.Description != nil {
		u.Description = new(strings.TrimSpace(*u.Description))
		length := utf8.RuneCountInString(*u.Description)
		if length < MinProductDescriptionLength {
			fields["description"] = append(fields["description"], "Must be at least "+strconv.Itoa(MinProductDescriptionLength)+" characters.")
		}
	}

	if u.Price != nil {
		if u.Price.LessThan(decimal.Zero) {
			fields["price"] = append(fields["price"], "Price must be greater than zero.")
		}
	}

	if len(fields) > 0 {
		return fields
	}

	return nil
}

func (u *UpdateProductRequest) IsEmpty() bool {
	return u.Name == nil && u.Description == nil && u.Price == nil && u.CategoryID == nil
}

func (p *ProductHandler) UpdateProduct(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	var req UpdateProductRequest
	if err = ctx.Bind(&req); err != nil {
		return NewInvalidJSONError(err)
	}
	if req.IsEmpty() {
		return &Error{
			HttpStatus: http.StatusBadRequest,
			Code:       ErrorCodeEmptyRequest,
			Message:    "Request cannot be empty.",
			Err:        errors.New("cannot proceed an empty request"),
		}
	}

	if fields := req.Validate(); fields != nil {
		return NewValidationError(fields)
	}

	if err = p.service.UpdateProduct(ctx.Request().Context(), &domain.UpdateProductRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryId:  req.CategoryID,
	}); err != nil {
		return NewError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// MarkProductForImageUpdateRequest represents the JSON request for marking a product for image update.
type MarkProductForImageUpdateRequest struct {
	ImageContentType string `json:"imageContentType"`
}

func (m *MarkProductForImageUpdateRequest) Validate() map[string][]string {
	if !isValidContentType(m.ImageContentType) {
		return map[string][]string{
			"imageContentType": {"Invalid image content type."},
		}
	}

	return nil
}

func (p *ProductHandler) MarkProductForImageUpdate(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	var req MarkProductForImageUpdateRequest
	if err = ctx.Bind(&req); err != nil {
		return NewInvalidJSONError(err)
	}

	if fields := req.Validate(); fields != nil {
		return NewValidationError(fields)
	}

	if err = p.service.MarkProductForImageUpdate(
		ctx.Request().Context(),
		id,
		domain.ImageContentType(req.ImageContentType),
	); err != nil {
		return NewError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (p *ProductHandler) DeleteProduct(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}
	if err = p.service.Delete(ctx.Request().Context(), id); err != nil {
		return NewError(err)
	}
	return ctx.NoContent(http.StatusNoContent)
}
