package menu

import (
	"restaurant/internal/domain"

	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultProductService is the default implementation of CategoryService.
type DefaultProductService struct {
	productRepository ProductRepository
	imageRepository   ImageRepository
}

var _ ProductService = (*DefaultProductService)(nil)

// NewDefaultProductService creates a new DefaultProductService.
func NewDefaultProductService(productRepository ProductRepository, imageRepository ImageRepository) *DefaultProductService {
	return &DefaultProductService{
		productRepository: productRepository,
		imageRepository:   imageRepository,
	}
}

func (s *DefaultProductService) AddProduct(ctx context.Context, request *AddProductRequest) (*Product, error) {
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

	if err = s.productRepository.AddProduct(ctx, product); err != nil {
		if deleteErr := s.imageRepository.DeleteImage(ctx, imagePath); deleteErr != nil {
			zap.L().Error("error cleaning up image", zap.Error(deleteErr))
		}

		return nil, err
	}

	return product, nil
}

func (s *DefaultProductService) UpdateProduct(ctx context.Context, request *UpdateProductRequest) error {
	update, err := NewProductUpdate(request.Id, request.NewName, request.NewDescription, request.NewCategoryId, request.NewPrice)
	if err != nil {
		return err
	}

	return s.productRepository.UpdateProduct(ctx, update)
}

func (s *DefaultProductService) ReplaceProductImage(ctx context.Context, request *ReplaceProductImageRequest) (string, error) {
	imageType, err := NewImageType(request.ImageType)
	if err != nil {
		errors := domain.NewValidationErrors("invalid image")
		errors.Add("imageType", err)
		return "", errors
	}

	oldPath, err := s.productRepository.GetProductImagePathById(ctx, request.Id)
	if err != nil {
		return "", err
	}

	if err = s.imageRepository.DeleteImage(ctx, oldPath); err != nil {
		return "", domain.NewInternalError("error deleting old image", err)
	}

	newPath, err := s.imageRepository.AddImage(ctx, request.ImageData, imageType)
	if err != nil {
		return "", err
	}

	if err = s.productRepository.UpdateProductImagePath(ctx, request.Id, newPath); err != nil {
		if err = s.imageRepository.DeleteImage(ctx, newPath); err != nil {
			zap.L().Error("error cleaning up image", zap.String("path", newPath), zap.Error(err))
		}

		return "", domain.NewInternalError("error updating product image", err)
	}

	return newPath, nil
}

func (s *DefaultProductService) DeleteOrphanImages(ctx context.Context) {
	productsFilePaths, err := s.productRepository.GetProductImagePaths(ctx)
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

func (s *DefaultProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	path, err := s.productRepository.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}

	if err = s.imageRepository.DeleteImage(ctx, path); err != nil {
		return domain.NewInternalError("error cleaning up image", err)
	}

	return nil
}

func (s *DefaultProductService) GetProducts(ctx context.Context) ([]Product, error) {
	return s.productRepository.GetProducts(ctx)

}
