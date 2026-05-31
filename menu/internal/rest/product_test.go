package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"menu/internal/database"
	"menu/internal/domain"
	"menu/internal/logger"
	"menu/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
)

type fakeProductService struct {
	onAdd                func(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductDraft, error)
	onGetDraft           func(ctx context.Context, id uuid.UUID) (*domain.ProductDraft, error)
	onConfirmImageUpload func(ctx context.Context, productID uuid.UUID) error
	onGetProduct         func(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	onGetImage           func(ctx context.Context, key string) (*domain.Image, error)
}

var _ domain.ProductService = (*fakeProductService)(nil)

func (f *fakeProductService) Add(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductDraft, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, request)
}

func (f *fakeProductService) GetDraft(ctx context.Context, id uuid.UUID) (*domain.ProductDraft, error) {
	if f.onGetDraft == nil {
		panic("onGetDraft not implemented")
	}
	return f.onGetDraft(ctx, id)
}

func (f *fakeProductService) ConfirmImageUpload(ctx context.Context, productID uuid.UUID) error {
	if f.onConfirmImageUpload != nil {
		return f.onConfirmImageUpload(ctx, productID)
	}
	return nil
}

func (f *fakeProductService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	if f.onGetProduct == nil {
		panic("onGetProduct not implemented")
	}
	return f.onGetProduct(ctx, id)
}

func (f *fakeProductService) GetImage(ctx context.Context, key string) (*domain.Image, error) {
	if f.onGetImage == nil {
		panic("onGetImage not implemented")
	}
	return f.onGetImage(ctx, key)
}

func TestAddProductRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *AddProductRequest
		wantField string
	}{
		{
			name: "valid",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
		},
		{
			name: "invalid name",
			request: &AddProductRequest{
				Name:             "",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
			wantField: "name",
		},
		{
			name: "invalid description",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
			wantField: "description",
		},
		{
			name: "invalid price",
			request: &AddProductRequest{
				Name:        "Valid name",
				Description: "Valid description",
				Price:       decimal.NewFromFloat(-10),
				CategoryID:  uuid.New(),
			},
			wantField: "price",
		},
		{
			name: "invalid image content type",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/pn",
			},
			wantField: "imageContentType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.request.Validate()
			if tt.wantField == "" {
				if got != nil {
					t.Errorf("Validate() = %v, want nil", got)
				}
				return
			}

			_, ok := got[tt.wantField]
			if !ok {
				t.Errorf("Missing field %v", tt.wantField)
			}
		})
	}
}

func TestProductHandlerAdd(t *testing.T) {
	tests := []struct {
		name           string
		handler        *ProductHandler
		request        string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			handler: &ProductHandler{
				service: &fakeProductService{
					onAdd: func(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductDraft, error) {
						return &domain.ProductDraft{
							Id:             uuid.New(),
							ImageUploadUrl: "http://upload.url",
						}, nil
					},
				},
			},
			request: `{
				"name": "French fries",
				"description": "Fryied potatoes with cheese",
				"price": "47.84",
				"categoryId": "a3c6116e-7070-4e68-b956-78a41929d324",
				"imageContentType": "image/png"
			}`,
			wantHttpStatus: http.StatusCreated,
		},
		{
			name:    "invalid payload",
			handler: &ProductHandler{service: &fakeProductService{}},
			request: `{
				"name": "Fr",
				"description": "Feese",
				"price": "-10",
				"categoryId": "a3c6116e-7070-4e68-b956-78a41929d324",
				"imageContentType": "image/png"
			}`,
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  ErrorCodeInvalidEntity,
		},
		{
			name:           "invalid json",
			handler:        &ProductHandler{service: &fakeProductService{}},
			request:        `{{}`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.NewWithConfig(echo.Config{
				HTTPErrorHandler: ErrorHandler,
			})
			e.POST("/products", tt.handler.AddProduct)

			req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.request))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if tt.wantHttpStatus != rec.Code {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatus, rec.Code)
			}

			if tt.wantHttpStatus == http.StatusCreated {
				var res domain.ProductDraft
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}
				return
			}

			var res ValidationErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

func TestAddProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	categoryRepository := database.NewPostgresCategoryRepository(testDb)
	productRepository := database.NewPostgresProductRepository(testDb)
	imageStorage := storage.NewS3ImageStorage(testS3Client, 15*time.Minute, testS3BucketName)
	service := domain.NewDefaultProductService(productRepository, imageStorage, logger.NewSilentLogger())
	handler := NewProductHandler("https://images", service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.POST("/products", handler.AddProduct)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(
			fmt.Sprintf(`{
				"name": "French fries",
				"description": "Fryied potatoes with cheese",
				"price": "47.84",
				"categoryId": "%s",
				"imageContentType": "image/png"
			}`,
				category.Id,
			),
		))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Want http status code %d, got %d", http.StatusCreated, rec.Code)
		}
		var res domain.ProductDraft
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}
	})
	t.Run("name conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "French fries",
			Description:      "Some test description",
			Price:            decimal.NewFromFloat(47.84),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(
			fmt.Sprintf(`{
				"name": "French fries",
				"description": "Fryied potatoes with cheese",
				"price": "47.84",
				"categoryId": "%s",
				"imageContentType": "image/png"
			}`,
				category.Id,
			),
		))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Want http status code %d, got %d", http.StatusConflict, rec.Code)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(
			fmt.Sprintf(`{
				"name": "French fries",
				"description": "Fryied potatoes with cheese",
				"price": "47.84",
				"categoryId": "%s",
				"imageContentType": "image/png"
			}`,
				uuid.New(),
			),
		))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusConflict, rec.Code)
		}
		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}
		if res.Code != domain.ErrorCodeCategoryNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNameConflict.String(), res.Code)
		}
	})
}

func TestProductHandlerGetDraft(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		handler        *ProductHandler
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			id:   uuid.NewString(),
			handler: &ProductHandler{
				service: &fakeProductService{
					onGetDraft: func(ctx context.Context, id uuid.UUID) (*domain.ProductDraft, error) {
						return &domain.ProductDraft{
							Id:             id,
							ImageUploadUrl: "http://upload/image",
						}, nil
					},
				},
			},
			wantHttpStatus: http.StatusOK,
		},
		{
			name: "invalid id",
			id:   "invalid",
			handler: &ProductHandler{
				service: &fakeProductService{},
			},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.GET("/draft/:id", tt.handler.GetDraft)

			req := httptest.NewRequest(http.MethodGet, "/draft/"+tt.id, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantHttpStatus {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatus, rec.Code)
			}
			if rec.Code == http.StatusOK {
				return
			}

			var res ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

func TestGetDraft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	categoryRepository := database.NewPostgresCategoryRepository(testDb)
	productRepository := database.NewPostgresProductRepository(testDb)

	imageStorage := storage.NewS3ImageStorage(testS3Client, 15*time.Minute, testS3BucketName)

	service := domain.NewDefaultProductService(productRepository, imageStorage, logger.NewSilentLogger())
	handler := NewProductHandler("https://images", service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.GET("/draft/:id", handler.GetDraft)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test product",
			Description:      "Test product description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypeJPEG,
			Status:           domain.ProductStatusAwaitingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/draft/"+product.Id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
		}

		var res ProductDraftResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}

		if product.Id != res.Id {
			t.Fatalf("Want product id %s, got %s", product.Id, res.Id)
		}
	})

	t.Run("product already finished", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test product",
			Description:      "Test product description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypeJPEG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/draft/"+product.Id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}

		if domain.ErrorCodeProductAlreadyFinished.String() != res.Code {
			t.Fatalf("Want error code %s, got %s", res.Code, domain.ErrorCodeProductAlreadyFinished.String())
		}
	})
}

func TestProductHandlerConfirmImageUpload(t *testing.T) {
	tests := []struct {
		name           string
		handler        *ProductHandler
		id             string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			handler: &ProductHandler{
				service: &fakeProductService{
					onConfirmImageUpload: func(ctx context.Context, productID uuid.UUID) error {
						return nil
					},
				},
			},
			id:             uuid.NewString(),
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name: "invalid id",
			handler: &ProductHandler{
				service: &fakeProductService{},
			},
			id:             "invalid",
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.POST("/products/confirm-image-upload/:id", tt.handler.ConfirmImageUpload)

			req := httptest.NewRequest(http.MethodPost, "/products/confirm-image-upload/"+tt.id, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantHttpStatus {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatus, rec.Code)
			}
			if rec.Code == http.StatusNoContent {
				return
			}

			var res ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

// generateTestImage generates a 10 * 10 png image.
func generateTestImage(t *testing.T) io.Reader {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer

	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Error encoding image: %v", err)
	}
	return &buf
}

func TestConfirmImageUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	categoryRepository := database.NewPostgresCategoryRepository(testDb)
	productRepository := database.NewPostgresProductRepository(testDb)

	imageStorage := storage.NewS3ImageStorage(testS3Client, 15*time.Minute, testS3BucketName)
	manager := transfermanager.New(testS3Client)

	service := domain.NewDefaultProductService(productRepository, imageStorage, logger.NewSilentLogger())
	handler := NewProductHandler("https://images", service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.POST("/products/confirm-image-upload/:id", handler.ConfirmImageUpload)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "French fries",
			Description:      "Some test description",
			Price:            decimal.NewFromFloat(47.84),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusAwaitingImage,
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
			Bucket: aws.String(testS3BucketName),
			Key:    aws.String(product.ImageKey),
			Body:   generateTestImage(t),
		}); err != nil {
			t.Fatalf("Error uploading image: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products/confirm-image-upload/"+product.Id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("Want http status code %d, got %d", http.StatusNoContent, rec.Code)
		}

		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Status != domain.ProductStatusReady {
			t.Fatalf("Want product status %s, got %s", domain.ProductStatusReady, fetchedProduct.Status)
		}
	})

	t.Run("invalid image type", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "French fries",
			Description:      "Some test description",
			Price:            decimal.NewFromFloat(47.84),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusAwaitingImage,
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
			Bucket: aws.String(testS3BucketName),
			Key:    aws.String(product.ImageKey),
			Body:   strings.NewReader("fakeImage"),
		}); err != nil {
			t.Fatalf("Error uploading image: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products/confirm-image-upload/"+product.Id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Want http status code %d, got %d", http.StatusUnprocessableEntity, rec.Code)
		}

		// validate product is deleted
		var exists bool
		row := testDb.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", product.Id)
		if err := row.Scan(&exists); err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}
		if exists {
			t.Fatalf("Product should be deleted if the image content type is invalid")
		}

		// validate the image is deleted
		_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
			Bucket: aws.String(testS3BucketName),
			Key:    aws.String(product.ImageKey),
		})
		if _, ok := errors.AsType[*types.NotFound](err); !ok {
			t.Fatalf("Want *types.NotFound error, got: %T", err)
		}
	})
}

func TestProductHandlerGetProduct(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		handler        *ProductHandler
		wantHttpStatus int
		wantImageUrl   string
		wantErrorCode  string
	}{
		{
			name: "success",
			id:   uuid.NewString(),
			handler: &ProductHandler{
				baseImageUrl: "http://images/download",
				service: &fakeProductService{
					onGetProduct: func(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
						return &domain.Product{
							Id:               uuid.New(),
							Name:             "French fries",
							Description:      "Some test description",
							Price:            decimal.NewFromFloat(47.84),
							CategoryId:       uuid.New(),
							ImageKey:         "imageKey",
							ImageContentType: "image/png",
							Status:           domain.ProductStatusAwaitingImage,
						}, nil
					},
				},
			},
			wantHttpStatus: http.StatusOK,
			wantImageUrl:   "http://images/download/imageKey",
		},
		{
			name: "invalid id",
			id:   "invalid",
			handler: &ProductHandler{
				baseImageUrl: "http://images/download",
				service:      &fakeProductService{},
			},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.GET("/products/:id", tt.handler.GetProduct)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantHttpStatus {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var res ProductResponse
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("Error decoding response: %v", err)
				}

				if res.ImageUrl != tt.wantImageUrl {
					t.Fatalf("Want image url %s, got %s", tt.wantImageUrl, res.ImageUrl)
				}
				return
			}

			var res ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}
