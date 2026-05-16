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

	if length < MinCategoryLength {
		fields["name"] = append(fields["name"], "Must be at least "+strconv.Itoa(MinCategoryLength)+" characters.")
	}
	if length > MaxCategoryLength {
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
		return NewError(err)
	}

	return ctx.JSON(http.StatusCreated, CategoryResponse{
		Id:        result.Id,
		Name:      result.Name,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	})
}

func (c *CategoryHandler) GetCategories(ctx *echo.Context) error {
	categories, err := c.service.GetAll(ctx.Request().Context())
	if err != nil {
		return NewError(err)
	}

	response := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, CategoryResponse{
			Id:        category.Id,
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		})
	}

	ctx.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=60")
	ctx.Response().Header().Set("Surrogate-Control", "public, max-age=3600")

	return ctx.JSON(http.StatusOK, response)
}

// UpdateCategoryRequest represents the JSON request for updating a category.
type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

func (u *UpdateCategoryRequest) Validate() map[string][]string {
	fields := make(map[string][]string)
	u.Name = strings.TrimSpace(u.Name)
	length := utf8.RuneCountInString(u.Name)

	if length < MinCategoryLength {
		fields["name"] = append(fields["name"], "Must be at least "+strconv.Itoa(MinCategoryLength)+" characters.")
	}
	if length > MaxCategoryLength {
		fields["name"] = append(fields["name"], "Must be at most "+strconv.Itoa(MaxCategoryLength)+" characters.")
	}

	if len(fields) > 0 {
		return fields
	}
	return nil
}

func (c *CategoryHandler) UpdateCategory(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	var req UpdateCategoryRequest
	if err = ctx.Bind(&req); err != nil {
		return NewInvalidJSONError(err)
	}

	if fields := req.Validate(); fields != nil {
		return NewValidationError(fields)
	}

	if err = c.service.Update(ctx.Request().Context(), id, req.Name); err != nil {
		return NewError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (c *CategoryHandler) DeleteCategory(ctx *echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return NewInvalidUUIDError(err)
	}

	if err = c.service.Delete(ctx.Request().Context(), id); err != nil {
		return NewError(err)
	}

	return ctx.NoContent(http.StatusOK)
}
