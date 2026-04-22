package rest

import (
	"menu/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	errors := make(map[string][]string)

	a.Name = strings.TrimSpace(a.Name)
	length := utf8.RuneCountInString(a.Name)

	if length <= MinCategoryLength {
		errors["name"] = append(errors["name"], "Must be at least "+strconv.Itoa(MinCategoryLength)+" characters.")
	}

	if length >= MaxCategoryLength {
		errors["name"] = append(errors["name"], "Must be at most "+strconv.Itoa(MaxCategoryLength)+" characters.")
	}

	return errors
}

// CategoryResponse represents the JSON response of a category
type CategoryResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AddCategory handles creation of a new category.
func (c *CategoryHandler) AddCategory(ctx *gin.Context) {
	var req AddCategoryRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(InvalidJSONErrorResponse.HTTPStatus, InvalidJSONErrorResponse)
		return
	}

	if errors := req.Validate(); len(errors) > 0 {
		handleValidationError(ctx, errors)
		return
	}

	category, err := c.service.Add(ctx, req.Name)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, CategoryResponse{
		Id:        category.Id,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	})
}
