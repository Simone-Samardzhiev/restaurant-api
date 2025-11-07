package http

import (
	"bytes"
	"net/http"
	"restaurant/internal/adapter/handler/http/request"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func (h *ProductHandler) AddProduct(c *fiber.Ctx) error {
	var req request.AddProductRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := h.validator.Struct(req); err != nil {
		return err
	}

	if err := h.productService.AddProduct(
		c.Context(),
		domain.NewAddProductDTO(
			req.Name,
			req.Description,
			req.Category,
			req.Price,
		),
	); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}
	var req request.UpdateProductRequest
	if err = c.BodyParser(&req); err != nil {
		return err
	}
	if err = h.validator.Struct(req); err != nil {
		return err
	}

	if err = h.productService.
		UpdateProduct(
			c.Context(),
			domain.NewUpdateProductDTO(
				id,
				req.NewName,
				req.NewDescription,
				req.NewCategory,
				req.NewPrice),
		); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func validateImageFormat(c *fiber.Ctx) error {
	contentType := c.Get("content-type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return domain.ErrInvalidImageFormat
	}

	imageData := c.Body()
	if len(imageData) < 512 {
		return domain.ErrInvalidImageFormat
	}

	buffer := imageData[:512]
	contentType = http.DetectContentType(buffer)

	if contentType != "image/jpeg" && contentType != "image/png" {
		return domain.ErrInvalidImageFormat
	}
	return nil
}

func (h *ProductHandler) AddImage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.ErrInvalidUUID
	}

	if err = validateImageFormat(c); err != nil {
		return err
	}

	if err = h.productService.AddImage(
		c.Context(),
		bytes.NewReader(c.Body()),
		id,
	); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}
