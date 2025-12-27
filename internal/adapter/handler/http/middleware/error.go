package middleware

import (
	"errors"
	"net/http"
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/domain"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// errorTypeToHTTPCode maps domain.ErrorType to the appropriate HTTP status code.
func errorTypeToHTTPCode(errorType domain.ErrorType) int {
	switch errorType {
	case domain.InternalError:
		return http.StatusInternalServerError
	case domain.NotFound:
		return http.StatusNotFound
	case domain.Conflict:
		return http.StatusConflict
	case domain.BadRequest:
		return http.StatusBadRequest
	case domain.InvalidState:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// handleDomainError handler domain.Error and sends appropriate ErrorResponse.
func handleDomainError(c *gin.Context, err *domain.Error) {
	c.JSON(errorTypeToHTTPCode(err.ErrorType), response.ErrorResponse{
		Code:     err.Code,
		Messages: []string{err.Message},
	})
}

func handleValidateError(c *gin.Context, translator ut.Translator, err validator.ValidationErrors) {
	messages := make([]string, 0, len(err))
	for _, v := range err {
		messages = append(messages, v.Translate(translator))
	}

	c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
		Code:     "invalid_input",
		Messages: messages,
	})
}

// Error is used to handle errors attached to the gin.Context.
func Error(translator ut.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		lastErr := c.Errors.Last()
		var domainErr *domain.Error
		var validationErrors validator.ValidationErrors

		switch {
		case errors.As(lastErr, &domainErr):
			handleDomainError(c, domainErr)
		case errors.As(lastErr, &validationErrors):
			handleValidateError(c, translator, validationErrors)
		default:
			zap.L().Error("unknown error", zap.Error(lastErr))
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Code:     "internal_error",
				Messages: []string{"Internal server error"},
			})
		}
	}
}
