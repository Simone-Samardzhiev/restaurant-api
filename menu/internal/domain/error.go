package domain

// ErrorCode represent a specific code of the error.
type ErrorCode int

const (
	ErrorCodeInternal ErrorCode = iota + 1

	ErrorCodeCategoryNameConflict
	ErrorCodeCategoryNotFound
	ErrorCodeCategoryHasProducts

	ErrorCodeProductNameConflict
)

var mapErrorCodes = map[ErrorCode]string{
	ErrorCodeInternal: "INTERNAL_ERROR",

	ErrorCodeCategoryNameConflict: "CATEGORY_NAME_CONFLICT",
	ErrorCodeCategoryNotFound:     "CATEGORY_NOT_FOUND",
	ErrorCodeCategoryHasProducts:  "CATEGORY_HAS_PRODUCTS",

	ErrorCodeProductNameConflict: "PRODUCT_NAME_CONFLICT",
}

func (e ErrorCode) String() string {
	code, ok := mapErrorCodes[e]
	if ok {
		return code
	}
	return mapErrorCodes[ErrorCodeInternal]
}

// Error represents an app error.
type Error struct {
	message string
	Code    ErrorCode
	cause   error
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Error() string {
	return e.message
}

// NewError creates and allocates a new error.
func NewError(message string, code ErrorCode, cause error) *Error {
	return &Error{
		message: message,
		Code:    code,
		cause:   cause,
	}
}
