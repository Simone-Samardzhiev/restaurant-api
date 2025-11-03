package domain

import "errors"

var (
	// ErrInternal represents a generic internal service failure.
	ErrInternal = errors.New("internal failure")

	// ErrProductNameAlreadyInUse indicates a product name is already in use
	ErrProductNameAlreadyInUse = errors.New("product name is already in use")
)
