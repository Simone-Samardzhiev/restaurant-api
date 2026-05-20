package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"menu/internal/database"
	"menu/internal/domain"
	"menu/internal/logger"
	"menu/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
)

type fakeProductService struct {
	onAdd                func(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductDraft, error)
	onConfirmImageUpload func(ctx context.Context, productID uuid.UUID)
}

var _ domain.ProductService = (*fakeProductService)(nil)

func (f *fakeProductService) Add(ctx context.Context, request *domain.AddProductRequest) (*domain.ProductDraft, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, request)
}

func (f *fakeProductService) ConfirmImageUpload(ctx context.Context, productID uuid.UUID) error {
	if f.onConfirmImageUpload != nil {
		f.onConfirmImageUpload(ctx, productID)
	}
	return nil
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

	repository := database.NewPostgresProductRepository(testDb)
	imageStorage := storage.NewS3ImageStorage(testS3Client, 15*time.Minute, testS3BucketName)
	service := domain.NewDefaultProductService(repository, imageStorage, logger.NewSilentLogger())
	handler := NewProductHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.POST("/products", handler.AddProduct)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	categoryId := uuid.New()
	if _, err := testDb.Exec(
		`INSERT INTO categories(id, name, created_at, updated_at)
		VALUES ($1, 'test', NOW(), NOW())`,
		categoryId,
	); err != nil {
		t.Fatalf("Error inserting category: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE products CASCADE`); err != nil {
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
				categoryId,
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
		if _, err := testDb.Exec(`TRUNCATE TABLE products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key,image_content_type, status, created_at, updated_at)
			VALUES (gen_random_uuid(), 'French fries', 'Some test description for product', 10, $1, 'testImageKey', 'image/png', 'ready', NOW(), NOW())`,
			categoryId,
		); err != nil {
			t.Fatalf("Error inserting product: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(
			fmt.Sprintf(`{
				"name": "French fries",
				"description": "Fryied potatoes with cheese",
				"price": "47.84",
				"categoryId": "%s",
				"imageContentType": "image/png"
			}`,
				categoryId,
			),
		))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Want http status code %d, got %d", http.StatusConflict, rec.Code)
		}
		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}
		if res.Code != domain.ErrorCodeProductNameConflict.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNameConflict.String(), res.Code)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE products CASCADE`); err != nil {
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
