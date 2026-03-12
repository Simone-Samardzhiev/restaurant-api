package middleware

import (
	"errors"
	"restaurant/internal/adapter/handler/translator"
	"restaurant/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorResponse represents an HTTP error response.
type ErrorResponse struct {
	Status int `json:"status"`
	translator.ErrorResponse
}

// Error returns a middleware handles stored errors in [gin.Context].
func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) == 0 {
			return
		}

		domainErr, ok := errors.AsType[*domain.Error](ctx.Errors.Last())
		if !ok {
			zap.L().Error("unknown error", zap.Error(ctx.Errors.Last()))
		}
		ctx.AbortWithStatusJSON(
			translator.ErrorKindToHTTPStatus(domainErr.Kind),
			translator.DomainError(domainErr),
		)
	}
}
