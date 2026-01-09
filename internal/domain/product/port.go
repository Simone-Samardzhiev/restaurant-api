package product

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

// Repository describes how product data is stored and managed.
type Repository interface {
	// AddCategory stores a new Category.
	AddCategory(ctx context.Context, category *Category) error

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, update *CategoryUpdate) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// Service describes product-related business logic.
type Service interface {
	// AddCategory creates and stores a new Category from AddCategoryRequest.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error

	// DeleteCategory deletes an existing Category by id.
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
