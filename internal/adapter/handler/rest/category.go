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
			domain.NewBadRequestError("invalid json", domain.ErrorCodeMalformedRequest, err),
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

// UpdateCategoryRequest represents JSON request for updating a category.
type UpdateCategoryRequest struct {
	Name *string `json:"name"`
}

// UpdateCategory decodes [UpdateCategoryRequest] and attempts to update the category.
// If the category is updates successfully the response is [http.StatusNoContent].
func (h *CategoryHandler) UpdateCategory(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidCategory, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	var req UpdateCategoryRequest
	if err = ctx.BindJSON(&req); err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid json", domain.ErrorCodeMalformedRequest, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	domainRequest, err := menu.ParseUpdateCategoryRequest(id, req.Name)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	if err = h.service.UpdateCategory(ctx, domainRequest); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// DeleteCategory parsed the id from the path parameter and attempts to delete it.
// If the category is deleted successfully the response is [http.StatusOK].
func (h *CategoryHandler) DeleteCategory(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidCategory, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	if err = h.service.DeleteCategory(ctx, id); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusOK)
}

// NewCategoryHandler allocates and returns a new CategoryHandler.
func NewCategoryHandler(service menu.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}
