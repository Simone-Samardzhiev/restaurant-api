package menu

import (
	"context"

	"github.com/google/uuid"
)

// CategoryRepository describes how categories are stored.
type CategoryRepository interface {
	// AddCategory saves a category.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error

	// DeleteCategory deletes a category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// CategoryService describes how category business logic is accessed.
type CategoryService interface {
	// AddCategory saves a category.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory applies updates to category.
	UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error

	// DeleteCategory deletes a category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
