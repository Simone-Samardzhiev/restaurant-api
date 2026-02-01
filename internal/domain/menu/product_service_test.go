package menu_test

import (
	"fmt"
	"io"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FakeProductRepository struct {
	OnAddProduct              func(ctx context.Context, product *menu.Product) error
	OnUpdateProduct           func(ctx context.Context, product *menu.ProductUpdate) error
	OnUpdateProductImagePath  func(ctx context.Context, id uuid.UUID, path string) error
	OnGetProductImagePathById func(ctx context.Context, id uuid.UUID) (string, error)
	OnGetProductImagePaths    func(ctx context.Context) (map[string]struct{}, error)
	OnGetProducts             func(ctx context.Context) ([]menu.Product, error)
	OnDeleteProduct           func(ctx context.Context, id uuid.UUID) (string, error)
}

var _ menu.ProductRepository = (*FakeProductRepository)(nil)

func (r *FakeProductRepository) AddProduct(ctx context.Context, product *menu.Product) error {
	if r.OnAddProduct != nil {
		return r.OnAddProduct(ctx, product)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) UpdateProduct(ctx context.Context, update *menu.ProductUpdate) error {
	if r.OnUpdateProduct != nil {
		return r.OnUpdateProduct(ctx, update)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) UpdateProductImagePath(ctx context.Context, id uuid.UUID, path string) error {
	if r.OnUpdateProductImagePath != nil {
		return r.OnUpdateProductImagePath(ctx, id, path)
	}
	return fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) GetProductImagePathById(ctx context.Context, id uuid.UUID) (string, error) {
	if r.OnGetProductImagePathById != nil {
		return r.OnGetProductImagePathById(ctx, id)
	}
	return "", fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) GetProductImagePaths(ctx context.Context) (map[string]struct{}, error) {
	if r.OnGetProductImagePaths != nil {
		return r.OnGetProductImagePaths(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) GetProducts(ctx context.Context) ([]menu.Product, error) {
	if r.OnGetProducts != nil {
		return r.OnGetProducts(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *FakeProductRepository) DeleteProduct(ctx context.Context, id uuid.UUID) (string, error) {
	if r.OnDeleteProduct != nil {
		return r.OnDeleteProduct(ctx, id)
	}
	return "", fmt.Errorf("not implemented")
}

type FakeImageRepository struct {
	OnAddImage         func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error)
	OnDeleteImage      func(ctx context.Context, path string) error
	OnGetAllImagePaths func(ctx context.Context) (map[string]struct{}, error)
}

var _ menu.ImageRepository = (*FakeImageRepository)(nil)

func (r *FakeImageRepository) AddImage(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
	if r.OnAddImage != nil {
		return r.OnAddImage(ctx, data, imageType)
	}

	return "", fmt.Errorf("not implemented")
}

func (r *FakeImageRepository) DeleteImage(ctx context.Context, path string) error {
	if r.OnDeleteImage != nil {
		return r.OnDeleteImage(ctx, path)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeImageRepository) GetAllImagePaths(ctx context.Context) (map[string]struct{}, error) {
	if r.OnGetAllImagePaths != nil {
		return r.OnGetAllImagePaths(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

var _ menu.ImageRepository = (*FakeImageRepository)(nil)

func TestDefaultProductServiceAddProduct(t *testing.T) {
	tests := []struct {
		name                  string
		fakeProductRepository *FakeProductRepository
		fakeImageRepository   *FakeImageRepository
		request               *menu.AddProductRequest
		checkErr              func(t *testing.T, err error)
		checkResult           func(t *testing.T, product *menu.Product)
	}{
		{
			name: "success",
			fakeProductRepository: &FakeProductRepository{
				OnAddProduct: func(ctx context.Context, product *menu.Product) error {
					return nil
				},
			},
			fakeImageRepository: &FakeImageRepository{
				OnAddImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return "path", nil
				},
			},
			request: menu.NewAddProductRequest(
				"cake",
				"baked chocolate cake",
				uuid.New(),
				decimal.NewFromFloat(1.2),
				nil,
				"jpeg",
			),
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, product *menu.Product) {
				if product == nil {
					t.Fatal("product is nil")
				}

				if product.ImagePath != "path" {
					t.Fatalf("product.ImagePath != path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.fakeProductRepository, tt.fakeImageRepository)
			result, err := service.AddProduct(context.Background(), tt.request)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func TestDefaultProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name                  string
		fakeProductRepository *FakeProductRepository
		request               *menu.UpdateProductRequest
		checkErr              func(t *testing.T, err error)
	}{
		{
			name: "success",
			fakeProductRepository: &FakeProductRepository{
				OnUpdateProduct: func(ctx context.Context, product *menu.ProductUpdate) error {
					return nil
				},
			},
			request: menu.NewUpdateProductRequest(
				uuid.New(),
				asPointer("NewName"),
				nil,
				nil,
				nil,
			),
			checkErr: assertNoErr,
		},
		{
			name:                  "invalid update",
			fakeProductRepository: &FakeProductRepository{},
			request:               &menu.UpdateProductRequest{},
			checkErr:              assertBadRequestErr,
		},
		{
			name:                  "invalid category name",
			fakeProductRepository: &FakeProductRepository{},
			request: menu.NewUpdateProductRequest(
				uuid.New(),
				asPointer("Na"),
				nil,
				nil,
				nil,
			),
			checkErr: asserValidationsErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.fakeProductRepository, &FakeImageRepository{})
			err := service.UpdateProduct(context.Background(), tt.request)
			tt.checkErr(t, err)
		})
	}
}

func TestDefaultProductServiceReplaceProductImage(t *testing.T) {
	tests := []struct {
		name                  string
		fakeProductRepository *FakeProductRepository
		fakeImageRepository   *FakeImageRepository
		request               *menu.ReplaceProductImageRequest
		checkErr              func(t *testing.T, err error)
		checkResult           func(t *testing.T, path string)
	}{
		{
			name: "success",
			fakeProductRepository: &FakeProductRepository{
				OnGetProductImagePathById: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "path", nil
				},
				OnUpdateProductImagePath: func(ctx context.Context, id uuid.UUID, path string) error {
					if path != "newPath" {
						t.Fatalf("expect newPath, but got %s", path)
					}
					return nil
				},
			},
			request: menu.NewReplaceProductImageRequest(uuid.New(), nil, "jpeg"),
			fakeImageRepository: &FakeImageRepository{
				OnDeleteImage: func(ctx context.Context, path string) error {
					if path != "path" {
						return fmt.Errorf("expected path, got %s", path)
					}
					return nil
				},
				OnAddImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return "newPath", nil
				},
			},
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, path string) {
				t.Helper()

				if path != "newPath" {
					t.Fatalf("expect newPath, but got %s", path)
				}
			},
		},
		{
			name:                  "invalid image type",
			fakeProductRepository: &FakeProductRepository{},
			fakeImageRepository:   &FakeImageRepository{},
			request:               menu.NewReplaceProductImageRequest(uuid.New(), nil, "invalid"),
			checkErr:              asserValidationsErr,
			checkResult: func(t *testing.T, path string) {
				t.Helper()

				if path != "" {
					t.Fatalf("expect empty, but got %s", path)
				}
			},
		},
		{
			name: "error updating product",
			fakeProductRepository: &FakeProductRepository{
				OnGetProductImagePathById: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "path", nil
				},
				OnUpdateProductImagePath: func(ctx context.Context, id uuid.UUID, path string) error {
					if path != "newPath" {
						return fmt.Errorf("expect newPath, but got %s", path)
					}
					return domain.NewInternalError(
						"error updating product",
						fmt.Errorf("test error"),
					)
				},
			},
			fakeImageRepository: &FakeImageRepository{
				OnAddImage: func(ctx context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
					return "newPath", nil
				},
				OnDeleteImage: func(ctx context.Context, path string) error {
					return nil
				},
			},
			request:  menu.NewReplaceProductImageRequest(uuid.New(), nil, "jpeg"),
			checkErr: assertInternalErr,
			checkResult: func(t *testing.T, path string) {
				t.Helper()

				if path != "" {
					t.Fatalf("expect empty, but got %s", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.fakeProductRepository, tt.fakeImageRepository)
			result, err := service.ReplaceProductImage(context.Background(), tt.request)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func TestDefaultProductServiceDeleteOrphanImages(t *testing.T) {
	fakeProductRepository := &FakeProductRepository{
		OnGetProductImagePaths: func(ctx context.Context) (map[string]struct{}, error) {
			return map[string]struct{}{
				"path1": {},
				"path2": {},
				"path3": {},
				"path4": {},
			}, nil
		},
	}
	fakeImageRepository := &FakeImageRepository{
		OnDeleteImage: func(ctx context.Context, path string) error {
			if path != "path1" && path != "path2" && path != "path3" && path != "path4" {
				return fmt.Errorf("unexpected path %s", path)
			}

			return nil
		},
	}

	service := menu.NewDefaultProductService(fakeProductRepository, fakeImageRepository)
	service.DeleteOrphanImages(context.Background())
}

func TestDefaultProductServiceDeleteProduct(t *testing.T) {
	tests := []struct {
		name                  string
		fakeProductRepository *FakeProductRepository
		fakeImageRepository   *FakeImageRepository
		checkErr              func(t *testing.T, err error)
	}{
		{
			name: "success",
			fakeProductRepository: &FakeProductRepository{
				OnDeleteProduct: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "path", nil
				},
			},
			fakeImageRepository: &FakeImageRepository{
				OnDeleteImage: func(ctx context.Context, path string) error {
					if path != "path" {
						return fmt.Errorf("expected path, got %s", path)
					}
					return nil
				},
			},
			checkErr: assertNoErr,
		},
		{
			name: "error deleting image",
			fakeProductRepository: &FakeProductRepository{
				OnDeleteProduct: func(ctx context.Context, id uuid.UUID) (string, error) {
					return "path", nil
				},
			},
			fakeImageRepository: &FakeImageRepository{
				OnDeleteImage: func(ctx context.Context, path string) error {
					return domain.NewNotFoundError("image with path:" + path + " not found")
				},
			},
			checkErr: assertInternalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultProductService(tt.fakeProductRepository, tt.fakeImageRepository)
			err := service.DeleteProduct(context.Background(), uuid.New())
			tt.checkErr(t, err)
		})
	}
}
