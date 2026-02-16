package domain

// ErrorKind represents different error kinds.
type ErrorKind int

const (
	ErrorKindInternal ErrorKind = iota + 1
	ErrorKindValidation
	ErrorKindConflict
	ErrorKindBadRequest
)

var mapErrorKind = map[ErrorKind]string{
	ErrorKindInternal:   "internal",
	ErrorKindValidation: "validation",
	ErrorKindConflict:   "conflict",
	ErrorKindBadRequest: "bad_request",
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

	ErrorCodeInvalidUUID
)

var mapErrorCode = map[ErrorCode]string{
	ErrorCodeInternal:             "INTERNAL",
	ErrorCodeInvalidCategory:      "INVALID_CATEGORY",
	ErrorCodeCategoryNameTooShort: "CATEGORY_NAME_TOO_SHORT",
	ErrorCodeCategoryNameTooLong:  "CATEGORY_NAME_TOO_LONG",
	ErrorCodeCategoryNameConflict: "CATEGORY_NAME_CONFLICT",

	ErrorCodeInvalidUUID: "INVALID_UUID",
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
