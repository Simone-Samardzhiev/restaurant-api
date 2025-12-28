package middleware

import (
	"errors"
	"net/http"
	"restaurant/internal/adapter/handler/http/response"
	"restaurant/internal/core/domain"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// errorTypeToHTTPCode maps domain.ErrorType to the appropriate HTTP status code.
func errorToHTTPCode(err error) int {
	var domainError *domain.Error
	var validationError validator.ValidationErrors

	switch {
	case errors.As(err, &domainError):
		switch domainError.ErrorType {
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
	case errors.As(err, &validationError):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// Error is used to handle errors attached to the gin.Context.
func Error(translator ut.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		lastErr := c.Errors.Last()
		status := errorToHTTPCode(lastErr)
		res := response.NewErrorResponse(lastErr, translator)
		c.JSON(status, res)
	}
}
