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

func (s *ProductService) AddProduct(ctx context.Context, dto *domain.AddProductDTO) error {
	id := uuid.New()
	imagePath, err := s.imageRepository.Save(ctx, dto.Image, id)
	if err != nil {
		return err
	}

	if err = s.productRepository.
		AddProduct(
			ctx,
			domain.NewProduct(
				id,
				dto.Name,
				dto.Description,
				imagePath,
				dto.Category,
				dto.Price,
			),
		); err != nil {
		return err
	}
	return nil
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
