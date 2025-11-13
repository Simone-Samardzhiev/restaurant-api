package service

import (
	"context"
	"io"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProductService implements port.ProductService and provides access to product-related business logic.
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

func (s *ProductService) AddCategory(ctx context.Context, name string) (*domain.ProductCategory, error) {
	category := domain.NewProductCategory(uuid.New(), name)
	if err := s.productRepository.AddCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
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

func (s *ProductService) GetProductCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	return s.productRepository.GetProductCategories(ctx)
}

func (s *ProductService) AddProduct(ctx context.Context, dto *domain.AddProductDTO) (*domain.Product, error) {
	product := domain.NewProduct(
		uuid.New(),
		dto.Name,
		dto.Description,
		nil,
		nil,
		dto.Category,
		dto.Price,
	)

	if err := s.productRepository.
		AddProduct(
			ctx,
			product,
		); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error {
	hasFieldToUpdate := false
	switch {
	case dto.Name != nil:
		hasFieldToUpdate = true
	case dto.Description != nil:
		hasFieldToUpdate = true
	case dto.Price != nil:
		hasFieldToUpdate = true
	case dto.Category != nil:
		hasFieldToUpdate = true
	}

	if !hasFieldToUpdate {
		return domain.ErrNothingToUpdate
	}
	return s.productRepository.UpdateProduct(ctx, dto)
}

func (s *ProductService) ReplaceProductImage(ctx context.Context, productId uuid.UUID, data io.Reader) (string, error) {
	product, err := s.productRepository.GetProductById(ctx, productId)
	if err != nil {
		return "", err
	}

	if product.DeleteImageUrl != nil {
		err = s.imageRepository.DeleteImage(ctx, *product.DeleteImageUrl)
		if err != nil {
			return "", err
		}
	}

	image, err := s.imageRepository.SaveImage(ctx, data)
	if err != nil {
		return "", err
	}

	if err = s.productRepository.UpdateProductImage(ctx, productId, image); err != nil {
		return "", err
	}

	return image.Url, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, dto *domain.DeleteProductDTO) error {
	switch {
	case dto.ProductId != nil:
		product, err := s.productRepository.DeleteProductById(ctx, *dto.ProductId)
		if err != nil {
			return err
		}

		if product.DeleteImageUrl == nil {
			return nil
		}
		if err = s.imageRepository.DeleteImage(ctx, *product.DeleteImageUrl); err != nil {
			return err
		}

		return nil
	case dto.CategoryId != nil:
		products, err := s.productRepository.DeleteProductsByCategory(ctx, *dto.CategoryId)
		if err != nil {
			return err
		}

		for _, product := range products {
			if product.DeleteImageUrl == nil {
				continue
			}

			if err = s.imageRepository.DeleteImage(ctx, *product.DeleteImageUrl); err != nil {
				zap.L().Error("error deleting image url", zap.Error(err))
			}
		}

		return nil
	default:
		return domain.ErrNothingToDelete
	}
}

func (s *ProductService) GetProducts(ctx context.Context, dto *domain.GetProductsDTO) ([]domain.Product, error) {
	if dto.CategoryId != nil {
		return s.productRepository.GetProductsByCategory(ctx, *dto.CategoryId)
	}

	return s.productRepository.GetProducts(ctx)
}
