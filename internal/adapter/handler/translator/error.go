package translator

import (
	"restaurant/internal/domain"

	"go.uber.org/zap"
)

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Details []ErrorResponseDetail `json:"details,omitempty"`
}

// ErrorResponseDetail represents error details response.
type ErrorResponseDetail struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata,omitempty"`
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

	domain.ErrorCodeInvalidProduct:             "Provided product contains invalid data.",
	domain.ErrorCodeProductNameTooShort:        "Product name is too short.",
	domain.ErrorCodeProductNameTooLong:         "Product name is too long.",
	domain.ErrorCodeProductDescriptionTooShort: "Product description is too short.",
	domain.ErrorCodeProductPriceLessThanZero:   "Product price cannot be less than 0.",
	domain.ErrorCodeProductNameConflict:        "Product name is already used.",
	domain.ErrorCodeInvalidProductUpdate:       "Provided product update contains invalid data.",
	domain.ErrorCodeProductNotFound:            "Product not found.",
	domain.ErrorCodeProductNotFoundByID:        "Product with provided id not found.",
	domain.ErrorCodeProductHasLinkedOrders:     "Product cannot be deleted as it has linked orders.",

	domain.ErrorCodeInvalidSession:       "Provided session contains invalid data.",
	domain.ErrorCodeInvalidSessionTable:  "Invalid session table.",
	domain.ErrorCodeInvalidSessionStatus: "Invalid session status.",
	domain.ErrorCodeSessionNotOpened:     "Session is not opened.",
	domain.ErrorCodeInvalidSessionUpdate: "Provided session update contains invalid data.",

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

// DomainError translates [domain.Error] into ErrorResponse.
func DomainError(err *domain.Error) ErrorResponse {
	if err.Kind == domain.ErrorKindInternal {
		zap.L().Error("internal server error", zap.Error(err), zap.NamedError("cause", err.Cause))
		return ErrorResponse{
			Code:    domain.ErrorCodeInternal.String(),
			Message: "Internal server error.",
		}
	}

	resp := ErrorResponse{
		Code:    err.Code.String(),
		Message: mapErrorCodeToMessage(err.Code),
		Details: make([]ErrorResponseDetail, 0, len(err.Details)),
	}

	for _, detail := range err.Details {
		resp.Details = append(resp.Details, ErrorResponseDetail{
			Code:     detail.Code.String(),
			Message:  mapErrorCodeToMessage(detail.Code),
			Metadata: detail.Metadata,
		})
	}

	return resp
}
