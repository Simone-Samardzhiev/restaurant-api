package rest

import (
	"errors"
	"log/slog"
	"menu/internal/domain"
	"net/http"

	"github.com/labstack/echo/v5"
)

// ErrorResponse represents JSON error response.
type ErrorResponse struct {
	HTTPStatus int    `json:"httpStatus"`
	Message    string `json:"message"`
	ErrorCode  string `json:"code"`
	err        error
}

func (e *ErrorResponse) Error() string {
	return e.err.Error()
}

func (e *ErrorResponse) Unwrap() error {
	return errors.Unwrap(e.err)
}

const invalidJSONErrorCode = "INVALID_JSON"

// NewInvalidJSONError creates and allocates new [ErrorResponse] from an error
// returned from decoding JSON.
func NewInvalidJSONError(err error) *ErrorResponse {
	return &ErrorResponse{
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid JSON payload.",
		ErrorCode:  invalidJSONErrorCode,
		err:        err,
	}
}

const invalidUUIDErrorCode = "INVALID_UUID"

// NewInvalidUUIDError creates and allocates new [ErrorResponse] from an error
// returned by UUID.
func NewInvalidUUIDError(err error) *ErrorResponse {
	return &ErrorResponse{
		HTTPStatus: http.StatusBadRequest,
		Message:    "Invalid UUID format.",
		ErrorCode:  invalidUUIDErrorCode,
		err:        err,
	}
}

var codesToStatus = map[domain.ErrorCode]int{
	domain.ErrorCodeInternal:             http.StatusInternalServerError,
	domain.ErrorCodeCategoryNameConflict: http.StatusConflict,
	domain.ErrorCodeCategoryNotFound:     http.StatusNotFound,
}

func translateCodeToStatus(code domain.ErrorCode) int {
	if status, ok := codesToStatus[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

var codesToMessage = map[domain.ErrorCode]string{
	domain.ErrorCodeInternal:             "Internal server error.",
	domain.ErrorCodeCategoryNameConflict: "Category name is already taken.",
	domain.ErrorCodeCategoryNotFound:     "Category not found.",
}

func translateCodeToMessage(code domain.ErrorCode) string {
	if status, ok := codesToMessage[code]; ok {
		return status
	}
	return codesToMessage[domain.ErrorCodeInternal]
}

// NewErrorResponse translates [domain.Error] into [ErrorResponse]
// if the error is not of type [domain.Error], the error will be translated
// into internal error.
func NewErrorResponse(err error) *ErrorResponse {
	if domainErr, ok := errors.AsType[*domain.Error](err); ok {
		return &ErrorResponse{
			HTTPStatus: translateCodeToStatus(domainErr.Code),
			Message:    translateCodeToMessage(domainErr.Code),
			ErrorCode:  domainErr.Code.String(),
			err:        domainErr,
		}
	}

	return &ErrorResponse{
		HTTPStatus: http.StatusInternalServerError,
		Message:    translateCodeToMessage(domain.ErrorCodeInternal),
		ErrorCode:  translateCodeToMessage(domain.ErrorCodeInternal),
		err:        err,
	}
}

type ValidationErrorResponse struct {
	ErrorResponse
	Fields map[string][]string `json:"fields"`
}

const invalidPayloadErrorCode = "INVALID_PAYLOAD"

var validationErr = errors.New("validation error")

func NewValidationError(fields map[string][]string) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		ErrorResponse: ErrorResponse{
			HTTPStatus: http.StatusUnprocessableEntity,
			Message:    "Payload validation error.",
			ErrorCode:  invalidPayloadErrorCode,
			err:        validationErr,
		},
		Fields: fields,
	}
}

func errorHandler(c *echo.Context, err error) {
	if appErr, ok := errors.AsType[*ErrorResponse](err); ok {
		if appErr.HTTPStatus >= http.StatusInternalServerError {
			c.Logger().Error(
				"Internal server error",
				slog.Any("error", appErr.err),
				slog.Any("cause", errors.Unwrap(appErr.err)),
			)
		}

		_ = c.JSON(appErr.HTTPStatus, appErr)

		return
	}

	if valErr, ok := errors.AsType[*ValidationErrorResponse](err); ok {
		_ = c.JSON(valErr.HTTPStatus, valErr)
		return
	}

	if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		_ = c.JSON(echoErr.Code, ErrorResponse{
			HTTPStatus: echoErr.Code,
			Message:    echoErr.Message,
			ErrorCode:  "METHOD_NOT_ALLOWED",
		})
	}

	c.Logger().Error("Unknown error", slog.Any("error", err))
	_ = c.JSON(http.StatusInternalServerError, ErrorResponse{
		HTTPStatus: http.StatusInternalServerError,
		Message:    translateCodeToMessage(domain.ErrorCodeInternal),
		ErrorCode:  domain.ErrorCodeInternal.String(),
	})
}
