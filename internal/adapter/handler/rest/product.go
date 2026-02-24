package rest

import (
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

type AddProductRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	CategoryID  uuid.UUID       `json:"categoryId"`
}

func readProductRequest(ctx *gin.Context) (*AddProductRequest, error) {
	var req AddProductRequest
	if err := json.NewDecoder(strings.NewReader(ctx.PostForm("product"))).Decode(&req); err != nil {
		return nil, domain.NewBadRequestError("invalid product data", domain.ErrorCodeMalformedRequest, err)
	}

	return &req, nil
}

type AddImageRequest struct {
	ImageData io.ReadCloser
	ImageType string
}

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

type ProductResponse struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	CategoryId  uuid.UUID       `json:"categoryId"`
	ImageURL    string          `json:"imageUrl"`
}

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
