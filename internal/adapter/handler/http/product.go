package http

import (
	"bytes"
	"io"
	"net/http"
	"restaurant/internal/adapter/handler/http/request"
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ProductHandler handler product-related HTTP requests.
type ProductHandler struct {
	productService port.ProductService
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(productService port.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) AddProductCategory(c *gin.Context) {
	var req request.AddProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	category, err := h.productService.AddCategory(c, req.Name)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusCreated, response.NewProductCategoryResponse(category.Id, req.Name))
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypePublic)
		return
	}

	var req request.UpdateCategoryRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.productService.UpdateCategory(c, domain.NewUpdateCategoryProductDTO(id, req.NewName)); err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.productService.DeleteCategory(c, id); err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.Status(fiber.StatusOK)
}

func (h *ProductHandler) GetProductCategories(c *gin.Context) {
	categories, err := h.productService.GetProductCategories(c)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	res := make([]response.ProductCategoryResponse, 0, len(categories))
	for _, category := range categories {
		res = append(res, response.NewProductCategoryResponse(category.Id, category.Name))
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProductHandler) AddProduct(c *gin.Context) {
	var req request.AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if uuid.IsInvalidLengthError(err) {
			c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
			return
		}

		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	product, err := h.productService.AddProduct(
		c,
		domain.NewAddProductDTO(
			req.Name,
			req.Description,
			req.Category,
			req.Price,
		),
	)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusCreated, response.NewProductResponse(product))
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)

	}
	var req request.UpdateProductRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	if err = h.productService.
		UpdateProduct(
			c,
			domain.NewUpdateProductDTO(
				id,
				req.NewName,
				req.NewDescription,
				req.NewCategory,
				req.NewPrice),
		); err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ProductHandler) ReplaceProductImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
		return
	}

	const maxImageSize = 5 << 20 // 5MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize)

	buf := make([]byte, 512)
	_, err = io.ReadAtLeast(c.Request.Body, buf, 512)
	if err != nil {
		c.Error(domain.NewBadRequestError(domain.InvalidImageFormat)).SetType(gin.ErrorTypePublic)
		return
	}

	content := http.DetectContentType(buf)
	if content != "image/jpeg" && content != "image/png" {
		c.Error(domain.NewBadRequestError(domain.InvalidImageFormat)).SetType(gin.ErrorTypeBind)
		return
	}

	c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf), c.Request.Body))

	url, err := h.productService.ReplaceProductImage(c, id, c.Request.Body)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusOK, response.NewUpdateImageResponse(url))
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	var (
		productID  *uuid.UUID
		categoryID *uuid.UUID
	)

	rawProductId, exist := c.GetQuery("product_id")
	if exist {
		id, err := uuid.Parse(rawProductId)
		if err != nil {
			c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
			return
		}
		productID = &id
	}
	rawCategoryId, exist := c.GetQuery("category_id")

	if exist {
		id, err := uuid.Parse(rawCategoryId)
		if err != nil {
			c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
			return
		}
		categoryID = &id
	}

	if err := h.productService.DeleteProduct(c, domain.NewDeleteProductDTO(productID, categoryID)); err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	c.Status(fiber.StatusOK)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	var (
		categoryID *uuid.UUID
	)

	rawCategoryId, exist := c.GetQuery("category_id")
	if exist {
		id, err := uuid.Parse(rawCategoryId)
		if err != nil {
			c.Error(domain.NewBadRequestError(domain.InvalidUUID)).SetType(gin.ErrorTypeBind)
			return
		}
		categoryID = &id
	}

	products, err := h.productService.GetProducts(c, domain.NewGetProductsDTO(categoryID))
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	res := make([]response.ProductResponse, 0, len(products))
	for _, product := range products {
		res = append(res, response.NewProductResponse(&product))
	}

	c.JSON(http.StatusOK, res)
}
