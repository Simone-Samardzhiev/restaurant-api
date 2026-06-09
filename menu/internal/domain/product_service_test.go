package domain

import (
	"context"
	"errors"
	"menu/internal/logger"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeProductRepository struct {
	onSave      func(ctx context.Context, product *Product) error
	onSaveCount atomic.Int32

	onGet      func(ctx context.Context, id uuid.UUID) (*Product, error)
	onGetCount atomic.Int32

	onGetAllWithImage      func(ctx context.Context) ([]Product, error)
	onGetAllWithImageCount atomic.Int32

	onUpdate      func(ctx context.Context, request *UpdateProductRequest) error
	onUpdateCount atomic.Int32

	onUpdateStatus      func(ctx context.Context, id uuid.UUID, status ProductStatus) error
	onUpdateStatusCount atomic.Int32

	onMarkForImageUpdate      func(ctx context.Context, id uuid.UUID, imageKey string, contentType ImageContentType) error
	onMarkForImageUpdateCount atomic.Int32

	onConfirmImageUpdate      func(ctx context.Context, id uuid.UUID) error
	onConfirmImageUpdateCount atomic.Int32

	onDelete      func(ctx context.Context, id uuid.UUID) error
	onDeleteCount atomic.Int32
	deletedSignal chan struct{}

	onDeleteExpiredByStatus      func(ctx context.Context, olderThan time.Duration) ([]string, error)
	onDeleteExpiredByStatusCount atomic.Int32
}

var _ ProductRepository = (*fakeProductRepository)(nil)

func (f *fakeProductRepository) Save(ctx context.Context, product *Product) error {
	if f.onSave == nil {
		panic("onSave not implemented")
	}
	f.onSaveCount.Add(1)
	return f.onSave(ctx, product)
}

func (f *fakeProductRepository) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
	if f.onGet == nil {
		panic("onGet not implemented")
	}
	f.onGetCount.Add(1)
	return f.onGet(ctx, id)
}

func (f *fakeProductRepository) GetAllWithImage(ctx context.Context) ([]Product, error) {
	if f.onGetAllWithImage == nil {
		panic("onGetAllReady not implemented")
	}
	f.onGetAllWithImageCount.Add(1)
	return f.onGetAllWithImage(ctx)
}

func (f *fakeProductRepository) Update(ctx context.Context, request *UpdateProductRequest) error {
	if f.onUpdate == nil {
		panic("onUpdate not implemented")
	}
	f.onUpdateCount.Add(1)
	return f.onUpdate(ctx, request)
}

func (f *fakeProductRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status ProductStatus) error {
	if f.onUpdateStatus == nil {
		panic("onUpdateStatus not implemented")
	}
	f.onUpdateStatusCount.Add(1)
	return f.onUpdateStatus(ctx, id, status)
}

func (f *fakeProductRepository) MarkForImageUpdate(ctx context.Context, id uuid.UUID, imageKey string, contentType ImageContentType) error {
	if f.onMarkForImageUpdate == nil {
		panic("onMarkForImageUpdate not implemented")
	}
	f.onMarkForImageUpdateCount.Add(1)
	return f.onMarkForImageUpdate(ctx, id, imageKey, contentType)
}

func (f *fakeProductRepository) ConfirmImageUpdate(ctx context.Context, id uuid.UUID) error {
	if f.onConfirmImageUpdate == nil {
		panic("onConfirmImageUpdate not implemented")
	}
	f.onConfirmImageUpdateCount.Add(1)
	return f.onConfirmImageUpdate(ctx, id)
}

func (f *fakeProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	f.onDeleteCount.Add(1)
	if f.deletedSignal != nil {
		f.deletedSignal <- struct{}{}
	}
	return f.onDelete(ctx, id)
}

func (f *fakeProductRepository) DeleteExpiredByStatus(ctx context.Context, olderThan time.Duration) ([]string, error) {
	if f.onDeleteExpiredByStatus == nil {
		panic("onDeleteExpiredByStatus not implemented")
	}
	f.onDeleteExpiredByStatusCount.Add(1)
	return f.onDeleteExpiredByStatus(ctx, olderThan)
}

type fakeImageStorage struct {
	onCreateUploadUrl      func(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)
	onCreateUploadUrlCount atomic.Int32

	onDelete      func(ctx context.Context, imageKey string) error
	onDeleteCount atomic.Int32
	deletedSignal chan struct{}

	onDeleteBatch      func(ctx context.Context, imageKeys []string) error
	onDeleteBatchCount atomic.Int32

	onValidate      func(ctx context.Context, imageKey string, contentType ImageContentType) error
	onValidateCount atomic.Int32

	onGet      func(ctx context.Context, imageKey string) (*Image, error)
	onGetCount atomic.Int32
}

var _ ImageStorage = (*fakeImageStorage)(nil)

func (f *fakeImageStorage) CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error) {
	if f.onCreateUploadUrl == nil {
		panic("onCreateUploadUrl not implemented")
	}
	f.onCreateUploadUrlCount.Add(1)
	return f.onCreateUploadUrl(ctx, imageKey, contentType)
}

func (f *fakeImageStorage) Delete(ctx context.Context, imageKey string) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	f.onDeleteCount.Add(1)
	if f.deletedSignal != nil {
		f.deletedSignal <- struct{}{}
	}
	return f.onDelete(ctx, imageKey)
}

func (f *fakeImageStorage) DeleteMultiple(ctx context.Context, imageKeys []string) error {
	if f.onDeleteBatch == nil {
		panic("onDeleteBatch not implemented")
	}
	f.onDeleteBatchCount.Add(1)
	return f.onDeleteBatch(ctx, imageKeys)
}

func (f *fakeImageStorage) Validate(ctx context.Context, imageKey string, contentType ImageContentType) error {
	if f.onValidate == nil {
		panic("onValidate not implemented")
	}
	f.onValidateCount.Add(1)
	return f.onValidate(ctx, imageKey, contentType)
}

func (f *fakeImageStorage) Get(ctx context.Context, imageKey string) (*Image, error) {
	if f.onGet == nil {
		panic("onGet not implemented")
	}
	f.onGetCount.Add(1)
	return f.onGet(ctx, imageKey)
}

func TestDefaultProductServiceAdd(t *testing.T) {
	tests := []struct {
		name string

		repository      *fakeProductRepository
		wantOnSaveCount int32

		storage                    *fakeImageStorage
		wantOnCreateUploadUrlCount int32

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

			if tt.repository.onSaveCount.Load() != tt.wantOnSaveCount {
				t.Errorf("Want onSave count: %d, got: %d", tt.wantOnSaveCount, tt.repository.onSaveCount.Load())
			}
			if tt.storage.onCreateUploadUrlCount.Load() != tt.wantOnCreateUploadUrlCount {
				t.Errorf("Want onCreateUploadUrl count: %d, got: %d", tt.wantOnCreateUploadUrlCount, tt.storage.onCreateUploadUrlCount.Load())
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
		wantOnGetCount int32

		storage                    *fakeImageStorage
		wantOnCreateUploadUrlCount int32

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
						Status:           ProductStatusMissingImage,
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
					return nil, NewError("product is already completed", ErrorCodeProductAlreadyHasImage, nil)
				},
			},
			wantOnGetCount: 1,
			storage:        &fakeImageStorage{},

			wantError: NewError("product is already completed", ErrorCodeProductAlreadyHasImage, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewDefaultProductService(tt.repository, tt.storage, logger.NewSilentLogger())
			draft, err := service.GetUploadInfo(context.Background(), tt.id)
			if tt.wantOnGetCount != tt.repository.onGetCount.Load() {
				t.Errorf("Want onGet count: %d, got: %d", tt.wantOnGetCount, tt.repository.onGetCount.Load())
			}
			if tt.wantOnCreateUploadUrlCount != tt.storage.onCreateUploadUrlCount.Load() {
				t.Errorf("Want onCreateUploadUrl count: %d, got: %d", tt.wantOnCreateUploadUrlCount, tt.storage.onCreateUploadUrlCount.Load())
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
		wantOnGetCount           int32
		wantOnUpdateStatusCount  int32
		wantOnConfirmImageUpdate int32
		wantOnDeleteProductCount int32

		storage                *fakeImageStorage
		wantOnValidateCount    int32
		wantOnDeleteImageCount int32

		wantError *Error
	}{
		{
			name: "success new image",
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
						Status:           ProductStatusMissingImage,
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
			name: "success update",
			id:   uuid.New(),

			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:                      uuid.New(),
						Name:                    "Test name",
						Description:             "Test description",
						Price:                   decimal.NewFromInt(10),
						CategoryId:              uuid.New(),
						ImageKey:                "imageKey",
						ImageContentType:        "image/png",
						PendingImageKey:         new("newImageKey"),
						PendingImageContentType: new(ImageContentTypePNG),
						Status:                  ProductStatusAwaitingImageUpdate,
						CreatedAt:               time.Now(),
						UpdatedAt:               time.Now(),
					}, nil
				},
				onConfirmImageUpdate: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			wantOnGetCount:           1,
			wantOnConfirmImageUpdate: 1,

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
						Status:           ProductStatusMissingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
				onDelete: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
				deletedSignal: make(chan struct{}),
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
				deletedSignal: make(chan struct{}),
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
						Status:           ProductStatusMissingImage,
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
				deletedSignal: make(chan struct{}),
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
				deletedSignal: make(chan struct{}),
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
			if tt.wantOnDeleteProductCount > 0 {
				<-tt.repository.deletedSignal
			}
			if tt.wantOnDeleteImageCount > 0 {
				<-tt.storage.deletedSignal
			}

			if tt.wantOnGetCount != tt.repository.onGetCount.Load() {
				t.Errorf("Want onGet count: %d, got: %d", tt.wantOnGetCount, tt.repository.onGetCount.Load())
			}
			if tt.wantOnUpdateStatusCount != tt.repository.onUpdateStatusCount.Load() {
				t.Errorf("Want onUpdateStatus count: %d, got: %d", tt.wantOnUpdateStatusCount, tt.repository.onUpdateStatusCount.Load())
			}
			if tt.wantOnConfirmImageUpdate != tt.repository.onConfirmImageUpdateCount.Load() {
				t.Errorf("Want wantOnConfirmImageUpdate: %d, got: %d", tt.wantOnConfirmImageUpdate, tt.repository.onConfirmImageUpdateCount.Load())
			}
			if tt.wantOnDeleteProductCount != tt.repository.onDeleteCount.Load() {
				t.Errorf("Want onDeleteProduct count: %d, got: %d", tt.wantOnDeleteProductCount, tt.repository.onDeleteCount.Load())
			}
			if tt.wantOnValidateCount != tt.storage.onValidateCount.Load() {
				t.Errorf("Want onValidate count: %d, got: %d", tt.wantOnValidateCount, tt.storage.onValidateCount.Load())
			}
			if tt.wantOnDeleteImageCount != tt.storage.onDeleteCount.Load() {
				t.Errorf("Want onDeleteImage count: %d, got: %d", tt.wantOnDeleteImageCount, tt.storage.onDeleteCount.Load())
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
						Status:           ProductStatusMissingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
			wantError: NewError("product not ready", ErrorCodeProductNotFound, nil),
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
				return
			}

			if err != nil {
				t.Fatalf("Want no error, got: %v", err)
			}
		})
	}
}

func TestDefaultProductServiceMarkProductForImageUpdate(t *testing.T) {
	tests := []struct {
		name        string
		id          uuid.UUID
		contentType ImageContentType

		repository             *fakeProductRepository
		wantGetCount           int32
		wantMarkForUpdateCount int32

		wantError *Error
	}{
		{
			name:        "success",
			id:          uuid.New(),
			contentType: ImageContentTypeJPEG,
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               id,
						Name:             "Test name",
						Description:      "Some test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: ImageContentTypeJPEG,
						Status:           ProductStatusReady,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
				onMarkForImageUpdate: func(ctx context.Context, id uuid.UUID, imageKey string, contentType ImageContentType) error {
					return nil
				},
			},
			wantGetCount:           1,
			wantMarkForUpdateCount: 1,
		},
		{
			name:        "invalid content type format",
			id:          uuid.New(),
			contentType: "invalid",
			repository:  &fakeProductRepository{},
			wantError:   NewError("invalid content type format", ErrorCodeInternal, nil),
		},
		{
			name:        "already marked for update",
			id:          uuid.New(),
			contentType: ImageContentTypeJPEG,
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:                      id,
						Name:                    "Test name",
						Description:             "Some test description",
						Price:                   decimal.NewFromInt(10),
						CategoryId:              uuid.New(),
						ImageKey:                "imageKey",
						ImageContentType:        ImageContentTypeJPEG,
						Status:                  ProductStatusAwaitingImageUpdate,
						PendingImageKey:         new("newImageKey"),
						PendingImageContentType: new(ImageContentTypeJPEG),
						CreatedAt:               time.Now(),
						UpdatedAt:               time.Now(),
					}, nil
				},
			},
			wantGetCount: 1,
		},
		{
			name:        "missing initial image",
			id:          uuid.New(),
			contentType: ImageContentTypeJPEG,
			repository: &fakeProductRepository{
				onGet: func(ctx context.Context, id uuid.UUID) (*Product, error) {
					return &Product{
						Id:               id,
						Name:             "Test name",
						Description:      "Some test description",
						Price:            decimal.NewFromInt(10),
						CategoryId:       uuid.New(),
						ImageKey:         "imageKey",
						ImageContentType: ImageContentTypeJPEG,
						Status:           ProductStatusMissingImage,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				},
			},
			wantGetCount: 1,
			wantError:    NewError("missing initial image", ErrorCodeProductMissingInitialImage, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := NewDefaultProductService(tt.repository, &fakeImageStorage{}, logger.NewSilentLogger())
			err := service.MarkProductForImageUpdate(context.Background(), tt.id, tt.contentType)

			if tt.wantGetCount != tt.repository.onGetCount.Load() {
				t.Errorf("Want get count %d, got %d", tt.wantGetCount, tt.repository.onGetCount.Load())
			}
			if tt.wantMarkForUpdateCount != tt.repository.onMarkForImageUpdateCount.Load() {
				t.Errorf("Want mark for update count %d, got %d", tt.wantGetCount, tt.repository.onMarkForImageUpdateCount.Load())
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
