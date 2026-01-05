package domain

import (
	"runtime"
)

// Field holds the name and value.
type Field struct {
	Name  string
	Value any
}

// F creates a Field instance with name and value.
func F(name string, value any) Field {
	return Field{
		Name:  name,
		Value: value,
	}
}

// InternalError represents an internal error.
type InternalError struct {
	Message  string
	Metadata []Field
	Filepath string
	Line     int
	Cause    error
}

// NewInternalError creates a new InternalError.
//
// Note:
//
//	For better logging and debugging this function captures its caller, so
//	call it as close as possible to where the internal error occurred.
func NewInternalError(message string, cause error, metadata ...Field) *InternalError {
	_, file, line, _ := runtime.Caller(1)

	return &InternalError{
		Message:  message,
		Filepath: file,
		Metadata: metadata,
		Line:     line,
		Cause:    cause,
	}
}

func (e *InternalError) Error() string {
	return e.Message
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}
