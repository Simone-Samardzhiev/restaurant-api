package domain

import "context"

// CategoryRepository describes how category data is accessed.
type CategoryRepository interface {

	// Save saves a new category.
	Save(ctx context.Context, category *Category) error
}
