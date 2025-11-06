package http

import (
	"encoding/json"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"restaurant/internal/adapter/handler/http/request"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProductHandler handler product-related HTTP requests.
type ProductHandler struct {
	productService port.ProductService
	validator      *validator.Validate
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(productService port.ProductService, validator *validator.Validate) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      validator,
	}
}

// readProductData reads product data from form values and parses the json.
func readProductData(c *fiber.Ctx, validator *validator.Validate) (*request.AddProductRequest, error) {
	productData := c.FormValue("product")
	var productRequest request.AddProductRequest
	if err := json.Unmarshal([]byte(productData), &productRequest); err != nil {
		return nil, err
	}
	if err := validator.Struct(productRequest); err != nil {
		return nil, err
	}
	return &productRequest, nil
}

// readImageData reads product image from form value and validates its jpeg.
func readImageData(c *fiber.Ctx) (multipart.File, error) {
	imageData, err := c.FormFile("image")
	if err != nil {
		return nil, domain.ErrInvalidImageFormat
	}

	imageFile, err := imageData.Open()
	if err != nil {
		return nil, domain.ErrInternal
	}

	buffer := make([]byte, 512)
	_, _ = io.ReadAtLeast(imageFile, buffer, 512)

	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" {
		return nil, domain.ErrInvalidImageFormat
	}

	_, err = imageFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, domain.ErrInternal
	}

	_, err = jpeg.Decode(imageFile)
	if err != nil {
		return nil, domain.ErrInvalidImageFormat
	}

	_, err = imageFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return imageFile, nil
}

func (h *ProductHandler) AddProduct(c *fiber.Ctx) error {
	productRequest, err := readProductData(c, h.validator)
	if err != nil {
		return err
	}

	file, err := readImageData(c)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			zap.L().Error(
				"error closing image",
				zap.Error(closeErr),
			)
		}
	}()

	err = h.productService.AddProduct(
		c.Context(),
		domain.NewAddProductDTO(
			productRequest.Name,
			productRequest.Description,
			productRequest.Category,
			productRequest.Price,
			file,
		),
	)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *ProductHandler) AddProductCategory(c *fiber.Ctx) error {
	var req request.AddProductCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := h.validator.Struct(req); err != nil {
		return err
	}

	if err := h.productService.AddCategory(c.Context(), req.Name); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}

	var req request.UpdateCategoryRequest
	if err = c.BodyParser(&req); err != nil {
		return err
	}

	if err = h.validator.Struct(req); err != nil {
		return err
	}

	if err = h.productService.UpdateCategory(c.Context(), domain.NewUpdateCategoryProductDTO(id, req.NewName)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}

	if err = h.productService.DeleteCategory(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}
