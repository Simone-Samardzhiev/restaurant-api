package domain

import "errors"

var (
	// ErrInternal represents a generic internal service failure.
	ErrInternal = errors.New("internal failure")

	// ErrNothingToUpdate indicates an updates request won't change any data.
	ErrNothingToUpdate = errors.New("nothing to update")

	// ErrNothingToDelete indicates a delete request won't delete any data.
	ErrNothingToDelete = errors.New("nothing to delete")

	// ErrMultipleDeleteCriteria indicates that more than one delete criteria is provided leading to inconsistent data.
	ErrMultipleDeleteCriteria = errors.New("multiple delete criteria found")

	// ErrNothingToFetch indicates a fetch request won't fetch any data.
	ErrNothingToFetch = errors.New("nothing to fetch")

	// ErrInvalidUUID indicates an id is not valid uuid.
	ErrInvalidUUID = errors.New("invalid entity")

	// ErrInvalidImageFormat indicates provided image format is not valid.
	ErrInvalidImageFormat = errors.New("invalid image format")

	// ErrProductCategoryNameAlreadyInUse indicates a product category name is already in use.
	ErrProductCategoryNameAlreadyInUse = errors.New("product category is already in use")

	// ErrProductCategoryNotFound indicates a product category couldn't be found.
	ErrProductCategoryNotFound = errors.New("product category not found")

	// ErrProductNameAlreadyInUse indicates a product name is already in use.
	ErrProductNameAlreadyInUse = errors.New("product name is already in use")

	// ErrProductNotFound indicates a product couldn't be found.
	ErrProductNotFound = errors.New("product not found")

	// ErrCategoryHasLinkedProducts indicates an attempt to delete a product category that has linked products.
	ErrCategoryHasLinkedProducts = errors.New("category has linked products")

	// ErrOrderSessionNotFound indicates order session couldn't be found.
	ErrOrderSessionNotFound = errors.New("order session not found")
)
