package menu

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

// DefaultCategoryService is the default implementation of CategoryService.
type DefaultCategoryService struct {
	repository CategoryRepository
}

var _ CategoryService = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService creates a new DefaultCategoryService.
func NewDefaultCategoryService(repository CategoryRepository) *DefaultCategoryService {
	return &DefaultCategoryService{
		repository: repository,
	}
}

func (s *DefaultCategoryService) AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error) {
	category, err := NewCategory(uuid.New(), request.Name)
	if err != nil {
		return nil, err
	}

	return category, s.repository.AddCategory(ctx, category)
}

func (s *DefaultCategoryService) UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error {
	update, err := NewCategoryUpdate(request.Id, request.NewName)
	if err != nil {
		return err
	}

	return s.repository.UpdateCategory(ctx, update)
}

func (s *DefaultCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteCategory(ctx, id)
}

func (s *DefaultCategoryService) GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error) {
	return s.repository.GetCategories(ctx, filter)
}
