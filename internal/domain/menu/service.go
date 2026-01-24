package menu

import (
	"context"
	"restaurant/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultService is the default implementation of Service.
type DefaultService struct {
	menuRepository  Repository
	imageRepository ImageRepository
}

var _ Service = (*DefaultService)(nil)

// NewService creates a new DefaultService.
func NewService(menuRepository Repository, imageRepository ImageRepository) *DefaultService {
	return &DefaultService{
		menuRepository:  menuRepository,
		imageRepository: imageRepository,
	}
}

func (s *DefaultService) AddCategory(ctx context.Context, request *AddCategoryRequest) (*Category, error) {
	category, err := NewCategory(uuid.New(), request.Name)
	if err != nil {
		return nil, err
	}

	return category, s.menuRepository.AddCategory(ctx, category)
}

func (s *DefaultService) UpdateCategory(ctx context.Context, request *CategoryUpdateRequest) error {
	update, err := NewCategoryUpdate(request.Id, request.NewName)
	if err != nil {
		return err
	}

	return s.menuRepository.UpdateCategory(ctx, update)
}

func (s *DefaultService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.menuRepository.DeleteCategory(ctx, id)
}

func (s *DefaultService) GetCategories(ctx context.Context, filter *CategoryFilter) ([]Category, error) {
	return s.menuRepository.GetCategories(ctx, filter)
}

func (s *DefaultService) AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error) {
	imageType, err := NewImageType(request.ImageType)
	if err != nil {
		errors := domain.NewValidationErrors("invalid image")
		errors.Add("imageType", err)
		return nil, errors
	}

	validationErrors := domain.NewValidationErrors("invalid product")
	parsedName, err := NewProductName(request.Name)
	if err != nil {
		validationErrors.Add("name", err)
	}

	parsedDescription, err := NewProductDescription(request.Description)
	if err != nil {
		validationErrors.Add("description", err)
	}

	parsedPrice, err := NewProductPrice(request.Price)
	if err != nil {
		validationErrors.Add("price", err)
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	imagePath, err := s.imageRepository.AddImage(ctx, request.ImageData, imageType)
	if err != nil {
		return nil, domain.NewInternalError("error adding image", err)
	}

	product := NewProductWithValidFields(uuid.New(), parsedName, parsedDescription, request.CategoryId, parsedPrice, imagePath)

	if err = s.menuRepository.AddProduct(ctx, product); err != nil {
		if deleteErr := s.imageRepository.DeleteImage(ctx, imagePath); deleteErr != nil {
			zap.L().Error("error cleaning up image", zap.Error(deleteErr))
		}

		return nil, err
	}

	return product, nil
}

func (s *DefaultService) DeleteOrphanImages(ctx context.Context) {
	productsFilePaths, err := s.menuRepository.GetProductImagePaths(ctx)
	if err != nil {
		zap.L().Warn("error getting product image paths", zap.Error(err))
		return
	}

	imageFilePaths, err := s.imageRepository.GetAllImagePaths(ctx)
	if err != nil {
		zap.L().Warn("error getting product image paths", zap.Error(err))
		return
	}

	for path := range imageFilePaths {
		_, ok := productsFilePaths[path]
		if ok {
			continue
		}

		if err = s.imageRepository.DeleteImage(ctx, path); err != nil {
			zap.L().Error("error cleaning up image", zap.String("path", path))
		}
	}
}
