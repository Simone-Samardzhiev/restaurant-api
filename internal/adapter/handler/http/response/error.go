package response

import (
	"errors"
	"restaurant/internal/core/domain"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ErrorResponse represent the standard JSON representation of the API error.
type ErrorResponse struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages"`
}

// NewErrorResponse creates a new ErrorResponse from error.
func NewErrorResponse(err error, ut ut.Translator) ErrorResponse {
	var domainError *domain.Error
	var validationError validator.ValidationErrors

	switch {
	case errors.As(err, &domainError):
		return ErrorResponse{
			Code:     domainError.Code,
			Messages: []string{domainError.Message},
		}
	case errors.As(err, &validationError):
		messages := make([]string, 0, len(validationError))
		for _, e := range validationError {
			messages = append(messages, e.Translate(ut))
		}

		return ErrorResponse{
			Code:     "invalid_input",
			Messages: messages,
		}
	default:
		zap.L().Error("Unknown error", zap.Error(errors.Unwrap(err)))
		return ErrorResponse{
			Code: "internal_server_error",
			Messages: []string{
				"Server cannot proceed the request.",
			},
		}
	}
}
