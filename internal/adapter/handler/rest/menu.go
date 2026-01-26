package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductHandler handles menu related requests.
type ProductHandler struct {
	service           menu.Service
	imagesServingPath string
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service menu.Service, imagesServingPath string) *ProductHandler {
	return &ProductHandler{
		service:           service,
		imagesServingPath: imagesServingPath,
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

	category, err := h.service.AddCategory(ctx, menu.NewAddCategoryRequest(req.Name))

	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.JSON(http.StatusCreated, addCategoryResponse{
		Id:   category.Id,
		Name: category.Name.String(),
	})
}

type updateCategoryRequest struct {
	NewName *string `json:"newName,omitempty"`
}

func (h *ProductHandler) UpdateCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
		return
	}

	var req updateCategoryRequest
	if err = ctx.BindJSON(&req); err != nil {
		ctx.Error(domain.NewBadRequestError("invalid request body")).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.service.UpdateCategory(ctx, menu.NewCategoryUpdateRequest(parsedId, req.NewName)); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *ProductHandler) DeleteCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.service.DeleteCategory(ctx, parsedId); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

type getCategoriesResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (h *ProductHandler) GetCategories(ctx *gin.Context) {
	var categoryFilter menu.CategoryFilter

	if id, ok := ctx.GetQuery("id"); ok {
		paredId, err := uuid.Parse(id)
		if err != nil {
			ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
			return
		}
		categoryFilter.Id = &paredId
	}

	result, err := h.service.GetCategories(
		ctx,
		&categoryFilter,
	)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	response := make([]getCategoriesResponse, 0, len(result))
	for _, category := range result {
		response = append(response, getCategoriesResponse{
			Id:   category.Id,
			Name: category.Name.String(),
		})
	}

	ctx.JSON(http.StatusOK, response)
}

type productMetadata struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	Price       decimal.Decimal `json:"price"`
}

type addProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	Price       decimal.Decimal `json:"price"`
	ImageUrl    string          `json:"imageUrl"`
}

func (h *ProductHandler) AddProduct(ctx *gin.Context) {
	imageHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid image")).SetType(gin.ErrorTypeBind)
		return
	}

	imageFile, err := imageHeader.Open()
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid image")).SetType(gin.ErrorTypeBind)
		return
	}

	defer imageFile.Close()

	buffer := make([]byte, 512)
	_, err = io.ReadFull(imageFile, buffer)
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid image")).SetType(gin.ErrorTypeBind)
	}

	contentType := strings.Split(http.DetectContentType(buffer), "/")[1]

	if _, err := imageFile.Seek(0, io.SeekStart); err != nil {
		ctx.Error(domain.NewBadRequestError("invalid image")).SetType(gin.ErrorTypeBind)
		return
	}

	var metadata productMetadata
	if err = json.NewDecoder(strings.NewReader(ctx.PostForm("metadata"))).Decode(&metadata); err != nil {
		ctx.Error(domain.NewBadRequestError("invalid metadata")).SetType(gin.ErrorTypeBind)
		return
	}

	product, err := h.service.AddProduct(
		ctx,
		menu.NewAddProductRequest(
			metadata.Name,
			metadata.Description,
			metadata.CategoryId,
			metadata.Price,
			imageFile,
			contentType,
		),
	)

	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.JSON(http.StatusCreated, addProductResponse{
		Id:          product.Id,
		Name:        product.Name.String(),
		Description: product.Description.String(),
		CategoryId:  product.CategoryId,
		Price:       product.Price.Value(),
		ImageUrl:    path.Join(h.imagesServingPath, product.ImagePath),
	})
}

type updateProductRequest struct {
	NewName        *string          `json:"newName"`
	NewDescription *string          `json:"newDescription"`
	NewCategoryId  *uuid.UUID       `json:"newCategoryId"`
	NewPrice       *decimal.Decimal `json:"newPrice"`
}

func (h *ProductHandler) UpdateProduct(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
		return
	}

	var req updateProductRequest
	if err = ctx.BindJSON(&req); err != nil {
		ctx.Error(domain.NewBadRequestError("invalid request body")).SetType(gin.ErrorTypeBind)
	}

	if err = h.service.UpdateProduct(
		ctx,
		menu.NewUpdateProductRequest(id, req.NewName, req.NewDescription, req.NewCategoryId, req.NewPrice)); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

type replaceProductImageResponse struct {
	ImageUrl string `json:"imageUrl"`
}

func (h *ProductHandler) ReplaceProductImage(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
	}

	image := ctx.Request.Body
	buffer := make([]byte, 512)
	_, err = io.ReadFull(image, buffer)
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid image")).SetType(gin.ErrorTypeBind)
	}

	contentType := strings.Split(http.DetectContentType(buffer), "/")[1]
	fullBody := io.MultiReader(bytes.NewReader(buffer), image)

	newPath, err := h.service.ReplaceProductImage(ctx, menu.NewReplaceProductImageRequest(id, fullBody, contentType))
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.JSON(http.StatusOK, replaceProductImageResponse{
		ImageUrl: path.Join(h.imagesServingPath, newPath),
	})
}
