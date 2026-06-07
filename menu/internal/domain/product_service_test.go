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
	onSave      func(ctx context.Context, product *Product) error
	onSaveCount int

	onGet      func(ctx context.Context, id uuid.UUID) (*Product, error)
	onGetCount int

	onGetAllReady      func(ctx context.Context) ([]Product, error)
	onGetAllReadyCount int

	onUpdate      func(ctx context.Context, request *UpdateProductRequest) error
	onUpdateCount int

	onDelete      func(ctx context.Context, id uuid.UUID) error
	onDeleteCount int

	onUpdateStatus      func(ctx context.Context, id uuid.UUID, status ProductStatus) error
	onUpdateStatusCount int

	onDeleteExpiredByStatus      func(ctx context.Context, olderThan time.Duration) ([]string, error)
	onDeleteExpiredByStatusCount int
}

var _ ProductRepository = (*fakeProductRepository)(nil)

func (f *fakeProductRepository) Save(ctx context.Context, product *Product) error {
	if f.onSave == nil {
		panic("onSave not implemented")
	}
	f.onSaveCount++
	return f.onSave(ctx, product)
}

func (f *fakeProductRepository) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
	if f.onGet == nil {
		panic("onGet not implemented")
	}
	f.onGetCount++
	return f.onGet(ctx, id)
}

func (f *fakeProductRepository) GetAllReady(ctx context.Context) ([]Product, error) {
	if f.onGetAllReady == nil {
		panic("onGetAllReady not implemented")
	}
	f.onGetAllReadyCount++
	return f.onGetAllReady(ctx)
}

func (f *fakeProductRepository) Update(ctx context.Context, request *UpdateProductRequest) error {
	if f.onUpdate == nil {
		panic("onUpdate not implemented")
	}
	f.onUpdateCount++
	return f.onUpdate(ctx, request)
}

func (f *fakeProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	f.onDeleteCount++
	return f.onDelete(ctx, id)
}

func (f *fakeProductRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status ProductStatus) error {
	if f.onUpdateStatus == nil {
		panic("onUpdateStatus not implemented")
	}
	f.onUpdateStatusCount++
	return f.onUpdateStatus(ctx, id, status)
}

func (f *fakeProductRepository) DeleteExpiredByStatus(ctx context.Context, olderThan time.Duration) ([]string, error) {
	if f.onDeleteExpiredByStatus == nil {
		panic("onDeleteExpiredByStatus not implemented")
	}
	f.onDeleteExpiredByStatusCount++
	return f.onDeleteExpiredByStatus(ctx, olderThan)
}

type fakeImageStorage struct {
	onCreateUploadUrl      func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)
	onCreateUploadUrlCount int

	onDelete      func(ctx context.Context, imageKey string) error
	onDeleteCount int

	onDeleteBatch      func(ctx context.Context, imageKeys []string) error
	onDeleteBatchCount int

	onValidate      func(ctx context.Context, imageKey string, contentType ImageContentType) error
	onValidateCount int

	onGet      func(ctx context.Context, imageKey string) (*Image, error)
	onGetCount int
}

var _ ImageStorage = (*fakeImageStorage)(nil)

func (f *fakeImageStorage) CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
	if f.onCreateUploadUrl == nil {
		panic("onCreateUploadUrl not implemented")
	}
	f.onCreateUploadUrlCount++
	return f.onCreateUploadUrl(ctx, imageKey, contentType)
}

func (f *fakeImageStorage) Delete(ctx context.Context, imageKey string) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	f.onDeleteCount++
	return f.onDelete(ctx, imageKey)
}

func (f *fakeImageStorage) DeleteMultiple(ctx context.Context, imageKeys []string) error {
	if f.onDeleteBatch == nil {
		panic("onDeleteBatch not implemented")
	}
	f.onDeleteBatchCount++
	return f.onDeleteBatch(ctx, imageKeys)
}

func (f *fakeImageStorage) Validate(ctx context.Context, imageKey string, contentType ImageContentType) error {
	if f.onValidate == nil {
		panic("onValidate not implemented")
	}
	f.onValidateCount++
	return f.onValidate(ctx, imageKey, contentType)
}

func (f *fakeImageStorage) Get(ctx context.Context, imageKey string) (*Image, error) {
	if f.onGet == nil {
		panic("onGet not implemented")
	}
	f.onGetCount++
	return f.onGet(ctx, imageKey)
}

func TestDefaultProductServiceAdd(t *testing.T) {
	tests := []struct {
		name string

		repository      *fakeProductRepository
		wantOnSaveCount int

		storage                    *fakeImageStorage
		wantOnCreateUploadUrlCount int

		request   *AddProductRequest
		wantDraft *ProductUploadInfo
		wantError *Error
	}{
		{
			name: "success",
			repository: &fakeProductRepository{
				onSave: func(ctx context.Context, product *Product) error { return nil },
			},
			wantOnSaveCount: 1,
			storage: &fakeImageStorage{
				onCreateUploadUrl: func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
					return "imageKey", nil
				},
			},
			wantOnCreateUploadUrlCount: 1,
			request: &AddProductRequest{
				Name:             "Product name",
				Description:      "Product description",
				Price:            decimal.NewFromFloat(10),
				CategoryId:       uuid.New(),
				ImageContentType: ImageContentTypePNG,
			},
			wantDraft: &ProductUploadInfo{
				ImageUploadUrl: "imageKey",
			},
		},
		{
			name:       "invalid content type format",
			repository: &fakeProductRepository{},
			storage:    &fakeImageStorage{},
			request: &AddProductRequest{
				ImageContentType: "invalid",
			},
			wantError: NewError("", ErrorCodeInternal, nil),
		},
		{
			name: "error saving product",
			repository: &fakeProductRepository{
				onSave: func(ctx context.Context, product *Product) error {
					return NewError("error saving", ErrorCodeInternal, nil)
				},
			},
			wantOnSaveCount: 1,
			storage:         &fakeImageStorage{},
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
			repository: &fakeProductRepository{
				onSave: func(ctx context.Context, product *Product) error { return nil },
			},
			wantOnSaveCount: 1,
			storage: &fakeImageStorage{
				onCreateUploadUrl: func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
					return "", NewError("error creating upload url", ErrorCodeInternal, nil)
				},
			},
			wantOnCreateUploadUrlCount: 1,
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
			service := NewDefaultProductService(tt.repository, tt.storage, logger.NewSilentLogger())
			draft, err := service.Add(context.Background(), tt.request)

			if tt.repository.onSaveCount != tt.wantOnSaveCount {
				t.Errorf("Want onSave count: %d, got: %d", tt.wantOnSaveCount, tt.repository.onSaveCount)
			}
			if tt.storage.onCreateUploadUrlCount != tt.wantOnCreateUploadUrlCount {
				t.Errorf("Want onCreateUploadUrl count: %d, got: %d", tt.wantOnCreateUploadUrlCount, tt.storage.onCreateUploadUrlCount)
			}

			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if domainErr.Code != tt.wantError.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				} else {
					t.Fatalf("Want error type *Error, got: %T", err)
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

func TestDefaultProductServiceGetUploadInfo(t *testing.T) {
	tests := []struct {
		name string
		id   uuid.UUID

		repository     *fakeProductRepository
		wantOnGetCount int

		storage                    *fakeImageStorage
		wantOnCreateUploadUrlCount int

		wantDraft *ProductUploadInfo
		wantError *Error
	}{
		{
			name: "success",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusAwaitingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
			wantOnGetCount: 1,

			storage: &fakeImageStorage{
				onCreateUploadUrl: func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
					return "https://images/upload", nil
				},
			},
			wantOnCreateUploadUrlCount: 1,

			wantDraft: &ProductUploadInfo{
				ImageUploadUrl: "https://images/upload",
			},
		},
		{
			name: "error product already finished",
			id:   uuid.New(),
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return nil, NewError("product is already completed", ErrorCodeProductAlreadyFinished, nil)
				},
			},
			wantOnGetCount: 1,
			storage:        &fakeImageStorage{},

			wantError: NewError("product is already completed", ErrorCodeProductAlreadyFinished, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewDefaultProductService(tt.repository, tt.storage, logger.NewSilentLogger())
			draft, err := service.GetUploadInfo(context.Background(), tt.id)
			if tt.wantOnGetCount != tt.repository.onGetCount {
				t.Errorf("Want onGet count: %d, got: %d", tt.wantOnGetCount, tt.repository.onGetCount)
			}
			if tt.wantOnCreateUploadUrlCount != tt.storage.onCreateUploadUrlCount {
				t.Errorf("Want onCreateUploadUrl count: %d, got: %d", tt.wantOnCreateUploadUrlCount, tt.storage.onCreateUploadUrlCount)
			}

			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if tt.wantError.Code != domainErr.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				} else {
					t.Fatalf("Want error type *Error, got: %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Want no error, got: %v", err)
			}
			if tt.wantDraft.ImageUploadUrl != draft.ImageUploadUrl {
				t.Fatalf("Want upload url: %s, got: %s", tt.wantDraft.ImageUploadUrl, draft.ImageUploadUrl)
			}
		})
	}
}

func TestDefaultProductServiceConfirmImageUpload(t *testing.T) {
	tests := []struct {
		name string
		id   uuid.UUID

		repository               *fakeProductRepository
		wantOnGetCount           int
		wantOnUpdateStatusCount  int
		wantOnDeleteProductCount int

		storage                *fakeImageStorage
		wantOnValidateCount    int
		wantOnDeleteImageCount int

		wantError *Error
	}{
		{
			name: "success",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusAwaitingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
				onUpdateStatus: func(ctx context.Context, id uuid.UUID, status ProductStatus) error {
					return nil
				},
			},
			wantOnGetCount:          1,
			wantOnUpdateStatusCount: 1,

			storage: &fakeImageStorage{
				onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
					return nil
				},
			},
			wantOnValidateCount: 1,
		},
		{
			name: "success already confirmed",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusReady,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
			wantOnGetCount: 1,
			storage:        &fakeImageStorage{},
		},
		{
			name: "invalid image",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusAwaitingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
				onDelete: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			wantOnGetCount:           1,
			wantOnDeleteProductCount: 1,

			storage: &fakeImageStorage{
				onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
					return NewError("invalid image type", ErrorCodeInvalidImage, nil)
				},
				onDelete: func(ctx context.Context, imageKey string) error {
					return nil
				},
			},
			wantOnValidateCount:    1,
			wantOnDeleteImageCount: 1,

			wantError: NewError("invalid image type", ErrorCodeInvalidImage, nil),
		},
		{
			name: "error updating product",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusAwaitingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
				onUpdateStatus: func(ctx context.Context, id uuid.UUID, status ProductStatus) error {
					return NewError("product not found", ErrorCodeProductNotFound, nil)
				},
				onDelete: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			wantOnGetCount:          1,
			wantOnUpdateStatusCount: 1,

			storage: &fakeImageStorage{
				onValidate: func(ctx context.Context, imageKey string, contentType ImageContentType) error {
					return nil
				},
				onDelete: func(ctx context.Context, imageKey string) error {
					return nil
				},
			},
			wantOnValidateCount:    1,
			wantOnDeleteImageCount: 1,

			wantError: NewError("product not found", ErrorCodeProductNotFound, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := NewDefaultProductService(tt.repository, tt.storage, logger.NewSilentLogger())
			err := service.ConfirmImageUpload(context.Background(), tt.id)

			if tt.wantOnGetCount != tt.repository.onGetCount {
				t.Errorf("Want onGet count: %d, got: %d", tt.wantOnGetCount, tt.repository.onGetCount)
			}
			if tt.wantOnUpdateStatusCount != tt.repository.onUpdateStatusCount {
				t.Errorf("Want onUpdateStatus count: %d, got: %d", tt.wantOnUpdateStatusCount, tt.repository.onUpdateStatusCount)
			}
			if tt.wantOnDeleteProductCount != tt.repository.onDeleteCount {
				t.Errorf("Want onDeleteProduct count: %d, got: %d", tt.wantOnDeleteProductCount, tt.repository.onDeleteCount)
			}
			if tt.wantOnValidateCount != tt.storage.onValidateCount {
				t.Errorf("Want onValidate count: %d, got: %d", tt.wantOnValidateCount, tt.storage.onValidateCount)
			}
			if tt.wantOnDeleteImageCount != tt.storage.onDeleteCount {
				t.Errorf("Want onDeleteImage count: %d, got: %d", tt.wantOnDeleteImageCount, tt.storage.onDeleteCount)
			}

			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if domainErr.Code != tt.wantError.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				} else {
					t.Fatalf("Want error type *Error, got: %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Want no error, got: %v", err)
			}
		})
	}
}

func TestDefaultProductServiceGetProduct(t *testing.T) {
	tests := []struct {
		name       string
		id         uuid.UUID
		repository *fakeProductRepository
		wantError  *Error
	}{
		{
			name: "success",
			id:   uuid.New(),
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusReady,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
		},
		{
			name: "product not ready",
			id:   uuid.New(),
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               uuid.New(),
						Name:             "Test name",
						Description:      "Test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: "image/png",
						Status:           ProductStatusAwaitingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := NewDefaultProductService(tt.repository, &fakeImageStorage{}, logger.NewSilentLogger())

			_, err := service.GetProduct(context.Background(), tt.id)
			if tt.wantError != nil {
				if domainErr, ok := errors.AsType[*Error](err); ok {
					if domainErr.Code != tt.wantError.Code {
						t.Errorf("Want error code: %s, got: %s", tt.wantError.Code, domainErr.Code)
					}
				} else {
					t.Fatalf("Want error type *Error, got: %T", err)
				}
			}
		})
	}
}
