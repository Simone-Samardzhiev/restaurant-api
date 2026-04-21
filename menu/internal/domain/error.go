package domain

// ErrorCode represent a specific code of the error.
type ErrorCode int

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
