package menu

import "context"

// CategoryRepository describes how categories are stored.
type CategoryRepository interface {
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)
}

// CategoryService describes how category business logic is accessed.
type CategoryService interface {
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)
}
