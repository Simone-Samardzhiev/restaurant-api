package service

import (
	"context"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/google/uuid"
)

// ProductService implements port.ProductService and provides access to product-related bussiness logic.
type ProductService struct {
	productRepository port.ProductRepository
	imageRepository   port.ImageRepository
}

// NewProductService creates a new ProductService instance.
func NewProductService(productRepository port.ProductRepository, imageRepository port.ImageRepository) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		imageRepository:   imageRepository,
	}
}

func (s *ProductService) AddCategory(ctx context.Context, name string) error {
	id := uuid.New()
	return s.productRepository.AddCategory(ctx, domain.NewProductCategory(id, name))
}

func (s *ProductService) UpdateCategory(ctx context.Context, dto *domain.UpdateCategoryProductDTO) error {
	if dto.Name == nil {
		return domain.ErrNothingToUpdate
	}
	return s.productRepository.UpdateCategory(ctx, dto)
}

func (s *ProductService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.productRepository.DeleteCategory(ctx, id)
}

func (s *ProductService) AddProduct(ctx context.Context, dto *domain.AddProductDTO) error {
	id := uuid.New()

	if err := s.productRepository.
		AddProduct(
			ctx,
			domain.NewProduct(
				id,
				dto.Name,
				dto.Description,
				nil,
				dto.Category,
				dto.Price,
			),
		); err != nil {
		return err
	}
	return nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error {
	hasFieldToUpdate := false
	if dto.Name != nil {
		hasFieldToUpdate = true
	}
	if dto.Description != nil {
		hasFieldToUpdate = true
	}
	if dto.Category != nil {
		hasFieldToUpdate = true
	}
	if dto.Price != nil {
		hasFieldToUpdate = true
	}

	if !hasFieldToUpdate {
		return domain.ErrNothingToUpdate
	}
	return s.productRepository.UpdateProduct(ctx, dto)
}

func (s *ProductService) AddImage(ctx context.Context, image *domain.Image, productId uuid.UUID) error {
	path, err := s.imageRepository.Save(ctx, image, productId)
	if err != nil {
		return err
	}

	return s.productRepository.UpdateProductImagePath(ctx, productId, &path)
}

func (s *ProductService) DeleteProduct(ctx context.Context, dto *domain.DeleteProductDTO) error {
	switch {
	case dto.ProductId != nil:
		product, err := s.productRepository.DeleteProductById(ctx, *dto.ProductId)
		if err != nil {
			return err
		}

		if product.ImagePath != nil {
			if err = s.imageRepository.Delete(ctx, *product.ImagePath); err != nil {
				return err
			}
		}
		return nil

	case dto.CategoryId != nil:
		products, err := s.productRepository.DeleteProductsByCategory(ctx, *dto.CategoryId)
		if err != nil {
			return err
		}

		for _, product := range products {
			if product.ImagePath != nil {
				if err = s.imageRepository.Delete(ctx, *product.ImagePath); err != nil {
					return err
				}
			}
		}
		return nil

	default:
		return domain.ErrNothingToUpdate
	}
}

func (s *ProductService) GetProductCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	return s.productRepository.GetProductCategories(ctx)
}
