package domain

import (
	"fmt"
)

// ErrorType represents the kind of error
type ErrorType int

const (
	InternalError ErrorType = iota
	NotFound
	Conflict
	BadRequest
	InvalidState
)

// ResourceType represents domain resource names
type ResourceType string

const (
	ProductResource         ResourceType = "product"
	ProductCategoryResource ResourceType = "product_category"
	OrderedProductResource  ResourceType = "ordered_product"
	OrderSessionResource    ResourceType = "order_session"
)

// BadRequestReason represents specific reasons for BadRequest errors
type BadRequestReason string

const (
	InvalidUUID        BadRequestReason = "invalid UUID"
	InvalidImageFormat BadRequestReason = "invalid image format"
	NothingToUpdate    BadRequestReason = "nothing to update"
	NothingToDelete    BadRequestReason = "nothing to delete"
	MultipleDelete     BadRequestReason = "multiple delete criteria"
)

// InvalidStateReason represents specific reasons for InvalidState errors.
type InvalidStateReason string

const (
	OrderSessionNotOpen      InvalidStateReason = "order session not open"
	OrderedProductNotPending InvalidStateReason = "ordered product not pending"
	ProductsAreIncomplete    InvalidStateReason = "products are incomplete"
)

// Error is the structured domain error
type Error struct {
	ErrorType ErrorType
	Code      string
	Message   string
	Resource  ResourceType
}

func (e *Error) Error() string {
	return e.Message
}

func NewInternalError() *Error {
	return &Error{
		ErrorType: InternalError,
		Code:      "internal_error",
		Message:   "internal error",
	}
}

// NewNotFoundError creates a new Error with specific not found ResourceType.
func NewNotFoundError(resource ResourceType) *Error {
	return &Error{
		ErrorType: NotFound,
		Code:      fmt.Sprintf("%s_not_found", resource),
		Message:   fmt.Sprintf("%s was not found", resource),
		Resource:  resource,
	}
}

// NewConflictError creates a new Error with specific conflicting ResourceType.
func NewConflictError(resource ResourceType) *Error {
	return &Error{
		ErrorType: Conflict,
		Code:      fmt.Sprintf("%s_conflict", resource),
		Message:   fmt.Sprintf("%s already exists", resource),
		Resource:  resource,
	}
}

// NewBadRequestError creates a new Error with specific BadRequestReason.
func NewBadRequestError(reason BadRequestReason) *Error {
	return &Error{
		ErrorType: BadRequest,
		Code:      "bad_request",
		Message:   string(reason),
	}
}

// NewInvalidStateError creates a new Error with specific InvalidStateReason.
func NewInvalidStateError(reason InvalidStateReason) *Error {
	return &Error{
		ErrorType: InvalidState,
		Code:      "bad_state",
		Message:   string(reason),
	}
}
