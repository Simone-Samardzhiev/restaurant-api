package rest

import (
	"net/http"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CategoryHandler handles category related http requests.
type CategoryHandler struct {
	service menu.CategoryService
}

// AddCategoryRequest represents JSON request for adding a category.
type AddCategoryRequest struct {
	Name string `json:"name"`
}

// AddCategoryResponse represents JSON response for successfully adding a category.
type AddCategoryResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AddCategory decodes [AddCategoryRequest] and attempts to add the category.
// If the category is added successfully the response is [AddCategoryResponse].
func (h *CategoryHandler) AddCategory(ctx *gin.Context) {
	var req AddCategoryRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(
			domain.NewBadRequestError(
				"invalid uuid", domain.ErrorCodeInvalidUUID, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	domainRequest, err := menu.ParseAddCategoryRequest(req.Name)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	result, err := h.service.AddCategory(ctx, domainRequest)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.JSON(http.StatusCreated, AddCategoryResponse{
		Id:   result.Id,
		Name: result.Name.String(),
	})
}

// NewCategoryHandler allocates and returns a new CategoryHandler.
func NewCategoryHandler(service menu.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}
