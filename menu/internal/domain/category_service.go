package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DefaultCategoryService is the default implementation of [CategoryService].
type DefaultCategoryService struct {
	repository CategoryRepository
}

var _ CategoryService = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService creates and allocates new [DefaultCategoryService].
func NewDefaultCategoryService(repository CategoryRepository) *DefaultCategoryService {
	return &DefaultCategoryService{repository: repository}
}

func (d *DefaultCategoryService) Add(ctx context.Context, name string) (*Category, error) {
	now := time.Now()

	category := Category{
		Id:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := d.repository.Save(ctx, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (d *DefaultCategoryService) GetAll(ctx context.Context) ([]Category, error) {
	return d.repository.GetAll(ctx)
}

func (d *DefaultCategoryService) Update(ctx context.Context, id uuid.UUID, name string) error {
	return d.repository.Update(ctx, id, name)
}
