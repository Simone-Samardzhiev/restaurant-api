package service_test

import (
	"context"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port/mock"
	"restaurant/internal/core/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
			name:          "nothing to update",
			dto:           &domain.UpdateProductDTO{},
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
				UpdateProduct(context.Background(), tt.dto)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name          string
		dto           *domain.DeleteProductDTO
		expectedError error
		mockSetup     func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name: "success",
			dto: &domain.DeleteProductDTO{
				ProductId:  &uuid.UUID{},
				CategoryId: &uuid.UUID{},
			},
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				gomock.InOrder(
					productRepository.EXPECT().
						DeleteProductById(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(uuid.UUID{}),
						).Return(
						&domain.Product{
							DeleteImageUrl: new(string),
						},
						nil,
					),

					imageRepository.EXPECT().
						DeleteImage(
							gomock.AssignableToTypeOf(context.Background()),
							gomock.AssignableToTypeOf(""),
						).
						Return(nil),
				)
			},
		}, {
			name:          "error nothing to delete",
			dto:           &domain.DeleteProductDTO{},
			expectedError: domain.ErrNothingToDelete,
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
				DeleteProduct(context.Background(), tt.dto)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestProductService_GetProductById(t *testing.T) {
	tests := []struct {
		name            string
		dto             *domain.GetProductsDTO
		expectedProduct []domain.Product
		expectedError   error
		mockSetup       func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository)
	}{
		{
			name: "success",
			dto: &domain.GetProductsDTO{
				CategoryId: &uuid.UUID{},
			},
			expectedProduct: []domain.Product{
				{
					Id:   uuid.UUID{},
					Name: "fetched product",
				},
			},
			mockSetup: func(productRepository *mock.MockProductRepository, imageRepository *mock.MockImageRepository) {
				productRepository.EXPECT().
					GetProductsByCategory(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(uuid.UUID{}),
					).Return([]domain.Product{
					{
						Id:   uuid.UUID{},
						Name: "fetched product",
					},
				}, nil,
				)
			},
		}, {
			name:          "error nothing to fetch",
			dto:           &domain.GetProductsDTO{},
			expectedError: domain.ErrNothingToFetch,
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

			res, err := service.NewProductService(productRepository, imageRepository).
				GetProducts(context.Background(), tt.dto)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedProduct, res)
		})
	}
}
