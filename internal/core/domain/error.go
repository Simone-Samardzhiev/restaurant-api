package domain

import "errors"

var (
	// ErrInternal represents a generic internal service failure.
	ErrInternal = errors.New("internal failure")

	// ErrProductNameAlreadyInUse indicates a product name is already in use.
	ErrProductNameAlreadyInUse = errors.New("product name is already in use")

	// ErrProductCategoryNotFound indicates a product category couldn't be found.
	ErrProductCategoryNotFound = errors.New("product category not found")

	// ErrInvalidImageFormat indicates a image format is not supported.
	ErrInvalidImageFormat = errors.New("invalid image format")
)
