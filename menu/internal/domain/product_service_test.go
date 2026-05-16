package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeProductRepository struct {
	onSave func(ctx context.Context, product *Product) error
}

func (f *fakeProductRepository) Save(ctx context.Context, product *Product) error {
	if f.onSave == nil {
		panic("onSave not implemented")
	}
	return f.onSave(ctx, product)
}

var _ ProductRepository = (*fakeProductRepository)(nil)

type fakeImageStorage struct {
	onCreateUploadUrl func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)
}

func (f *fakeImageStorage) CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
	if f.onCreateUploadUrl == nil {
		panic("onCreateUploadUrl not implemented")
	}
	return f.onCreateUploadUrl(ctx, imageKey, contentType)
}

var _ ImageStorage = (*fakeImageStorage)(nil)

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
