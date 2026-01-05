package product

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

// DefaultService is the default implementation of Service.
type DefaultService struct {
	repository Repository
}

var _ Service = (*DefaultService)(nil)

// NewService creates a new DefaultService.
func NewService(repository Repository) *DefaultService {
	return &DefaultService{
		repository: repository,
	}
}

func (s *DefaultService) AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error) {
	category, err := NewCategory(uuid.New(), request.Name)
	if err != nil {
		return nil, err
	}

	return category, s.repository.AddCategory(ctx, category)
}

func (s *DefaultService) UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error {
	update, err := NewCategoryUpdate(request.Id, request.NewName)
	if err != nil {
		return err
	}

	return s.repository.UpdateCategory(ctx, update)
}
