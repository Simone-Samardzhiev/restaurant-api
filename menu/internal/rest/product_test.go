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
	"slices"
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
	onAdd                       func(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductUploadInfo, error)
	onGetUploadInfo             func(ctx context.Context, id uuid.UUID) (*domain.ProductUploadInfo, error)
	onConfirmImageUpload        func(ctx context.Context, productID uuid.UUID) error
	onGetProduct                func(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	onGetAllWithImage           func(ctx context.Context) ([]domain.Product, error)
	onGetImage                  func(ctx context.Context, key string) (*domain.Image, error)
	onUpdateProduct             func(ctx context.Context, request *domain.UpdateProductRequest) error
	onMarkProductForImageUpdate func(ctx context.Context, id uuid.UUID, contentType domain.ImageContentType) error
	onDelete                    func(ctx context.Context, id uuid.UUID) error
}

var _ domain.ProductService = (*fakeProductService)(nil)

func (f *fakeProductService) Add(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductUploadInfo, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, request)
}

func (f *fakeProductService) GetUploadInfo(ctx context.Context, id uuid.UUID) (*domain.ProductUploadInfo, error) {
	if f.onGetUploadInfo == nil {
		panic("onGetUploadInfo not implemented")
	}
	return f.onGetUploadInfo(ctx, id)
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

func (f *fakeProductService) GetAllWithImage(ctx context.Context) ([]domain.Product, error) {
	if f.onGetAllWithImage == nil {
		panic("onGetAllReadyProducts not implemented")
	}
	return f.onGetAllWithImage(ctx)
}

func (f *fakeProductService) GetImage(ctx context.Context, key string) (*domain.Image, error) {
	if f.onGetImage == nil {
		panic("onGetImage not implemented")
	}
	return f.onGetImage(ctx, key)
}

func (f *fakeProductService) UpdateProduct(ctx context.Context, request *domain.UpdateProductRequest) error {
	if f.onUpdateProduct == nil {
		panic("onUpdateProduct not implemented")
	}
	return f.onUpdateProduct(ctx, request)
}

func (f *fakeProductService) MarkProductForImageUpdate(ctx context.Context, id uuid.UUID, contentType domain.ImageContentType) error {
	if f.onMarkProductForImageUpdate == nil {
		panic("onMarkProductForImageUpdate not implemented")
	}
	return f.onMarkProductForImageUpdate(ctx, id, contentType)
}

func (f *fakeProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDeleteProduct not implemented")
	}
	return f.onDelete(ctx, id)
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
					onAdd: func(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductUploadInfo, error) {
						return &domain.ProductUploadInfo{
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
				var res domain.ProductUploadInfo
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
		var res domain.ProductUploadInfo
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

func TestProductHandlerGetUploadInfo(t *testing.T) {
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
					onGetUploadInfo: func(ctx context.Context, id uuid.UUID) (*domain.ProductUploadInfo, error) {
						return &domain.ProductUploadInfo{
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
			e.GET("/draft/:id", tt.handler.GetUploadInfo)

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

func TestGetUploadInfo(t *testing.T) {
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
	e.GET("/draft/:id", handler.GetUploadInfo)

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
			Status:           domain.ProductStatusMissingImage,
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

		if domain.ErrorCodeProductAlreadyHasImage.String() != res.Code {
			t.Fatalf("Want error code %s, got %s", res.Code, domain.ErrorCodeProductAlreadyHasImage.String())
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
func generateTestImage(t *testing.T) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer

	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Error encoding image: %v", err)
	}
	return buf.Bytes()
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
			Status:           domain.ProductStatusMissingImage,
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
			Bucket: aws.String(testS3BucketName),
			Key:    aws.String(product.ImageKey),
			Body:   bytes.NewReader(generateTestImage(t)),
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
			Status:           domain.ProductStatusMissingImage,
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				t.Fatalf("Timeout for checking if product and image are deleted")
			default:
				t.Log("Checking if product exists")

				var exists bool
				row := testDb.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", product.Id)
				if err := row.Scan(&exists); err != nil {
					t.Fatalf("Error fetching product: %v", err)
				}

				_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
					Bucket: aws.String(testS3BucketName),
					Key:    aws.String(product.ImageKey),
				})
				_, ok := errors.AsType[*types.NotFound](err)
				if ok && !exists {
					return
				}
			}
			time.Sleep(1 * time.Second)
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
							Status:           domain.ProductStatusMissingImage,
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

func TestGetAllProductsWithImageReadyProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := database.NewPostgresCategoryRepository(testDb)
	productRepository := database.NewPostgresProductRepository(testDb)
	imageStorage := storage.NewS3ImageStorage(testS3Client, 10*time.Second, testS3BucketName)
	productService := domain.NewDefaultProductService(productRepository, imageStorage, logger.NewSilentLogger())
	productHandler := NewProductHandler("https://images/download", productService)

	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.GET("/products", productHandler.GetAllProductsWithImage)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories, products`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	category := &domain.Category{
		Id:        uuid.New(),
		Name:      "Test name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := categoryRepository.Save(context.Background(), category); err != nil {
		t.Fatalf("Error saving category: %v", err)
	}

	products := []domain.Product{
		{
			Id:               uuid.New(),
			Name:             "Test name 1",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 2",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey2",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 3",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey3",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 4",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey4",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
	}

	for _, product := range products {
		if err := productRepository.Save(context.Background(), &product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}
	}

	products = slices.DeleteFunc(products, func(product domain.Product) bool {
		return product.Status == domain.ProductStatusMissingImage
	})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}
	var res []ProductResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if len(res) != len(products) {
		t.Fatalf("Want %d products, got %d", len(products), len(res))
	}

	slices.SortFunc(products, func(a, b domain.Product) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.SortFunc(res, func(a, b ProductResponse) int {
		return strings.Compare(a.Name, b.Name)
	})

	for i := 0; i < len(products); i++ {
		if products[i].Name != res[i].Name {
			t.Errorf("Want product %s, got %s", products[i].Name, res[i].Name)
		}
	}
}

func TestGetImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	productRepository := database.NewPostgresProductRepository(testDb)
	imageStorage := storage.NewS3ImageStorage(testS3Client, 15*time.Minute, testS3BucketName)
	manager := transfermanager.New(testS3Client)

	service := domain.NewDefaultProductService(productRepository, imageStorage, logger.NewSilentLogger())
	handler := NewProductHandler("https://images", service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.GET("/product/image/:key", handler.GetImage)

	const key = "imageKey"
	fakeImage := generateTestImage(t)

	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket:      aws.String(testS3BucketName),
		Key:         aws.String(key),
		ContentType: aws.String("image/png"),
		Body:        bytes.NewReader(fakeImage),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/product/image/"+key, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}

	data, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	if !bytes.Equal(fakeImage, data) {
		t.Fatalf("Want image data %s, got %s", fakeImage, data)
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("Want image/png got %s", rec.Header().Get("Content-Type"))
	}
}

func TestUpdateProductRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *UpdateProductRequest
		wantField string
	}{
		{
			name: "Valid",
			request: &UpdateProductRequest{
				Name: new("New name"),
			},
		},
		{
			name: "Invalid name",
			request: &UpdateProductRequest{
				Name: new(""),
			},
			wantField: "name",
		},
		{
			name: "Invalid description",
			request: &UpdateProductRequest{
				Description: new(""),
			},
			wantField: "description",
		},
		{
			name: "Invalid price",
			request: &UpdateProductRequest{
				Price: new(decimal.NewFromInt(-1)),
			},
			wantField: "price",
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

func TestProductHandlerUpdateProduct(t *testing.T) {
	tests := []struct {
		name           string
		handler        *ProductHandler
		id             string
		request        string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "sucess",
			handler: &ProductHandler{
				service: &fakeProductService{onUpdateProduct: func(ctx context.Context, request *domain.UpdateProductRequest) error {
					return nil
				}},
			},
			id:             uuid.NewString(),
			request:        `{"name":"New name"}`,
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "empty",
			handler:        &ProductHandler{service: &fakeProductService{}},
			id:             uuid.NewString(),
			request:        `{}`,
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			handler:        &ProductHandler{service: &fakeProductService{}},
			id:             "invalid",
			request:        `{"name":"New name"}`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
		{
			name:           "invalid json",
			handler:        &ProductHandler{service: &fakeProductService{}},
			id:             uuid.NewString(),
			request:        `{"name":"New name}`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.PATCH("/products/:id", tt.handler.UpdateProduct)

			req := httptest.NewRequest(http.MethodPatch, "/products/"+tt.id, strings.NewReader(tt.request))
			req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
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
				t.Fatalf("Error decoding response: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Errorf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
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
	e.PATCH("/products/:id", handler.UpdateProduct)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE `); err != nil {
			t.Fatalf("Error truncating table categories: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error creating category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test name",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/products/"+product.Id.String(), strings.NewReader(`{"name":"New name"}`))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("Want http status code %d, got %d", http.StatusNoContent, rec.Code)
		}

		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Name != "New name" {
			t.Fatalf("Want New name got %s", fetchedProduct.Name)
		}
	})

	t.Run("conflicting name", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product1 := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test1",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product1); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		product2 := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test2",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey2",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product2); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/products/"+product1.Id.String(), strings.NewReader(`{"name":"Test2"}`))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Want http status code %d, got %d", http.StatusConflict, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}

		if res.Code != domain.ErrorCodeProductNameConflict.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNameConflict, res.Code)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error creating category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test name",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/products/"+product.Id.String(), strings.NewReader(fmt.Sprintf(`{"categoryId": "%s"}`, uuid.NewString())))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}
		if res.Code != domain.ErrorCodeCategoryNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNotFound, res.Code)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/products/"+uuid.NewString(), strings.NewReader(`{"name":"New name"}`))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}
		if res.Code != domain.ErrorCodeProductNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNotFound, res.Code)
		}
	})
}

func TestMarkProductForImageUpdateRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *MarkProductForImageUpdateRequest
		wantField string
	}{
		{
			name: "valid",
			request: &MarkProductForImageUpdateRequest{
				ImageContentType: "image/png",
			},
		},
		{
			name: "invalid",
			request: &MarkProductForImageUpdateRequest{
				ImageContentType: "application/json",
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

func TestProductHandlerMarkProductForImageUpdate(t *testing.T) {
	tests := []struct {
		name           string
		handler        *ProductHandler
		id             string
		request        string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			handler: &ProductHandler{
				service: &fakeProductService{
					onMarkProductForImageUpdate: func(ctx context.Context, id uuid.UUID, contentType domain.ImageContentType) error {
						return nil
					}},
			},
			id:             uuid.NewString(),
			request:        `{"imageContentType":"image/png"}`,
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			handler:        &ProductHandler{service: &fakeProductService{}},
			id:             "invalid",
			request:        `{"imageContentType":"image/png"}`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
		{
			name:           "invalid request",
			handler:        &ProductHandler{service: &fakeProductService{}},
			id:             uuid.NewString(),
			request:        `{"imageContentType":"image/png`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.PATCH("/products/:id", test.handler.MarkProductForImageUpdate)

			req := httptest.NewRequest(http.MethodPatch, "/products/"+test.id, strings.NewReader(test.request))
			req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != test.wantHttpStatus {
				t.Fatalf("Want http status code %d, got %d", test.wantHttpStatus, rec.Code)
			}

			if rec.Code == http.StatusNoContent {
				return
			}

			var res ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response: %v", err)
			}
			if res.Code != test.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", test.wantErrorCode, res.Code)
			}
		})
	}
}

func TestMarkProductForImageUpdate(t *testing.T) {
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
	e.PATCH("/products/:id", handler.MarkProductForImageUpdate)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test name",
			Description:      "Some test description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/products/"+product.Id.String(), strings.NewReader(`{"imageContentType":"image/jpeg"}`))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("Want http status code %d, got %d", http.StatusNoContent, rec.Code)
		}

		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}
		if fetchedProduct.PendingImageKey == nil {
			t.Errorf("Product does not have a pending image key")
		}
		if *fetchedProduct.PendingImageContentType != domain.ImageContentTypeJPEG {
			t.Errorf("Want pending image content type %s, got %s", domain.ImageContentTypeJPEG, *fetchedProduct.PendingImageContentType)
		}
	})
	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/products/"+uuid.NewString(), strings.NewReader(`{"imageContentType":"image/jpeg"}`))
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}
		if res.Code != domain.ErrorCodeProductNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNotFound, res.Code)
		}
	})
}

func TestProductHandlerDeleteProduct(t *testing.T) {
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
					onDelete: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
				},
			},
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			id:             "invalid",
			handler:        &ProductHandler{service: &fakeProductService{}},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			e.DELETE("/products/:id", tt.handler.DeleteProduct)
			e.HTTPErrorHandler = ErrorHandler

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.id, nil)
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
				t.Fatalf("Error decoding response: %v", err)
			}
			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
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
	e.DELETE("/products/:id", handler.DeleteProduct)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test name",
			Description:      "Some test description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/products/"+product.Id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("Want http status code %d, got %d", http.StatusNoContent, rec.Code)
		}

		// validate product is deleted
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				t.Fatalf("Timeout for checking if product and image are deleted")
			default:
				t.Log("Checking if product exists")

				var exists bool
				row := testDb.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", product.Id)
				if err := row.Scan(&exists); err != nil {
					t.Fatalf("Error fetching product: %v", err)
				}

				_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
					Bucket: aws.String(testS3BucketName),
					Key:    aws.String(product.ImageKey),
				})
				_, ok := errors.AsType[*types.NotFound](err)
				if ok && !exists {
					return
				}
			}
			time.Sleep(1 * time.Second)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.NewString(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response: %v", err)
		}
		if res.Code != domain.ErrorCodeProductNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNotFound, res.Code)
		}
	})
}
