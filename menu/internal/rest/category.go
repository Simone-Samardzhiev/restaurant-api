package rest

import (
	"menu/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// CategoryHandler handles HTTP request for categories.
type CategoryHandler struct {
	service domain.CategoryService
}

// NewCategoryHandler creates and allocates new [CategoryHandler].
func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

const (
	// MinCategoryLength is the minimum length of the category name (included).
	MinCategoryLength = 3
	// MaxCategoryLength is the maximum length of category name (included).
	MaxCategoryLength = 64
)

// AddCategoryRequest represents the JSON request for creating a new category.
type AddCategoryRequest struct {
	Name string `json:"name"`
}

func (a *AddCategoryRequest) Validate() map[string][]string {
	fields := make(map[string][]string)

	a.Name = strings.TrimSpace(a.Name)
	length := utf8.RuneCountInString(a.Name)

	if length <= MinCategoryLength {
		fields["name"] = append(fields["name"], "Must be at least "+strconv.Itoa(MinCategoryLength)+" characters.")
	}

	if length >= MaxCategoryLength {
		fields["name"] = append(fields["name"], "Must be at most "+strconv.Itoa(MaxCategoryLength)+" characters.")
	}

	if len(fields) > 0 {
		return fields
	}

	return nil
}

// CategoryResponse represents the JSON response of a category
type CategoryResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AddCategory handles creation of a new category.
func (c *CategoryHandler) AddCategory(ctx *echo.Context) error {
	var req AddCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		return NewInvalidJSONError(err)
	}

	if fields := req.Validate(); fields != nil {
		return NewValidationError(fields)
	}

	result, err := c.service.Add(ctx.Request().Context(), req.Name)
	if err != nil {
		return NewErrorResponse(err)
	}

	return ctx.JSON(http.StatusCreated, CategoryResponse{
		Id:        result.Id,
		Name:      result.Name,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	})
}
