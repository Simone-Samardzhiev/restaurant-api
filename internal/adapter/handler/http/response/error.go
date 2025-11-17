package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"restaurant/internal/core/domain"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorResponse represents a API error response.
type ErrorResponse struct {
	StatusCode int      `json:"-"`
	Code       string   `json:"code"`
	Messages   []string `json:"messages"`
}

// mapValidationErrors maps validator.ValidationErrors into ErrorResponse.
func mapValidationErrors(err validator.ValidationErrors) ErrorResponse {
	messages := make([]string, 0, len(err))

	for _, e := range err {
		messages = append(messages, fmt.Sprintf("%s failed on %s", e.Field(), e.Tag()))
	}

	return ErrorResponse{
		StatusCode: fiber.StatusUnprocessableEntity,
		Code:       "validation_error",
		Messages:   messages,
	}
}

// mapFiberError maps fiber.Error into ErrorResponse.
func mapFiberError(fiberErr *fiber.Error) ErrorResponse {
	if errors.Is(fiberErr, fiber.ErrBadRequest) {
		return ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Code:       "bad_request",
			Messages: []string{
				"Invalid request",
			},
		}
	} else {
		return ErrorResponse{
			StatusCode: fiber.StatusInternalServerError,
			Code:       "internal_error",
			Messages: []string{
				"Server cannot proceed the request",
			},
		}
	}
}

// mapUnmarshalJsonError maps json.UnmarshalTypeError into ErrorResponse.
func mapUnmarshalJsonError(jsonErr *json.UnmarshalTypeError) ErrorResponse {
	return ErrorResponse{
		StatusCode: fiber.StatusUnprocessableEntity,
		Code:       "bad_request",
		Messages: []string{
			fmt.Sprintf("Invalid value %s for field %s", jsonErr.Value, jsonErr.Field),
		},
	}
}

// mapJsonSyntaxError maps json.SyntaxError into ErrorResponse.
func mapJsonSyntaxError(jsonErr *json.SyntaxError) ErrorResponse {
	return ErrorResponse{
		StatusCode: fiber.StatusUnprocessableEntity,
		Code:       "bad_request",
		Messages: []string{
			fmt.Sprintf("Error parsing JSON near %d", jsonErr.Offset),
		},
	}
}

// domainErrorsMap is map used to map domain errors to ErrorResponse.
var domainErrorsMap = map[error]ErrorResponse{
	domain.ErrInternal: {
		StatusCode: fiber.StatusInternalServerError,
		Code:       "internal_error",
		Messages: []string{
			"Server cannot proceed the request",
		},
	},
	domain.ErrNothingToUpdate: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "nothing_to_update",
		Messages: []string{
			"Update request won't change any data.",
		},
	},
	domain.ErrNothingToDelete: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "nothing_to_delete",
		Messages: []string{
			"Delete request won't change any data.",
		},
	},
	domain.ErrNothingToFetch: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "nothing_to_fetch",
		Messages: []string{
			"Fetch request won't change any data.",
		},
	},
	domain.ErrInvalidUUID: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "invalid_uuid",
		Messages: []string{
			"Invalid uuid",
		},
	},
	domain.ErrInvalidImageFormat: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "invalid_image_format",
		Messages: []string{
			"Invalid image format",
			"Supported formats are jpeg and png",
		},
	},
	domain.ErrProductCategoryNameAlreadyInUse: {
		StatusCode: fiber.StatusConflict,
		Code:       "product_category_name_already_exists",
		Messages: []string{
			"Product category is already in use",
		},
	},
	domain.ErrProductCategoryNotFound: {
		StatusCode: fiber.StatusNotFound,
		Code:       "product_category_not_found",
		Messages: []string{
			"Product category not found",
		},
	},
	domain.ErrProductNameAlreadyInUse: {
		StatusCode: fiber.StatusConflict,
		Code:       "product_name_already_exists",
		Messages: []string{
			"Product name is already in use",
		},
	},
	domain.ErrProductNotFound: {
		StatusCode: fiber.StatusNotFound,
		Code:       "product_not_found",
		Messages: []string{
			"Product not found",
		},
	},
	domain.ErrCategoryHasLinkedProducts: {
		StatusCode: fiber.StatusBadRequest,
		Code:       "category_has_linked_products",
		Messages: []string{
			"Product category has linked products",
			"Delete the products first",
		},
	},
}

// mapDomainError maps domain errors into ErrorResponse.
func mapDomainError(err error) ErrorResponse {
	response, ok := domainErrorsMap[err]
	if !ok {
		zap.L().Error("unknown error", zap.Error(err))
		response = ErrorResponse{
			StatusCode: fiber.StatusInternalServerError,
			Code:       "internal_error",
			Messages: []string{
				"Server cannot proceed the request",
			},
		}
	}
	return response
}

// ErrorHandler is a handler used to handle all returned errors.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var validatorErr validator.ValidationErrors
	var fiberErr *fiber.Error
	var jsonErr *json.UnmarshalTypeError
	var jsonSyntaxErr *json.SyntaxError
	var response ErrorResponse

	switch {
	case errors.As(err, &validatorErr):
		response = mapValidationErrors(validatorErr)
	case errors.As(err, &fiberErr):
		response = mapFiberError(fiberErr)
	case errors.As(err, &jsonErr):
		response = mapUnmarshalJsonError(jsonErr)
	case errors.As(err, &jsonSyntaxErr):
		response = mapJsonSyntaxError(jsonSyntaxErr)
	default:
		response = mapDomainError(err)
	}

	return c.Status(response.StatusCode).JSON(response)
}
