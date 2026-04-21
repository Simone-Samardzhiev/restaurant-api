package domain

// ErrorCode represent a specific code of the error.
type ErrorCode int

const (
	ErrorCodeInternal = iota + 1
	ErrorCodeCategoryNameConflict
)

func (e ErrorCode) String() string {
	switch e {
	case ErrorCodeInternal:
		return "INTERNAL_ERROR"
	case ErrorCodeCategoryNameConflict:
		return "CATEGORY_NAME_CONFLICT"
	default:
		return "UNKNOWN_ERROR"
	}
}

// Error represents an app error.
type Error struct {
	message string
	code    ErrorCode
	cause   error
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Error() string {
	return e.message
}

// NewError creates and allocates a new error.
func NewError(message string, code ErrorCode, cause error) *Error {
	return &Error{
		message: message,
		code:    code,
		cause:   cause,
	}
}
