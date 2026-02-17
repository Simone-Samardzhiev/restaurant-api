package menu

import "context"

// CategoryRepository describes how categories are stored.
type CategoryRepository interface {
	// AddCategory saves a category.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error
}

// CategoryService describes how category business logic is accessed.
type CategoryService interface {
	// AddCategory saves a category.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error
}
