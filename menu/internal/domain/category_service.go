package domain

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// DefaultCategoryService is the default implementation of [CategoryService].
type DefaultCategoryService struct {
	repository CategoryRepository
	purger     CachePurger
	logger     *slog.Logger
}

var _ CategoryService = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService creates and allocates new [DefaultCategoryService].
func NewDefaultCategoryService(repository CategoryRepository, purger CachePurger, logger *slog.Logger) *DefaultCategoryService {
	return &DefaultCategoryService{repository: repository, purger: purger, logger: logger}
}

func (d *DefaultCategoryService) purgeCache(ctx context.Context) {
	go func() {
		gtCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
		if err := d.purger.Categories(gtCtx); err != nil {
			d.logger.LogAttrs(ctx, slog.LevelWarn, "Error purging categories cache", slog.String("error", err.Error()))
		}
	}()
}

func (d *DefaultCategoryService) Add(ctx context.Context, name string) (*Category, error) {
	now := time.Now()
	categoryId, err := uuid.NewV7()
	if err != nil {
		return nil, NewError("error create uuid for category", ErrorCodeInternal, err)
	}

	category := Category{
		Id:        categoryId,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := d.repository.Save(ctx, &category); err != nil {
		return nil, err
	}
	d.purgeCache(ctx)
	return &category, nil
}

func (d *DefaultCategoryService) GetAll(ctx context.Context) ([]Category, error) {
	return d.repository.GetAll(ctx)
}

func (d *DefaultCategoryService) Update(ctx context.Context, id uuid.UUID, name string) error {
	if err := d.repository.Update(ctx, id, name); err != nil {
		return err
	}
	d.purgeCache(ctx)
	return nil
}

func (d *DefaultCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := d.repository.Delete(ctx, id); err != nil {
		return err
	}
	d.purgeCache(ctx)
	return nil
}
