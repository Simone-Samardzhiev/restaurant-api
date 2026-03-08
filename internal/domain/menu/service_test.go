package menu_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeImageRepository struct {
	onSaveImage        func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error)
	onDeleteImage      func(ctx context.Context, path string) error
	deleteImageCounter int
}

var _ menu.ImageRepository = (*fakeImageRepository)(nil)

func (r *fakeImageRepository) SaveImage(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
	if r.onSaveImage == nil {
		panic("onSaveImage function is not implemented")
	}
	return r.onSaveImage(ctx, data, imageType)
}

func (r *fakeImageRepository) DeleteImage(ctx context.Context, path string) error {
	if r.onDeleteImage == nil {
		panic("onDeleteImage function is not implemented")
	}
	r.deleteImageCounter++
	return r.onDeleteImage(ctx, path)
}

type fakeProductRepository struct {
	onSaveProduct     func(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error)
	onUpdateImagePath func(ctx context.Context, id uuid.UUID, path string) (string, error)
	onDeleteProduct   func(ctx context.Context, id uuid.UUID) (string, error)
}

var _ menu.ProductRepository = (*fakeProductRepository)(nil)

func (r *fakeProductRepository) SaveProduct(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error) {
	if r.onSaveProduct == nil {
		panic("onSaveProduct function is not implemented")
	}
	return r.onSaveProduct(ctx, request)
}

func (r *fakeProductRepository) UpdateProduct(_ context.Context, _ *menu.UpdateProductRequest) error {
	panic("implement me")
}

func (r *fakeProductRepository) UpdateImagePath(ctx context.Context, id uuid.UUID, path string) (string, error) {
	if r.onUpdateImagePath == nil {
		panic("onUpdateImagePath function is not implemented")
	}
	return r.onUpdateImagePath(ctx, id, path)
}

func (r *fakeProductRepository) DeleteProduct(ctx context.Context, id uuid.UUID) (string, error) {
	if r.onDeleteProduct == nil {
		panic("onDeleteProduct function is not implemented")
	}
	return r.onDeleteProduct(ctx, id)
}

func (r *fakeProductRepository) GetProducts(_ context.Context, _ *menu.ProductFilter) ([]menu.Product, error) {
	panic("implement me")
}

func checkAddProductResult(
	t *testing.T,
	product *menu.Product,
	request *menu.AddProductRequest,
) {
	t.Helper()

	if request.Name.String() != product.Name.String() {
		t.Errorf("want name %v, got %v", product.Name, request.Name)
	}
	if request.Description.String() != product.Description.String() {
		t.Errorf("want description %v, got %v", product.Description, request.Description)
	}
	if !request.Price.Value().Equal(product.Price.Value()) {
		t.Errorf("want price %v, got %v", product.Price, request.Price)
	}
	if request.CategoryId != product.CategoryId {
		t.Errorf("want category id %v, got %v", product.CategoryId, request.CategoryId)
	}
}

func TestDefaultProductServiceAddProduct(t *testing.T) {
	tests := []struct {
		name                 string
		request              *menu.AddProductRequest
		imageRepository      *fakeImageRepository
		productRepository    *fakeProductRepository
		wantErr              bool
		wantErrorKind        domain.ErrorKind
		wantErrorCode        domain.ErrorCode
		wantDeleteImageCount int
	}{
		{
			name: "success",
			request: test.Must(menu.ParseAddProductRequest(
				"Valid product name",
				"Valid product description",
				decimal.NewFromFloat(10.5),
				uuid.New(),
				strings.NewReader("image data"),
				"png",
			)),
			imageRepository: &fakeImageRepository{
				onSaveImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return filepath.Join("save", uuid.NewString()+"."+imageType.String()), nil
				},
			},
			productRepository: &fakeProductRepository{
				onSaveProduct: func(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error) {
					return &menu.Product{
						Id:          uuid.New(),
						Name:        request.Name,
						Description: request.Description,
						Price:       request.Price,
						CategoryId:  request.CategoryId,
						ImagePath:   request.ImagePath,
					}, nil
				},
			},
		},
		{
			name: "error saving product",
			request: test.Must(menu.ParseAddProductRequest(
				"Valid product name",
				"Valid product description",
				decimal.NewFromFloat(10.5),
				uuid.New(),
				strings.NewReader("image data"),
				"png",
			)),
			imageRepository: &fakeImageRepository{
				onSaveImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return filepath.Join("save", uuid.NewString()+"."+imageType.String()), nil
				},
				onDeleteImage: func(ctx context.Context, path string) error {
					return nil
				},
			},
			productRepository: &fakeProductRepository{
				onSaveProduct: func(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error) {
					return nil, domain.NewInternalError("error saving product", errors.New("postgres not connected"))
				},
			},
			wantErr:              true,
			wantErrorKind:        domain.ErrorKindInternal,
			wantErrorCode:        domain.ErrorCodeInternal,
			wantDeleteImageCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.productRepository, tt.imageRepository)
			product, err := service.AddProduct(context.Background(), tt.request)

			if tt.wantDeleteImageCount != tt.imageRepository.deleteImageCounter {
				t.Errorf("want delete image count %d, got %d", tt.wantDeleteImageCount, tt.imageRepository.deleteImageCounter)
			}

			if tt.wantErr {
				if tt.wantDeleteImageCount != tt.imageRepository.deleteImageCounter {
					t.Errorf("want delete image called %d, got %d", tt.wantDeleteImageCount, tt.imageRepository.deleteImageCounter)
				}
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}

			checkAddProductResult(t, product, tt.request)
		})
	}
}

func TestDefaultProductServiceUpdateImage(t *testing.T) {
	tests := []struct {
		name                 string
		request              *menu.UpdateImageRequest
		productRepository    *fakeProductRepository
		imageRepository      *fakeImageRepository
		wantDeleteImageCount int
		wantPath             string
		wantErr              bool
		wantErrorKind        domain.ErrorKind
		wantErrorCode        domain.ErrorCode
		wantDetailsCodes     []domain.ErrorCode
	}{
		{
			name:    "success",
			request: menu.MustParseUpdateImageRequest(uuid.New(), nil, "png"),
			productRepository: &fakeProductRepository{
				onUpdateImagePath: func(ctx context.Context, id uuid.UUID, path string) (string, error) {
					return "oldPath", nil
				},
			},
			imageRepository: &fakeImageRepository{
				onSaveImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return "newPath", nil
				},
				onDeleteImage: func(ctx context.Context, path string) error {
					return nil
				},
			},
			wantDeleteImageCount: 1,
			wantPath:             "newPath",
		},
		{
			name:    "product not found",
			request: menu.MustParseUpdateImageRequest(uuid.New(), nil, "png"),
			productRepository: &fakeProductRepository{
				onUpdateImagePath: func(ctx context.Context, id uuid.UUID, path string) (string, error) {
					return "", domain.NewNotFoundError(
						"product not found",
						domain.ErrorCodeProductNotFound,
						domain.ErrorDetail{
							Code:     domain.ErrorCodeProductNotFoundByID,
							Message:  "product not found by id",
							Metadata: map[string]interface{}{"id": id},
						},
					)
				},
			},
			imageRepository: &fakeImageRepository{
				onSaveImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return "newPath", nil
				},
				onDeleteImage: func(ctx context.Context, path string) error {
					return nil
				},
			},
			wantDeleteImageCount: 1,
			wantErr:              true,
			wantErrorKind:        domain.ErrorKindNotFound,
			wantErrorCode:        domain.ErrorCodeProductNotFound,
			wantDetailsCodes:     []domain.ErrorCode{domain.ErrorCodeProductNotFoundByID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.productRepository, tt.imageRepository)
			path, err := service.UpdateImage(context.Background(), tt.request)
			if tt.wantDeleteImageCount != tt.imageRepository.deleteImageCounter {
				t.Errorf("want delete image count %d, got %d", tt.wantDeleteImageCount, tt.imageRepository.deleteImageCounter)
			}

			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			if tt.wantPath != path {
				t.Errorf("want path %s, got %s", tt.wantPath, path)
			}
		})
	}
}

func TestDefaultProductServiceDeleteProduct(t *testing.T) {
	tests := []struct {
		name                 string
		productRepository    *fakeProductRepository
		imageRepository      *fakeImageRepository
		wantDeleteImageCount int
		wantErr              bool
		wantErrorKind        domain.ErrorKind
		wantErrorCode        domain.ErrorCode
		wantDetailsCodes     []domain.ErrorCode
	}{
		{
			name: "success",
			productRepository: &fakeProductRepository{
				onDeleteProduct: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "oldPath", nil
				},
			},
			imageRepository: &fakeImageRepository{
				onDeleteImage: func(ctx context.Context, path string) error {
					return nil
				},
			},
			wantDeleteImageCount: 1,
		},
		{
			name: "product not found",
			productRepository: &fakeProductRepository{
				onDeleteProduct: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "", domain.NewNotFoundError(
						"product not found",
						domain.ErrorCodeProductNotFound,
						domain.ErrorDetail{
							Code:     domain.ErrorCodeProductNotFoundByID,
							Message:  "product not found by id",
							Metadata: map[string]interface{}{"id": id},
						},
					)
				},
			},
			imageRepository:      &fakeImageRepository{},
			wantDeleteImageCount: 0,
			wantErr:              true,
			wantErrorKind:        domain.ErrorKindNotFound,
			wantErrorCode:        domain.ErrorCodeProductNotFound,
			wantDetailsCodes:     []domain.ErrorCode{domain.ErrorCodeProductNotFoundByID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.productRepository, tt.imageRepository)
			err := service.DeleteProduct(context.Background(), uuid.New())
			if tt.wantDeleteImageCount != tt.imageRepository.deleteImageCounter {
				t.Errorf("want delete image count %d, got %d", tt.wantDeleteImageCount, tt.imageRepository.deleteImageCounter)
			}

			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
		})
	}

}
