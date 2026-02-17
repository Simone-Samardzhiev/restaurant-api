package menu

import (
	"context"

	"github.com/google/uuid"
)

// DefaultCategoryService is the default implementation of [CategoryService].
type DefaultCategoryService struct {
	repository CategoryRepository
}

var _ CategoryService = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService allocates and creates a new DefaultCategoryService.
func NewDefaultCategoryService(repository CategoryRepository) *DefaultCategoryService {
	return &DefaultCategoryService{repository: repository}
}

func (s *DefaultCategoryService) AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error) {
	return s.repository.AddCategory(ctx, request)
}

func (s *DefaultCategoryService) UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error {
	return s.repository.UpdateCategory(ctx, request)
}

func (s *DefaultCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteCategory(ctx, id)
}
