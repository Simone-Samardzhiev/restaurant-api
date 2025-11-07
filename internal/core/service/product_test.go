package service_test

import (
	"context"
	"io"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port/mock"
	"restaurant/internal/core/service"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductService_AddCategory(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name:     "success",
			category: "New Category",
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					AddCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.ProductCategory{}),
					).Return(nil)
			},
		}, {
			name:          "error category name already exists",
			category:      "Duplicate Category",
			expectedError: domain.ErrProductCategoryNameAlreadyInUse,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					AddCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.ProductCategory{}),
					).Return(domain.ErrProductCategoryNameAlreadyInUse)
			},
		}, {
			name:          "error adding category",
			category:      "New Category",
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					AddCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.ProductCategory{}),
					).Return(domain.ErrInternal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				AddCategory(context.Background(), tt.category)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_UpdateCategory(t *testing.T) {
	newName := "New Name"

	tests := []struct {
		name          string
		dto           *domain.UpdateCategoryProductDTO
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name: "success",
			dto: &domain.UpdateCategoryProductDTO{
				Id:   uuid.UUID{},
				Name: &newName,
			},
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateCategoryProductDTO{}),
					).Return(nil)
			},
		}, {
			name: "error category name already exists",
			dto: &domain.UpdateCategoryProductDTO{
				Id:   uuid.UUID{},
				Name: &newName,
			},
			expectedError: domain.ErrProductCategoryNameAlreadyInUse,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateCategoryProductDTO{}),
					).Return(domain.ErrProductCategoryNameAlreadyInUse)
			},
		}, {
			name: "error category not found",
			dto: &domain.UpdateCategoryProductDTO{
				Id:   uuid.UUID{},
				Name: &newName,
			},
			expectedError: domain.ErrProductCategoryNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateCategoryProductDTO{}),
					).Return(domain.ErrProductCategoryNotFound)
			},
		}, {
			name:          "error nothing to update",
			dto:           &domain.UpdateCategoryProductDTO{},
			expectedError: domain.ErrNothingToUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				UpdateCategory(context.Background(), tt.dto)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_DeleteCategory(t *testing.T) {
	tests := []struct {
		name          string
		categoryID    uuid.UUID
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name:       "success",
			categoryID: uuid.New(),
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					DeleteCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return(nil)
			},
		},
		{
			name:          "error category has linked products",
			categoryID:    uuid.New(),
			expectedError: domain.ErrCategoryHasLinkedProducts,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					DeleteCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return(domain.ErrCategoryHasLinkedProducts)
			},
		},
		{
			name:          "error category not found",
			categoryID:    uuid.New(),
			expectedError: domain.ErrProductCategoryNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					DeleteCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return(domain.ErrProductCategoryNotFound)
			},
		},
		{
			name:          "error deleting category",
			categoryID:    uuid.New(),
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					DeleteCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return(domain.ErrInternal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				DeleteCategory(context.Background(), tt.categoryID)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_AddProduct(t *testing.T) {
	tests := []struct {
		name          string
		dto           *domain.AddProductDTO
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name: "success",
			dto: &domain.AddProductDTO{
				Name:        "New Product",
				Description: "New Description",
				Category:    uuid.UUID{},
				Price:       decimal.NewFromFloat(1.33),
			},
			expectedError: nil,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("", nil),
					productRepository.EXPECT().
						AddProduct(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(&domain.Product{}),
						).Return(nil),
				)
			},
		}, {
			name: "error saving image",
			dto: &domain.AddProductDTO{
				Name:        "New Product",
				Description: "New Description",
				Category:    uuid.UUID{},
				Price:       decimal.NewFromFloat(1.33),
			},
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				imageRepository.EXPECT().
					Save(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.Any(),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return("", domain.ErrInternal)
			},
		}, {
			name: "error product already exists",
			dto: &domain.AddProductDTO{
				Name:        "Duplicate name",
				Description: "New Description",
				Category:    uuid.UUID{},
				Price:       decimal.NewFromFloat(1.33),
			},
			expectedError: domain.ErrProductNameAlreadyInUse,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("", nil),
					productRepository.EXPECT().
						AddProduct(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(&domain.Product{}),
						).Return(domain.ErrProductNameAlreadyInUse),
				)
			},
		}, {
			name: "error product already exists",
			dto: &domain.AddProductDTO{
				Name:        "Product Name",
				Description: "New Description",
				Category:    uuid.Nil,
				Price:       decimal.NewFromFloat(1.33),
			},
			expectedError: domain.ErrProductCategoryNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("", nil),
					productRepository.EXPECT().
						AddProduct(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(&domain.Product{}),
						).Return(domain.ErrProductCategoryNotFound),
				)
			},
		}, {
			name: "error adding product",
			dto: &domain.AddProductDTO{
				Name:        "New Product",
				Description: "New Description",
				Category:    uuid.UUID{},
				Price:       decimal.NewFromFloat(1.33),
			},
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("", nil),
					productRepository.EXPECT().
						AddProduct(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(&domain.Product{}),
						).Return(domain.ErrInternal),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				AddProduct(context.Background(), tt.dto)

			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	name := "New Product"
	tests := []struct {
		name          string
		dto           *domain.UpdateProductDTO
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name: "success",
			dto: &domain.UpdateProductDTO{
				Id:   uuid.UUID{},
				Name: &name,
			},
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateProduct(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateProductDTO{}),
					).
					Return(nil)
			},
		}, {
			name: "product not found",
			dto: &domain.UpdateProductDTO{
				Id:   uuid.UUID{},
				Name: &name,
			},
			expectedError: domain.ErrProductNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateProduct(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateProductDTO{}),
					).
					Return(domain.ErrProductNotFound)
			},
		}, {
			name: "product name already in use",
			dto: &domain.UpdateProductDTO{
				Id:   uuid.UUID{},
				Name: &name,
			},
			expectedError: domain.ErrProductNameAlreadyInUse,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateProduct(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateProductDTO{}),
					).
					Return(domain.ErrProductNameAlreadyInUse)
			},
		}, {
			name: "product category not found",
			dto: &domain.UpdateProductDTO{
				Id:   uuid.UUID{},
				Name: &name,
			},
			expectedError: domain.ErrProductCategoryNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateProduct(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateProductDTO{}),
					).
					Return(domain.ErrProductCategoryNotFound)
			},
		}, {
			name:          "nothing to update",
			dto:           &domain.UpdateProductDTO{},
			expectedError: domain.ErrNothingToUpdate,
		}, {
			name: "error updating product",
			dto: &domain.UpdateProductDTO{
				Id:   uuid.UUID{},
				Name: &name,
			},
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					UpdateProduct(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateProductDTO{}),
					).
					Return(domain.ErrInternal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				UpdateProduct(context.Background(), tt.dto)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_AddImage(t *testing.T) {
	tests := []struct {
		name          string
		image         io.Reader
		productId     uuid.UUID
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name:      "success",
			image:     nil,
			productId: uuid.UUID{},
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("path/to/image", nil),
					productRepository.EXPECT().
						UpdateProductImagePath(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(uuid.UUID{}),
							gomock.Any(),
						).
						Return(nil),
				)
			},
		}, {
			name:          "error product not found",
			image:         nil,
			productId:     uuid.UUID{},
			expectedError: domain.ErrProductNotFound,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("path/to/image", nil),
					productRepository.EXPECT().
						UpdateProductImagePath(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(uuid.UUID{}),
							gomock.Any(),
						).
						Return(domain.ErrProductNotFound),
				)
			},
		}, {
			name:          "error saving image",
			image:         nil,
			productId:     uuid.UUID{},
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("", domain.ErrInternal),
				)
			},
		}, {
			name:          "error updating product",
			image:         nil,
			productId:     uuid.UUID{},
			expectedError: domain.ErrInternal,
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					imageRepository.EXPECT().
						Save(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.Any(),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).
						Return("path/to/image", nil),
					productRepository.EXPECT().
						UpdateProductImagePath(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(uuid.UUID{}),
							gomock.Any(),
						).
						Return(domain.ErrInternal),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			productRepository := mock.NewMockProductRepository(ctrl)
			imageRepository := mock.NewMockImageRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(productRepository, imageRepository)
			}

			err := service.NewProductService(productRepository, imageRepository).
				AddImage(context.Background(), tt.image, tt.productId)

			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}
