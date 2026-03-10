package translator

import (
	"errors"
	"net/http"
	"restaurant/internal/domain"

	"go.uber.org/zap"
)

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

// mapErrorCode is used to map [domain.ErrorCode] to user message.
var mapErrorCode = map[domain.ErrorCode]string{
	domain.ErrorCodeInternal: "Internal server error.",

	domain.ErrorCodeInvalidCategory:           "Provided category contains invalid data.",
	domain.ErrorCodeCategoryNameTooShort:      "Category name is too short.",
	domain.ErrorCodeCategoryNameTooLong:       "Category name is too long.",
	domain.ErrorCodeCategoryNameConflict:      "Category name is already used.",
	domain.ErrorCodeInvalidCategoryUpdate:     "Provided category update contains invalid data.",
	domain.ErrorCodeCategoryNotFound:          "Category not found.",
	domain.ErrorCodeCategoryNotFoundByID:      "Category with provided id not found.",
	domain.ErrorCodeCategoryHasLinkedProducts: "Category cannot be deleted as it has linked products.",

	domain.ErrorCodeInvalidImageType:   "Invalid image type.",
	domain.ErrorCodeImageNotFound:      "Image not found.",
	domain.ErrorCodeInvalidImageUpdate: "Invalid image update.",

	domain.ErrorCodeInvalidProduct:             "Invalid product.",
	domain.ErrorCodeProductNameTooShort:        "Product name is too short.",
	domain.ErrorCodeProductNameTooLong:         "Product name is too long.",
	domain.ErrorCodeProductDescriptionTooShort: "Product description is too short.",
	domain.ErrorCodeProductPriceLessThanZero:   "Product price cannot be less than 0.",
	domain.ErrorCodeProductNameConflict:        "Product name is already used.",
	domain.ErrorCodeInvalidProductUpdate:       "Provided product update contains invalid data.",
	domain.ErrorCodeProductNotFound:            "Product not found.",
	domain.ErrorCodeProductNotFoundByID:        "Product with provided id not found.",
	domain.ErrorCodeProductHasLinkedOrders:     "Product cannot be deleted as it has linked orders.",

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

// TranslateError translates errors into [ErrorResponse].
// If the error type is not [domain.Error] or the [domain.ErrorKind]
// is internal, the error is logged automatically.
func TranslateError(err error) ErrorResponse {
	domainErr, ok := errors.AsType[*domain.Error](err)
	if !ok {
		zap.L().Error("unknown error", zap.Error(err))

		return ErrorResponse{
			Status:  http.StatusInternalServerError,
			Code:    domain.ErrorCodeInternal.String(),
			Message: mapErrorCodeToMessage(domain.ErrorCodeInternal),
		}
	}

	if domainErr.Kind == domain.ErrorKindInternal {
		zap.L().Error("internal error", zap.Error(err), zap.NamedError("cause", domainErr.Cause))
	}

	resp := ErrorResponse{
		Status:  mapErrorKindHttpCode(domainErr.Kind),
		Code:    domainErr.Code.String(),
		Message: mapErrorCodeToMessage(domainErr.Code),
		Details: make([]ErrorDetailsResponse, 0, len(domainErr.Details)),
	}

	for _, details := range domainErr.Details {
		resp.Details = append(resp.Details, ErrorDetailsResponse{
			Code:     mapErrorCodeToMessage(details.Code),
			Message:  details.Message,
			Metadata: details.Metadata,
		})
	}
	return resp
}
