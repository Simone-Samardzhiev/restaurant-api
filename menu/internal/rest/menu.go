package rest

import (
	"menu/internal/domain"
	"net/http"

	"github.com/labstack/echo/v5"
)

// MenuHandler handles HTTP request for menu.
type MenuHandler struct {
	baseImageUrl string
	service      domain.MenuService
}

// NewMenuHandler creates and allocates new [MenuHandler].
func NewMenuHandler(baseImageUrl string, service domain.MenuService) *MenuHandler {
	return &MenuHandler{baseImageUrl: baseImageUrl, service: service}
}

// MenuSectionResponse represents the JSON response of menu section.
type MenuSectionResponse struct {
	CategoryResponse
	Products []ProductResponse `json:"products"`
}

func (m *MenuHandler) GetMenu(ctx *echo.Context) error {
	menu, err := m.service.Get(ctx.Request().Context())
	if err != nil {
		return NewError(err)
	}

	res := make([]MenuSectionResponse, 0, len(menu))
	for _, section := range menu {
		s := MenuSectionResponse{
			CategoryResponse: CategoryResponse{
				Id:        section.Category.Id,
				Name:      section.Category.Name,
				CreatedAt: section.Category.CreatedAt,
				UpdatedAt: section.Category.UpdatedAt,
			},
			Products: make([]ProductResponse, 0, len(section.Products)),
		}
		for _, product := range section.Products {
			s.Products = append(s.Products, ProductResponse{
				Id:          product.Id,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				CategoryId:  product.CategoryId,
				ImageUrl:    m.baseImageUrl + "/" + product.ImageKey,
				CreatedAt:   product.CreatedAt,
				UpdatedAt:   product.UpdatedAt,
			})
		}
		res = append(res, s)
	}

	return ctx.JSON(http.StatusOK, res)
}
