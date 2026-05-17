package rest

import (
	"menu/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
)

type ProductHandler struct {
	service domain.ProductService
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
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
	switch imageContentType {
	case domain.ImageContentTypeJPEG, domain.ImageContentTypePNG, domain.ImageContentTypeWebP:
		return true
	default:
		return false
	}
}

type AddProductRequest struct {
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Price            decimal.Decimal `json:"price"`
	CategoryID       uuid.UUID       `json:"categoryId"`
	ImageContentType string          `json:"imageContentType"`
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

func (p *ProductHandler) Add(ctx *echo.Context) error {
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

	return ctx.JSON(http.StatusCreated, draft)
}
