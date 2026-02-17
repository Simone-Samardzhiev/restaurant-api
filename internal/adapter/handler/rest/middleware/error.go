package middleware

import (
	"errors"
	"net/http"
	"restaurant/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// mapErrorCode is used to map [domain.ErrorCode] to user message.
var mapErrorCode = map[domain.ErrorCode]string{
	domain.ErrorCodeInternal: "Internal server error.",

	domain.ErrorCodeInvalidCategory:       "Provided category contains invalid data.",
	domain.ErrorCodeCategoryNameTooShort:  "Category name is too short.",
	domain.ErrorCodeCategoryNameTooLong:   "Category name is too long.",
	domain.ErrorCodeCategoryNameConflict:  "Category name is already used.",
	domain.ErrorCodeInvalidCategoryUpdate: "Provided category update contains invalid data.",
	domain.ErrorCodeCategoryNotFound:      "Category not found.",
	domain.ErrorCodeCategoryNotFoundByID:  "Category with provided id not found.",

	domain.ErrorCodeMalformedRequest: "Request payload is malformed.",
	domain.ErrorCodeNoData:           "Request does not have any data.",
	domain.ErrorCodeInvalidUUID:      "Invalid UUID.",
}

// mapErrorCodeToMessage maps error codes to user message.
// If the code is not found the function returns `Internal server error.`.
func mapErrorCodeToMessage(code domain.ErrorCode) string {
	result, ok := mapErrorCode[code]
	if !ok {
		return "Internal server error."
	}

	return result
}

// mapErrorKind is used to map [domain.ErrorKind] to http status codes.
var mapErrorKind = map[domain.ErrorKind]int{
	domain.ErrorKindInternal:   http.StatusInternalServerError,
	domain.ErrorKindValidation: http.StatusUnprocessableEntity,
	domain.ErrorKindConflict:   http.StatusConflict,
	domain.ErrorKindBadRequest: http.StatusBadRequest,
	domain.ErrorKindNotFound:   http.StatusNotFound,
}

// mapErrorKindHttpCode maps [domain.ErrorKind] to http status codes.
// If the kind is not found [http.StatusInternalServerError] is returned.
func mapErrorKindHttpCode(kind domain.ErrorKind) int {
	result, ok := mapErrorKind[kind]
	if !ok {
		return http.StatusInternalServerError
	}
	return result
}

// ErrorResponse represents an api error response.
type ErrorResponse struct {
	Status  int                    `json:"status"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details []ErrorDetailsResponse `json:"details,omitempty"`
}

// ErrorDetailsResponse represents more details about an error.
type ErrorDetailsResponse struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

// handleDomainError responses with appropriate [ErrorResponse], from [domain.Error].
func handleDomainError(c *gin.Context, err *domain.Error) {
	if err.Kind == domain.ErrorKindInternal {
		zap.L().Error(
			"internal server error",
			zap.NamedError("cause", err.Cause),
			zap.Error(err),
		)
	}

	response := ErrorResponse{
		Status:  mapErrorKindHttpCode(err.Kind),
		Code:    err.Code.String(),
		Message: mapErrorCodeToMessage(err.Code),
		Details: []ErrorDetailsResponse{},
	}

	for _, detail := range err.Details {
		response.Details = append(response.Details, ErrorDetailsResponse{
			Code:     detail.Code.String(),
			Message:  mapErrorCodeToMessage(detail.Code),
			Metadata: detail.Metadata,
		})
	}

	c.AbortWithStatusJSON(response.Status, response)
}

// handleUnknownError  responses with appropriate [ErrorResponse], from any errors and logs it.
func handleUnknownError(c *gin.Context, err error) {
	zap.L().Error("unknown error", zap.Error(err))
	response := ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    domain.ErrorCodeInternal.String(),
		Message: mapErrorCodeToMessage(domain.ErrorCodeInternal),
	}

	c.AbortWithStatusJSON(response.Status, response)
}

// Error returns a middleware handles all stored errors in [gin.Context].
func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		lastError := ctx.Errors.Last()
		if lastError == nil {
			return
		}

		if domainErr, ok := errors.AsType[*domain.Error](lastError.Err); ok {
			handleDomainError(ctx, domainErr)
			return
		}

		handleUnknownError(ctx, lastError)
	}
}
