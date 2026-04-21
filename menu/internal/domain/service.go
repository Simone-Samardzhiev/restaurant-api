package domain

import "context"

// CategoryService describes how category business logic is accessed.
type CategoryService interface {

	// Add creates a new category with specific name.
	Add(ctx context.Context, name string) (*Category, error)
}
