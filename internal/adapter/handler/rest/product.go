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

// ProductHandler handles product related http requests.
type ProductHandler struct {
	service          menu.ProductService
	imageServingPath string
}

// NewProductHandler allocates and creates a new [ProductHandler].
func NewProductHandler(service menu.ProductService, imageServingPath string) *ProductHandler {
	return &ProductHandler{
		service:          service,
		imageServingPath: imageServingPath,
	}
}

// AddProductRequest represent the JSON request for adding a product.
type AddProductRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	CategoryID  uuid.UUID       `json:"categoryId"`
}

// readProductRequest reads [AddProductRequest] from multipart data.
func readProductRequest(ctx *gin.Context) (*AddProductRequest, error) {
	var req AddProductRequest
	if err := json.NewDecoder(strings.NewReader(ctx.PostForm("product"))).Decode(&req); err != nil {
		return nil, domain.NewBadRequestError("invalid product data", domain.ErrorCodeMalformedRequest, err)
	}

	return &req, nil
}

// AddImageRequest represents the request for adding product image.
type AddImageRequest struct {
	ImageData io.ReadCloser
	ImageType string
}

// getImageType returns the string representation of the image type
// using [http.DetectContentType]. The default value is "application/octet-stream".
func getImageType(image io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)
	_, err := image.Read(buffer)
	if err != nil {
		return "", err
	}

	if _, err = image.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	imageType := strings.Split(http.DetectContentType(buffer), "/")[1]
	return imageType, nil
}

// readImageImage reads [AddImageRequest] from multipart data.
func readImageImage(ctx *gin.Context) (*AddImageRequest, error) {
	file, err := ctx.FormFile("image")
	if err != nil {
		return nil, domain.NewBadRequestError("invalid image data", domain.ErrorCodeMalformedRequest, err)
	}

	reader, err := file.Open()
	if err != nil {
		return nil, domain.NewBadRequestError("invalid image data", domain.ErrorCodeMalformedRequest, err)
	}

	imageType, err := getImageType(reader)
	if err != nil {
		return nil, domain.NewBadRequestError("invalid image data", domain.ErrorCodeMalformedRequest, err)
	}

	return &AddImageRequest{
		ImageData: reader,
		ImageType: imageType,
	}, nil
}

// ProductResponse represents JSON response of a product.
type ProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	ImageURL    string          `json:"imageUrl"`
}

// AddProduct decodes both [AddProductRequest] and [AddImageRequest], and attempts to add the product.
// If the product is added successfully the response is [ProductResponse].
func (h *ProductHandler) AddProduct(ctx *gin.Context) {
	productReq, err := readProductRequest(ctx)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	imageReq, err := readImageImage(ctx)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	domainRequest, err := menu.ParseAddProductRequest(
		productReq.Name,
		productReq.Description,
		productReq.Price,
		productReq.CategoryID,
		imageReq.ImageData,
		imageReq.ImageType,
	)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	product, err := h.service.AddProduct(ctx, domainRequest)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}
	ctx.JSON(http.StatusCreated, ProductResponse{
		Id:          product.Id,
		Name:        product.Name.String(),
		Description: product.Description.String(),
		Price:       product.Price.Value(),
		CategoryId:  product.CategoryId,
		ImageURL:    path.Join(h.imageServingPath, product.ImagePath),
	})
}

// UpdateProductRequest represents the JSON request for updating a product.
type UpdateProductRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Price       *decimal.Decimal `json:"price,omitempty"`
	CategoryID  *uuid.UUID       `json:"categoryId,omitempty"`
}

// UpdateProduct decodes [UpdateProductRequest] and attempts to update the product.
// If the product is updated successfully the response is [http.StatusNoContent].
func (h *ProductHandler) UpdateProduct(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidUUID, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	var req UpdateProductRequest
	if err = ctx.BindJSON(&req); err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid json", domain.ErrorCodeMalformedRequest, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	domainReq, err := menu.ParseUpdateProductRequest(id, req.Name, req.Description, req.Price, req.CategoryID)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	if err = h.service.UpdateProduct(ctx, domainReq); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// UpdateImageResponse represents JSON response when updating product image.
type UpdateImageResponse struct {
	ImageURL string `json:"imageUrl"`
}

// UpdateImage updates the image of the product, by the id of the product.
// If the image is updates successfully the response is [UpdateImageResponse].
func (h *ProductHandler) UpdateImage(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidUUID, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	buffer := make([]byte, 512)
	if _, err = ctx.Request.Body.Read(buffer); err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid image data", domain.ErrorCodeMalformedRequest, err),
		).SetType(gin.ErrorTypePublic)
		return
	}
	imageType := strings.Split(http.DetectContentType(buffer), "/")[1]

	domainRequest, err := menu.ParseUpdateImageRequest(id, io.MultiReader(bytes.NewReader(buffer), ctx.Request.Body), imageType)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	imagePath, err := h.service.UpdateImage(ctx, domainRequest)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}
	ctx.JSON(http.StatusCreated, UpdateImageResponse{
		ImageURL: path.Join(h.imageServingPath, imagePath),
	})
}

// DeleteProduct parses the id from the path parameter "id"
// and attempts to delete a product.
// If the product is deleted successfully the response is [http.StatusOK].
func (h *ProductHandler) DeleteProduct(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(
			domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidUUID, err),
		).SetType(gin.ErrorTypePublic)
		return
	}

	if err = h.service.DeleteProduct(ctx, id); err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
	}
	ctx.Status(http.StatusOK)
}
