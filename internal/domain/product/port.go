package product

import "golang.org/x/net/context"

// Repository describes how product data is stored and managed.
type Repository interface {
	// AddCategory stores a new Category.
	AddCategory(ctx context.Context, category *Category) error

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, update *CategoryUpdate) error
}

// Service describes product-related business logic.
type Service interface {
	// AddCategory creates and stores a new Category from AddCategoryRequest.
	AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error)

	// UpdateCategory updates an existing Category.
	UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error
}
