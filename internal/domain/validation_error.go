package domain

import (
	"strings"
)

// ValidationError represents failed validation of a field.
type ValidationError struct {
	Field string
	Cause error
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, cause error) ValidationError {
	return ValidationError{field, cause}
}

func (e ValidationError) Error() string {
	return e.Cause.Error()
}

func (e ValidationError) Unwrap() error {
	return e.Cause
}

// ValidationErrors aggregates ValidationError.
type ValidationErrors struct {
	Message string
	Errors  []ValidationError
}

// NewValidationErrors creates a new empty ValidationErrors.
func NewValidationErrors(message string) *ValidationErrors {
	return &ValidationErrors{
		Message: message,
		Errors:  []ValidationError{},
	}
}

func (e *ValidationErrors) Error() string {
	builder := strings.Builder{}

	builder.WriteString(e.Message + ": ")

	size := len(e.Errors)
	for i := 0; i < size; i++ {
		if i < size-1 {
			builder.WriteString(e.Errors[i].Error() + ", ")
			continue
		}

		builder.WriteString(e.Errors[i].Error())
	}

	return builder.String()
}

// Add adds a new ValidationError from a field and error.
func (e *ValidationErrors) Add(field string, err error) {
	e.Errors = append(e.Errors, NewValidationError(field, err))
}

// AddError adds a single ValidationError.
func (e *ValidationErrors) AddError(validationError ValidationError) {
	e.Errors = append(e.Errors, validationError)
}

// AddErrors adds multiple ValidationError.
func (e *ValidationErrors) AddErrors(validationErrors ...ValidationError) {
	e.Errors = append(e.Errors, validationErrors...)
}

// HasErrors returns contains any errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}
