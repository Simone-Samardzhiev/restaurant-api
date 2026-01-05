package rest

import (
	"net/http"
	"restaurant/internal/domain"
	"restaurant/internal/domain/product"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles product related requests.
type ProductHandler struct {
	service product.Service
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service product.Service) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

type addCategoryRequest struct {
	Name string `json:"name"`
}

type addCategoryResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AddCategory is a gin handler for adding a new category.
func (h *ProductHandler) AddCategory(ctx *gin.Context) {
	var req addCategoryRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(domain.NewBadRequestError("invalid request body")).SetType(gin.ErrorTypeBind)
		return
	}

	category, err := h.service.AddCategory(ctx, product.NewAddCategoryRequest(req.Name))

	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.JSON(http.StatusCreated, addCategoryResponse{
		Id:   category.Id,
		Name: category.RawName(),
	})
}
