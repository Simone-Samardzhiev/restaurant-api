package rest

import (
	"errors"
	"log/slog"
	"menu/internal/domain"
	"net/http"

	"github.com/labstack/echo/v5"
)

// Error represents an app error.
type Error struct {
	HttpStatus int
	Code       string
	Message    string
	Err        error
}

// mapErrorCodesToHTTPStatus maps [domain.ErrorCode] to appropriate HTTP status code.
var mapErrorCodesToHTTPStatus = map[domain.ErrorCode]int{
	domain.ErrorCodeInternal:             http.StatusInternalServerError,
	domain.ErrorCodeCategoryNameConflict: http.StatusConflict,
	domain.ErrorCodeCategoryNotFound:     http.StatusNotFound,
}

// mapErrorCodesToMessage maps [domain.ErrorCode] to appropriate end user message.
var mapErrorCodesToMessage = map[domain.ErrorCode]string{
	domain.ErrorCodeInternal:             "Internal server error.",
	domain.ErrorCodeCategoryNameConflict: "Category name already exists.",
	domain.ErrorCodeCategoryNotFound:     "Category not found.",
}

// NewError creates a new [Error]. If the error of type [domain.Error] the message and the code gets translated,
// otherwise the [domain.ErrorCodeInternal] is used.
func NewError(err error) *Error {
	code := domain.ErrorCodeInternal

	if domainErr, ok := errors.AsType[*domain.Error](err); ok {
		code = domainErr.Code
	}

	status := mapErrorCodesToHTTPStatus[code]
	message := mapErrorCodesToMessage[code]

	return &Error{
		HttpStatus: status,
		Code:       code.String(),
		Message:    message,
		Err:        err,
	}
}

// Rest side error codes.
const (
	ErrorCodeInvalidJSON   = "INVALID_JSON"
	ErrorCodeInvalidUUID   = "INVALID_UUID"
	ErrorCodeInvalidEntity = "INVALID_ENTITY"
)

// NewInvalidJSONError creates new [Error] from JSON unmarshaling error.
func NewInvalidJSONError(err error) *Error {
	return &Error{
		HttpStatus: http.StatusBadRequest,
		Code:       ErrorCodeInvalidJSON,
		Message:    "JSON request is malformed.",
		Err:        err,
	}
}

// NewInvalidUUIDError creates new [Error] from UUID parsing error.
func NewInvalidUUIDError(err error) *Error {
	return &Error{
		HttpStatus: http.StatusBadRequest,
		Code:       ErrorCodeInvalidUUID,
		Message:    "UUID is malformed.",
		Err:        err,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

// ValidationError represents an error from validation payload.
type ValidationError struct {
	Fields map[string][]string
}

// NewValidationError create and allocates new [ValidationError].
func NewValidationError(fields map[string][]string) *ValidationError {
	return &ValidationError{Fields: fields}
}

func (e *ValidationError) Error() string {
	return "validation error"
}

// ErrorResponse represents a JSON response from [Error].
type ErrorResponse struct {
	Code       string `json:"code"`
	HttpStatus int    `json:"httpStatus"`
	Message    string `json:"message"`
	RequestID  string `json:"requestId"`
}

// ValidationErrorResponse represents response from [ValidationError].
type ValidationErrorResponse struct {
	ErrorResponse
	Fields map[string][]string `json:"fields"`
}

// ErrorHandler handles root errors in [echo.Echo].
func ErrorHandler(ctx *echo.Context, err error) {
	requestId := ctx.Request().Header.Get(echo.HeaderXRequestID)

	if e, ok := errors.AsType[*Error](err); ok {
		if e.HttpStatus >= http.StatusInternalServerError {
			ctx.Logger().LogAttrs(
				ctx.Request().Context(),
				slog.LevelError,
				"Internal error",
				slog.Any("error", err),
				slog.Any("cause", errors.Unwrap(err)),
			)
		}

		response := ErrorResponse{
			Code:       e.Code,
			HttpStatus: e.HttpStatus,
			Message:    e.Error(),
			RequestID:  requestId,
		}
		_ = ctx.JSON(response.HttpStatus, response)
		return
	}

	if e, ok := errors.AsType[*ValidationError](err); ok {
		response := ValidationErrorResponse{
			ErrorResponse: ErrorResponse{
				Code:       ErrorCodeInvalidEntity,
				HttpStatus: http.StatusUnprocessableEntity,
				Message:    "Request data is invalid.",
				RequestID:  requestId,
			},
			Fields: e.Fields,
		}

		_ = ctx.JSON(response.HttpStatus, response)
		return
	}

	ctx.Logger().LogAttrs(
		ctx.Request().Context(),
		slog.LevelError,
		"Unknow error",
		slog.Any("error", err),
	)

	response := ErrorResponse{
		Code:       domain.ErrorCodeInternal.String(),
		HttpStatus: mapErrorCodesToHTTPStatus[domain.ErrorCodeInternal],
		Message:    mapErrorCodesToMessage[domain.ErrorCodeInternal],
		RequestID:  requestId,
	}
	_ = ctx.JSON(response.HttpStatus, response)
}
