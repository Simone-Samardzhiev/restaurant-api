package domain

import "errors"

var (
	// ErrInternal represents a generic internal service failure.
	ErrInternal = errors.New("internal failure")

	// ErrNothingToUpdate indicates an updates doesn't change any data.
	ErrNothingToUpdate = errors.New("nothing to update")

	// ErrInvalidUUID indicates an id is not valid uuid.
	ErrInvalidUUID = errors.New("invalid entity")

	// ErrProductNameAlreadyInUse indicates a product name is already in use.
	ErrProductNameAlreadyInUse = errors.New("product name is already in use")

	// ErrProductCategoryNameAlreadyInUse indicates a product category name is already in use.
	ErrProductCategoryNameAlreadyInUse = errors.New("product category is already in use")

	// ErrProductCategoryNotFound indicates a product category couldn't be found.
	ErrProductCategoryNotFound = errors.New("product category not found")

	// ErrCategoryHasLinkedProducts indicates an attempt to delete a product category that has linked products.
	ErrCategoryHasLinkedProducts = errors.New("category has linked products")

	// ErrInvalidImageFormat indicates a image format is not supported.
	ErrInvalidImageFormat = errors.New("invalid image format")
)
