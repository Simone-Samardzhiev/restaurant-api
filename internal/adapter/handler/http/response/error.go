package response

import (
	"net/http"
	"restaurant/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	StatusCode int      `json:"-"`
	Code       string   `json:"code"`
	Messages   []string `json:"messages"`
}

var errorMap = map[error]ErrorResponse{
	domain.ErrInternal: {
		StatusCode: http.StatusInternalServerError,
		Code:       "internal_error",
		Messages: []string{
			"Server cannot proceed the request.",
		},
	},
	domain.ErrProductNameAlreadyInUse: {
		StatusCode: http.StatusConflict,
		Code:       "product_name_already_exists",
		Messages: []string{
			"Product name is already in use.",
		},
	},
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	response, ok := errorMap[err]
	if !ok {
		return c.JSON(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Code:       "internal_error",
			Messages: []string{
				"Server cannot proceed the request.",
			},
		})
	}
	return c.JSON(response)
}
