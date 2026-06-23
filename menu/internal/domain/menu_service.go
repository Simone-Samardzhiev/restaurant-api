package domain

import (
	"context"

	"github.com/google/uuid"
)

// DefaultMenuService is the default implementation of [MenuService].
type DefaultMenuService struct {
	categoryRepository CategoryRepository
	productRepository  ProductRepository
}

var _ MenuService = (*DefaultMenuService)(nil)

// NewDefaultMenuService creates and allocates new [DefaultMenuService].
func NewDefaultMenuService(categoryRepository CategoryRepository, productRepository ProductRepository) *DefaultMenuService {
	return &DefaultMenuService{
		categoryRepository: categoryRepository,
		productRepository:  productRepository,
	}
}

func (d *DefaultMenuService) Get(ctx context.Context) (Menu, error) {
	categories, err := d.categoryRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	products, err := d.productRepository.GetAllWithImage(ctx)
	if err != nil {
		return nil, err
	}

	categoryToProducts := make(map[uuid.UUID][]Product, len(categories))
	for _, p := range products {
		categoryToProducts[p.CategoryId] = append(categoryToProducts[p.CategoryId], p)
	}

	menu := make([]MenuSection, 0, len(categoryToProducts))
	for _, category := range categories {
		p := categoryToProducts[category.Id]
		if p == nil {
			p = []Product{}
		}

		menu = append(menu, MenuSection{
			Category: category,
			Products: p,
		})
	}
	return menu, nil
}
