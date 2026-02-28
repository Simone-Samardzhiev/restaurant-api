package menu

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultCategoryService is the default implementation of [CategoryService].
type DefaultCategoryService struct {
	repository CategoryRepository
}

var _ CategoryService = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService allocates and creates a new [DefaultCategoryService].
func NewDefaultCategoryService(repository CategoryRepository) *DefaultCategoryService {
	return &DefaultCategoryService{repository: repository}
}

func (s *DefaultCategoryService) AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error) {
	return s.repository.SaveCategory(ctx, request)
}

func (s *DefaultCategoryService) UpdateCategory(ctx context.Context, request *UpdateCategoryRequest) error {
	return s.repository.UpdateCategory(ctx, request)
}

func (s *DefaultCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteCategory(ctx, id)
}

func (s *DefaultCategoryService) GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error) {
	return s.repository.GetCategories(ctx, filter)
}

// DefaultProductService is the default implementation of [ProductService].
type DefaultProductService struct {
	productRepository ProductRepository
	imageRepository   ImageRepository
}

var _ ProductService = (*DefaultProductService)(nil)

// NewDefaultProductService allocates and creates a new [DefaultProductService].
func NewDefaultProductService(productRepository ProductRepository, imageRepository ImageRepository) *DefaultProductService {
	return &DefaultProductService{
		productRepository: productRepository,
		imageRepository:   imageRepository,
	}
}

func (s *DefaultProductService) AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error) {
	imagePath, err := s.imageRepository.SaveImage(ctx, request.ImageData, request.ImageType)
	if err != nil {
		return nil, err
	}

	saveRequest := NewSaveProductRequest(request.Name, request.Description, request.Price, request.CategoryId, imagePath)
	product, err := s.productRepository.SaveProduct(ctx, saveRequest)
	if err != nil {
		if deleteErr := s.imageRepository.DeleteImage(ctx, imagePath); deleteErr != nil {
			zap.L().Warn("error cleaning up image", zap.String("imagePath", imagePath), zap.Error(deleteErr))
		}

		return nil, err
	}

	return product, nil
}

func (s *DefaultProductService) UpdateProduct(ctx context.Context, request *UpdateProductRequest) error {
	return s.productRepository.UpdateProduct(ctx, request)
}

func (s *DefaultProductService) UpdateImage(ctx context.Context, request *UpdateImageRequest) (string, error) {
	newPath, err := s.imageRepository.SaveImage(ctx, request.Data, request.ImageType)
	if err != nil {
		return "", err
	}

	oldPath, err := s.productRepository.UpdateImagePath(ctx, request.Id, newPath)
	if err != nil {
		if deleteErr := s.imageRepository.DeleteImage(ctx, newPath); deleteErr != nil {
			zap.L().Warn("error cleaning up image", zap.String("imagePath", newPath), zap.Error(deleteErr))
		}

		return "", err
	}

	if err = s.imageRepository.DeleteImage(ctx, oldPath); err != nil {
		zap.L().Warn("error cleaning up image", zap.String("imagePath", oldPath), zap.Error(err))
	}

	return newPath, nil
}
