package domain

// ErrorType used for classification of errors.
type ErrorType int

const (
	Conflict = iota
	BadRequest
	NotFound
)

// Error represent a generic error.
type Error struct {
	Type    ErrorType
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new Error.
func NewError(t ErrorType, msg string) *Error {
	return &Error{
		Type:    t,
		Message: msg,
	}
}

// NewConflictError creates a new Error with ErrorType (Conflict).
func NewConflictError(msg string) *Error {
	return NewError(Conflict, msg)
}

// NewBadRequestError creates a new Error with ErrorType (BadRequest).
func NewBadRequestError(msg string) *Error {
	return NewError(BadRequest, msg)
}

// NewNotFoundError creates a new Error with ErrorType (NotFound).
func NewNotFoundError(msg string) *Error {
	return NewError(NotFound, msg)
}
