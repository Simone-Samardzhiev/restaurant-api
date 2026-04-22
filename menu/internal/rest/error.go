package rest

import (
	"errors"
	"log/slog"
	"menu/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents JSON error response.
type ErrorResponse struct {
	HTTPStatus int    `json:"httpStatus"`
	Message    string `json:"message"`
	ErrorCode  string `json:"code"`
}

// InvalidJSONErrorResponse should be returned if decoding JSON fails.
var InvalidJSONErrorResponse = ErrorResponse{
	HTTPStatus: http.StatusBadRequest,
	Message:    "Malformed JSON payload.",
	ErrorCode:  "INVALID_JSON",
}

// translateErrorCodeToHTTPStatus translates [domain.ErrorCode] into http status code.
func translateErrorCodeToHTTPStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ErrorCodeInternal:
		return http.StatusInternalServerError
	case domain.ErrorCodeCategoryNameConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// translateErrorCodeToMessage translates [domain.ErrorCode] into end client message.
func translateErrorCodeToMessage(code domain.ErrorCode) string {
	switch code {
	case domain.ErrorCodeInternal:
		return "Internal server error."
	case domain.ErrorCodeCategoryNameConflict:
		return "Category name is already taken."
	default:
		return "Internal server error."
	}
}

// translateError translates the error into [ErrorResponse].
//
// If the error is not of type [domain.Error], the error will be logged.
//
// If the error is of type [domain.Error], and when translating the returned status code
// is 500, the error will also be logged.
func translateError(err error) ErrorResponse {
	domainErr, ok := errors.AsType[*domain.Error](err)
	if !ok {
		slog.Error("Unknown error type", slog.Any("error", err))

		return ErrorResponse{
			HTTPStatus: http.StatusInternalServerError,
			Message:    "Internal server error.",
			ErrorCode:  domain.ErrorCodeInternal.String(),
		}
	}

	status := translateErrorCodeToHTTPStatus(domainErr.Code)
	if status == http.StatusInternalServerError {
		slog.Error(
			"Internal server error.",
			slog.Any("error", err),
			slog.Any("cause", domainErr.Cause()),
		)
	}

	return ErrorResponse{
		HTTPStatus: status,
		Message:    translateErrorCodeToMessage(domainErr.Code),
		ErrorCode:  domainErr.Code.String(),
	}
}

// handleError translates the error and sends [ErrorResponse] as JSON.
func handleError(ctx *gin.Context, err error) {
	response := translateError(err)
	ctx.JSON(response.HTTPStatus, response)
}

// ValidationError represents JSON error response from validation payload.
type ValidationError struct {
	ErrorResponse
	Fields map[string][]string `json:"fields"`
}

// handleValidationError sends [ValidationError] as JSON with the provided errors.
func handleValidationError(ctx *gin.Context, errors map[string][]string) {
	ctx.JSON(http.StatusUnprocessableEntity, ValidationError{
		ErrorResponse: ErrorResponse{
			HTTPStatus: http.StatusUnprocessableEntity,
			Message:    "Payload validation failed.",
			ErrorCode:  "INVALID_PAYLOAD",
		},
		Fields: errors,
	})
}
