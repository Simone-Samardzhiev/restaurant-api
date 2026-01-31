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

// ProductHandler handles product-related http requests.
type ProductHandler struct {
	service           menu.ProductService
	imagesServingPath string
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service menu.ProductService, imagesServingPath string) *ProductHandler {
	return &ProductHandler{
		service:           service,
		imagesServingPath: imagesServingPath,
	}
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

func (h *ProductHandler) DeleteProduct(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid")).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.service.DeleteProduct(ctx, id); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

type getProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	Price       decimal.Decimal `json:"price"`
	ImageUrl    string          `json:"imageUrl"`
}

func (h *ProductHandler) GetProducts(ctx *gin.Context) {
	products, err := h.service.GetProducts(ctx)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	response := make([]getProductResponse, 0, len(products))
	for _, product := range products {
		response = append(response, getProductResponse{
			Id:          product.Id,
			Name:        product.Name.String(),
			Description: product.Description.String(),
			CategoryId:  product.CategoryId,
			Price:       product.Price.Value(),
			ImageUrl:    path.Join(h.imagesServingPath, product.ImagePath),
		})
	}

	ctx.JSON(http.StatusOK, response)
}
