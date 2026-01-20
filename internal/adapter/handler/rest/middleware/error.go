package middleware

import (
	"errors"
	"net/http"
	"restaurant/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type internalErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleInternalError(ctx *gin.Context, err *domain.InternalError) {
	fields := make([]zap.Field, 0, len(err.Metadata)+6)
	fields = append(
		fields,
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.Request.URL.Path),
		zap.String("query", ctx.Request.URL.RawQuery),
		zap.String("filePath", err.Filepath),
		zap.Int("line", err.Line),
		zap.Error(err.Cause),
	)

	for _, el := range err.Metadata {
		fields = append(fields, zap.Any(el.Name, el.Value))
	}

	zap.L().Error(
		err.Message,
		fields...,
	)

	ctx.JSON(http.StatusInternalServerError, internalErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}

type validationErrorResponse struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Details []validationErrorDetail `json:"details"`
}

type validationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func handleValidationErrors(ctx *gin.Context, err *domain.ValidationErrors) {
	validationErrors := make([]validationErrorDetail, 0, len(err.Errors))
	for _, el := range err.Errors {
		validationErrors = append(validationErrors, validationErrorDetail{
			Field:   el.Field,
			Message: el.Error(),
		})
	}

	ctx.JSON(http.StatusUnprocessableEntity,
		validationErrorResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: err.Message,
			Details: validationErrors,
		},
	)
}

type genericErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleError(ctx *gin.Context, err *domain.Error) {
	var statusCode int

	switch err.Type {
	case domain.Conflict:
		statusCode = http.StatusConflict
	case domain.BadRequest:
		statusCode = http.StatusBadRequest
	case domain.NotFound:
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
	}

	var message string
	if statusCode == http.StatusInternalServerError {
		message = "internal server error"
	} else {
		message = err.Message
	}

	ctx.JSON(statusCode, genericErrorResponse{
		Code:    statusCode,
		Message: message,
	})
}

func handleUnknownError(ctx *gin.Context, err error) {
	zap.L().Error("unknown error", zap.Error(err))
	ctx.JSON(http.StatusInternalServerError, internalErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}

// ErrorMiddleware returns a Gin middleware that converts errors in the context into JSON responses.
func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		lstErr := ctx.Errors.Last().Err

		var internalErr *domain.InternalError
		var validationErr *domain.ValidationErrors
		var err *domain.Error

		switch {
		case errors.As(lstErr, &internalErr):
			handleInternalError(ctx, internalErr)
		case errors.As(lstErr, &validationErr):
			handleValidationErrors(ctx, validationErr)
		case errors.As(lstErr, &err):
			handleError(ctx, err)
		default:
			handleUnknownError(ctx, lstErr)
		}
	}
}
