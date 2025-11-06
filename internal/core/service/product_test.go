package service_test

import (
	"context"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port/mock"
	"restaurant/internal/core/service"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
				Image:       nil,
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
				Image:       nil,
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
				Image:       nil,
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
