package rest

import (
	"context"
	"encoding/json"
	"menu/internal/db"
	"menu/internal/domain"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type fakeCategoryService struct {
	onAdd    func(ctx context.Context, name string) (*domain.Category, error)
	onGetAll func(ctx context.Context) ([]domain.Category, error)
}

var _ domain.CategoryService = (*fakeCategoryService)(nil)

func (f fakeCategoryService) Add(ctx context.Context, name string) (*domain.Category, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, name)
}

func (f fakeCategoryService) GetAll(ctx context.Context) ([]domain.Category, error) {
	if f.onGetAll == nil {
		panic("onGetAll not implemented")
	}
	return f.onGetAll(ctx)
}

func TestAddCategoryRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *AddCategoryRequest
		wantField string
	}{
		{
			name: "valid",
			request: &AddCategoryRequest{
				Name: "test",
			},
		},
		{
			name: "short name",
			request: &AddCategoryRequest{
				Name: "",
			},
			wantField: "name",
		},
		{
			name: "long name",
			request: &AddCategoryRequest{
				Name: "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
			},
			wantField: "name",
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
				t.Errorf("Missing field: %s", tt.wantField)
			}
		})
	}

}

func TestCategoryHandlerAddCategory(t *testing.T) {
	tests := []struct {
		name               string
		service            domain.CategoryService
		request            string
		wantHttpStatusCode int
		wantName           string
		wantErrorCode      string
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onAdd: func(ctx context.Context, name string) (*domain.Category, error) {
					return &domain.Category{
						Id:        uuid.New(),
						Name:      name,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
			},
			request:            `{ "name" : "test" }`,
			wantHttpStatusCode: http.StatusCreated,
			wantName:           "test",
		},
		{
			name:               "invalid payload",
			service:            &fakeCategoryService{},
			request:            `{ "name": "n" }`,
			wantHttpStatusCode: http.StatusUnprocessableEntity,
			wantErrorCode:      invalidPayloadErrorCode,
		},
		{
			name:               "invalid JSON",
			service:            &fakeCategoryService{},
			request:            `{{}`,
			wantHttpStatusCode: http.StatusBadRequest,
			wantErrorCode:      invalidJSONErrorCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewCategoryHandler(tt.service)
			e := echo.NewWithConfig(echo.Config{
				HTTPErrorHandler: errorHandler,
			})
			e.POST("/categories", handler.AddCategory)

			req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(tt.request))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantHttpStatusCode {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatusCode, rec.Code)
			}

			if rec.Code == http.StatusCreated {
				var res CategoryResponse

				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}

				if res.Name != tt.wantName {
					t.Fatalf("Want name %s, got %s", tt.wantName, res.Name)
				}
				return
			}

			var res ValidationErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}

			if res.ErrorCode != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.ErrorCode)
			}
		})
	}
}

func TestAddCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := db.NewCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)

	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
	})
	e.POST("/categories", handler.AddCategory)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.ExecContext(context.Background(), `TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{ "name":"test"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Want http status code %d, got %d", http.StatusCreated, rec.Code)
		}

		var res CategoryResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}

		if res.Name != "test" {
			t.Fatalf("Want name %s, got %s", "test", res.Name)
		}
	})

	t.Run("conflicting name", func(t *testing.T) {
		if _, err := testDb.ExecContext(context.Background(), `TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		if _, err := testDb.ExecContext(
			context.Background(),
			`INSERT INTO categories(id, name, created_at, updated_at)
			VALUES (gen_random_uuid(), 'test', NOW(), NOW())`); err != nil {
			t.Fatalf("Error inserting data: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{ "name":"test"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Want http status code %d, got %d", http.StatusCreated, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}

		if res.ErrorCode != domain.ErrorCodeCategoryNameConflict.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNameConflict, res.ErrorCode)
		}
	})
}

func TestCategoryHandlerGetCategories(t *testing.T) {
	service := &fakeCategoryService{
		onGetAll: func(ctx context.Context) ([]domain.Category, error) {
			return []domain.Category{
				{
					Id:        uuid.New(),
					Name:      "test",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}, nil
		},
	}
	handler := NewCategoryHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
	})
	e.GET("/categories", handler.GetCategories)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}

	var res []CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Error decoding response body: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("Want 1 category, got %d", len(res))
	}
	if res[0].Name != "test" {
		t.Fatalf("Want name %s, got %s", "test", res[0].Name)
	}

}

func TestGetCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if _, err := testDb.ExecContext(context.Background(), `TRUNCATE categories CASCADE `); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	repository := db.NewCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: errorHandler,
	})

	e.POST("/categories", handler.AddCategory)
	e.GET("/categories", handler.GetCategories)

	// Add categories before fetching
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{ "name":"test1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{ "name":"test2"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
	}

	var res []CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Error decoding response body: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("Want 2 category, got %d", len(res))
	}

	wantNames := []string{"test1", "test2"}
	slices.Sort(wantNames)

	slices.SortFunc(res, func(a, b CategoryResponse) int {
		return strings.Compare(a.Name, b.Name)
	})

	for i := 0; i < len(wantNames); i++ {
		if wantNames[i] != res[i].Name {
			t.Errorf("Want name %s, got %s", wantNames[i], res[i].Name)
		}
	}
}
