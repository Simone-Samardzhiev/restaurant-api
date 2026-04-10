package domain

// ErrorKind represents different error kinds.
type ErrorKind int

const (
	ErrorKindInternal ErrorKind = iota + 1
	ErrorKindValidation
	ErrorKindConflict
	ErrorKindBadRequest
	ErrorKindNotFound
)

var mapErrorKind = map[ErrorKind]string{
	ErrorKindInternal:   "internal",
	ErrorKindValidation: "validation",
	ErrorKindConflict:   "conflict",
	ErrorKindBadRequest: "bad_request",
	ErrorKindNotFound:   "not_found",
}

func (e ErrorKind) String() string {
	result, ok := mapErrorKind[e]
	if !ok {
		return "unknown"
	}
	return result
}

// ErrorCode represents different error codes for easier identification
// of an error.
type ErrorCode int

const (
	ErrorCodeInternal ErrorCode = iota + 1

	ErrorCodeInvalidCategory
	ErrorCodeCategoryNameTooShort
	ErrorCodeCategoryNameTooLong
	ErrorCodeCategoryNameConflict
	ErrorCodeInvalidCategoryUpdate
	ErrorCodeCategoryNotFound
	ErrorCodeCategoryNotFoundByID
	ErrorCodeCategoryHasLinkedProducts

	ErrorCodeInvalidImageType
	ErrorCodeImageNotFound
	ErrorCodeInvalidImageUpdate

	ErrorCodeInvalidProduct
	ErrorCodeProductNameTooShort
	ErrorCodeProductNameTooLong
	ErrorCodeProductDescriptionTooShort
	ErrorCodeProductPriceLessThanZero
	ErrorCodeProductNameConflict
	ErrorCodeInvalidProductUpdate
	ErrorCodeProductNotFound
	ErrorCodeProductNotFoundByID
	ErrorCodeProductHasLinkedOrders

	ErrorCodeInvalidSession
	ErrorCodeInvalidSessionTable
	ErrorCodeInvalidSessionStatus

	ErrorCodeMalformedRequest
	ErrorCodeNoData
	ErrorCodeInvalidUUID
)

var mapErrorCode = map[ErrorCode]string{
	ErrorCodeInternal: "INTERNAL",

	ErrorCodeInvalidCategory:           "INVALID_CATEGORY",
	ErrorCodeCategoryNameTooShort:      "CATEGORY_NAME_TOO_SHORT",
	ErrorCodeCategoryNameTooLong:       "CATEGORY_NAME_TOO_LONG",
	ErrorCodeCategoryNameConflict:      "CATEGORY_NAME_CONFLICT",
	ErrorCodeInvalidCategoryUpdate:     "INVALID_CATEGORY_UPDATE",
	ErrorCodeCategoryNotFound:          "CATEGORY_NOT_FOUND",
	ErrorCodeCategoryNotFoundByID:      "CATEGORY_NOT_FOUND_BY_ID",
	ErrorCodeCategoryHasLinkedProducts: "CATEGORY_HAS_LINKED_PRODUCTS",

	ErrorCodeInvalidImageType:   "INVALID_IMAGE_TYPE",
	ErrorCodeImageNotFound:      "IMAGE_NOT_FOUND",
	ErrorCodeInvalidImageUpdate: "INVALID_IMAGE_UPDATE",

	ErrorCodeInvalidProduct:             "INVALID_PRODUCT",
	ErrorCodeProductNameTooShort:        "PRODUCT_NAME_TOO_SHORT",
	ErrorCodeProductNameTooLong:         "PRODUCT_NAME_TOO_LONG",
	ErrorCodeProductDescriptionTooShort: "PRODUCT_DESCRIPTION_TOO_SHORT",
	ErrorCodeProductPriceLessThanZero:   "PRODUCT_PRICE_LESS_THAN_ZERO",
	ErrorCodeProductNameConflict:        "PRODUCT_NAME_CONFLICT",
	ErrorCodeInvalidProductUpdate:       "INVALID_PRODUCT_UPDATE",
	ErrorCodeProductNotFound:            "PRODUCT_NOT_FOUND",
	ErrorCodeProductNotFoundByID:        "PRODUCT_NOT_FOUND_BY_ID",
	ErrorCodeProductHasLinkedOrders:     "PRODUCT_HAS_LINKED_ORDERS",

	ErrorCodeInvalidSession:       "INVALID_SESSION",
	ErrorCodeInvalidSessionTable:  "INVALID_SESSION_TABLE",
	ErrorCodeInvalidSessionStatus: "INVALID_SESSION_STATUS",

	ErrorCodeMalformedRequest: "MALFORMED_REQUEST",
	ErrorCodeNoData:           "NO_DATA",
	ErrorCodeInvalidUUID:      "INVALID_UUID",
}

func (e ErrorCode) String() string {
	result, ok := mapErrorCode[e]
	if !ok {
		return "UNKNOWN"
	}
	return result
}

// ErrorDetail represents a further describing detail of an error.
type ErrorDetail struct {
	Code     ErrorCode
	Message  string
	Metadata map[string]any
}

func (e *ErrorDetail) Error() string {
	return e.Message
}

// Error represents an descriptive error.
type Error struct {
	Kind    ErrorKind
	Code    ErrorCode
	Message string
	Details []ErrorDetail
	Cause   error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// NewInternalError creates and allocates a new Error.
//
// It sets the [Error.Kind] to [ErrorKindInternal] and [Error.Code] to [ErrorCodeInternal].
func NewInternalError(message string, cause error) *Error {
	return &Error{
		Kind:    ErrorKindInternal,
		Code:    ErrorCodeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates and allocates a new Error.
//
// It sets the [Error.Kind] to [ErrorKindValidation].
func NewValidationError(message string, code ErrorCode, details ...ErrorDetail) *Error {
	return &Error{
		Kind:    ErrorKindValidation,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewConflictError creates and allocates a new Error.
// It sets the [Error.Kind] to [ErrorKindConflict].
func NewConflictError(message string, code ErrorCode, cause error) *Error {
	return &Error{
		Kind:    ErrorKindConflict,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewBadRequestError creates and allocates a new Error.
// It sets the [Error.Kind] to [ErrorKindBadRequest].
func NewBadRequestError(message string, code ErrorCode, cause error) *Error {
	return &Error{
		Kind:    ErrorKindBadRequest,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewNotFoundError creates and allocates a new Error.
// It sets the [Error.Kind] tp [ErrorKindNotFound].
func NewNotFoundError(message string, code ErrorCode, details ...ErrorDetail) *Error {
	return &Error{
		Kind:    ErrorKindNotFound,
		Code:    code,
		Message: message,
		Details: details,
	}
}
