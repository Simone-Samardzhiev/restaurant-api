package domain

import (
	"context"
	"errors"
	"menu/internal/logger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeProductRepository struct {
	onSave                  func(ctx context.Context, product *Product) error
	onGet                   func(ctx context.Context, id uuid.UUID) (*Product, error)
	onDelete                func(ctx context.Context, id uuid.UUID) error
	onUpdateStatus          func(ctx context.Context, id uuid.UUID, status ProductStatus) error
	onDeleteExpiredByStatus func(ctx context.Context, olderThan time.Duration) error
}

var _ ProductRepository = (*fakeProductRepository)(nil)

func (f *fakeProductRepository) Save(ctx context.Context, product *Product) error {
	if f.onSave == nil {
		panic("onSave not implemented")
	}
	return f.onSave(ctx, product)
}

func (f *fakeProductRepository) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
	if f.onGet == nil {
		panic("onGet not implemented")
	}
	return f.onGet(ctx, id)
}

func (f *fakeProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	return f.onDelete(ctx, id)
}

func (f *fakeProductRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status ProductStatus) error {
	if f.onUpdateStatus == nil {
		panic("onUpdateStatus not implemented")
	}
	return f.onUpdateStatus(ctx, id, status)
}

func (f *fakeProductRepository) DeleteExpiredByStatus(ctx context.Context, olderThan time.Duration) error {
	if f.onDeleteExpiredByStatus == nil {
		panic("onDeleteExpiredByStatus not implemented")
	}
	return f.onDeleteExpiredByStatus(ctx, olderThan)
}

type fakeImageStorage struct {
	onCreateUploadUrl func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)
	onDelete          func(ctx context.Context, imageKey string) error
	onValidate        func(ctx context.Context, imageKey string, contentType ImageContentType) error
}

var _ ImageStorage = (*fakeImageStorage)(nil)

func (f *fakeImageStorage) CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
	if f.onCreateUploadUrl == nil {
		panic("onCreateUploadUrl not implemented")
	}
	return f.onCreateUploadUrl(ctx, imageKey, contentType)
}

func (f *fakeImageStorage) Delete(ctx context.Context, imageKey string) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	return f.onDelete(ctx, imageKey)
}

func (f *fakeImageStorage) Validate(ctx context.Context, imageKey string, contentType ImageContentType) error {
	if f.onValidate == nil {
		panic("onValidate not implemented")
	}
	return f.onValidate(ctx, imageKey, contentType)
}

func TestDefaultProductServiceAdd(t *testing.T) {
	tests := []struct {
		name      string
		request   *AddProductRequest
		service   ProductService
		wantDraft *ProductDraft
		wantError *Error
	}{
		{
			name: "success",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onSave: func(ctx context.Context, product *Product) error {
						return nil
					},
				},
				storage: &fakeImageStorage{
					onCreateUploadUrl: func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
						return "upload.url", nil
					},
				},
				logger: logger.NewSilentLogger(),
			},
			request: &AddProductRequest{
				Name:             "Product name",
				Description:      "Product description",
				Price:            decimal.NewFromFloat(10),
				CategoryId:       uuid.New(),
				ImageContentType: ImageContentTypePNG,
			},
			wantDraft: &ProductDraft{
				ImageUploadUrl: "upload.url",
			},
		},
		{
			name: "invalid content type format",
			service: &DefaultProductService{
				repository: &fakeProductRepository{},
				storage:    &fakeImageStorage{},
				logger:     logger.NewSilentLogger(),
			},
			request: &AddProductRequest{
				ImageContentType: "invalid",
			},
			wantError: NewError("", ErrorCodeInternal, nil),
		},
		{
			name: "error saving product",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onSave: func(ctx context.Context, product *Product) error {
						return NewError("error saving product", ErrorCodeInternal, nil)
					},
				},
				storage: &fakeImageStorage{},
			},
			request: &AddProductRequest{
				Name:             "Product name",
				Description:      "Product description",
				Price:            decimal.NewFromFloat(10),
				CategoryId:       uuid.New(),
				ImageContentType: ImageContentTypePNG,
			},
			wantError: NewError("error saving product", ErrorCodeInternal, nil),
		},
		{
			name: "error creating upload url",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onSave: func(ctx context.Context, product *Product) error {
						return nil
					},
				},
				storage: &fakeImageStorage{
					onCreateUploadUrl: func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
						return "", NewError("error creating upload url", ErrorCodeInternal, nil)
					},
				},
				logger: logger.NewSilentLogger(),
			},
			request: &AddProductRequest{
				Name:             "Product name",
				Description:      "Product description",
				Price:            decimal.NewFromFloat(10),
				CategoryId:       uuid.New(),
				ImageContentType: ImageContentTypePNG,
			},
			wantError: NewError("error creating upload url", ErrorCodeInternal, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			draft, err := tt.service.Add(context.Background(), tt.request)
			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if domainErr.Code != tt.wantError.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Want no error, got: %v", err)
			}
			if draft.ImageUploadUrl != tt.wantDraft.ImageUploadUrl {
				t.Fatalf("Want upload url: %s, got: %s", tt.wantDraft.ImageUploadUrl, draft.ImageUploadUrl)
			}
		})
	}
}

func TestDefaultProductServiceConfirmImageUpload(t *testing.T) {
	tests := []struct {
		name      string
		service   *DefaultProductService
		wantError *Error
	}{
		{
			name: "success",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
						return &Product{}, nil
					},
					onUpdateStatus: func(ctx context.Context, id uuid.UUID, status ProductStatus) error {
						return nil
					},
				},
				storage: &fakeImageStorage{
					onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
						return nil
					},
				},
			},
		},
		{
			name: "invalid image",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
						return &Product{}, nil
					},
					onDelete: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
					onUpdateStatus: func(ctx context.Context, id uuid.UUID, status ProductStatus) error {
						return nil
					},
				},
				storage: &fakeImageStorage{
					onDelete: func(ctx context.Context, imageKey string) error {
						return nil
					},
					onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
						return NewError("invalid image", ErrorCodeInvalidImage, nil)
					},
				},
			},
			wantError: NewError("invalid image", ErrorCodeInvalidImage, nil),
		},
		{
			name: "error updating status",
			service: &DefaultProductService{
				repository: &fakeProductRepository{
					onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
						return &Product{}, nil
					},
					onDelete: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
					onUpdateStatus: func(ctx context.Context, id uuid.UUID, status ProductStatus) error {
						return NewError("error updating product status", ErrorCodeInternal, nil)
					},
				},
				storage: &fakeImageStorage{
					onDelete: func(ctx context.Context, imageKey string) error {
						return nil
					},
					onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
						return nil
					},
				},
			},
			wantError: NewError("error updating product status", ErrorCodeInternal, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.service.ConfirmImageUpload(context.Background(), uuid.New())
			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if domainErr.Code != tt.wantError.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Want no error, got: %v", err)
			}
		})
	}
}
