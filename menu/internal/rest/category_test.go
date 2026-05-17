package rest

import (
	"context"
	"encoding/json"
	"menu/internal/database"
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
	onUpdate func(ctx context.Context, id uuid.UUID, name string) error
	onDelete func(ctx context.Context, id uuid.UUID) error
}

var _ domain.CategoryService = (*fakeCategoryService)(nil)

func (f *fakeCategoryService) Add(ctx context.Context, name string) (*domain.Category, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, name)
}

func (f *fakeCategoryService) GetAll(ctx context.Context) ([]domain.Category, error) {
	if f.onGetAll == nil {
		panic("onGetAll not implemented")
	}
	return f.onGetAll(ctx)
}

func (f *fakeCategoryService) Update(ctx context.Context, id uuid.UUID, name string) error {
	if f.onUpdate == nil {
		panic("onUpdate not implemented")
	}
	return f.onUpdate(ctx, id, name)
}

func (f *fakeCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	return f.onDelete(ctx, id)
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
		handler            *CategoryHandler
		request            string
		wantHttpStatusCode int
		wantName           string
		wantErrorCode      string
	}{
		{
			name: "success",
			handler: &CategoryHandler{
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
			},
			request:            `{ "name" : "test" }`,
			wantHttpStatusCode: http.StatusCreated,
			wantName:           "test",
		},
		{
			name:               "invalid payload",
			handler:            &CategoryHandler{service: &fakeCategoryService{}},
			request:            `{ "name": "n" }`,
			wantHttpStatusCode: http.StatusUnprocessableEntity,
			wantErrorCode:      ErrorCodeInvalidEntity,
		},
		{
			name:               "invalid JSON",
			handler:            &CategoryHandler{service: &fakeCategoryService{}},
			request:            `{{}`,
			wantHttpStatusCode: http.StatusBadRequest,
			wantErrorCode:      ErrorCodeInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.NewWithConfig(echo.Config{
				HTTPErrorHandler: ErrorHandler,
			})
			e.POST("/categories", tt.handler.AddCategory)

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

			if res.Code != tt.wantErrorCode {
				t.Fatalf("Want error code %s, got %s", tt.wantErrorCode, res.Code)
			}
		})
	}
}

func TestAddCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := database.NewPostgresCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)

	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.POST("/categories", handler.AddCategory)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
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
		if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at)
			VALUES (gen_random_uuid(), 'test', NOW(), NOW())`); err != nil {
			t.Fatalf("Error seeding data: %v", err)
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

		if res.Code != domain.ErrorCodeCategoryNameConflict.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNameConflict, res.Code)
		}
	})
}

func TestCategoryHandlerGetCategories(t *testing.T) {
	t.Parallel()

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
		HTTPErrorHandler: ErrorHandler,
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

	if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	repository := database.NewPostgresCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
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

func TestUpdateCategoryRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		req       *UpdateCategoryRequest
		wantField string
	}{
		{
			name: "valid",
			req: &UpdateCategoryRequest{
				Name: "test",
			},
		},
		{
			name: "short name",
			req: &UpdateCategoryRequest{
				Name: "t",
			},
			wantField: "name",
		},
		{
			name: "long name",
			req: &UpdateCategoryRequest{
				Name: "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
			},
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.req.Validate()

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

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	tests := []struct {
		name           string
		handler        *CategoryHandler
		id             string
		request        string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			handler: &CategoryHandler{
				service: &fakeCategoryService{
					onUpdate: func(ctx context.Context, id uuid.UUID, name string) error {
						return nil
					},
				},
			},
			id:             uuid.NewString(),
			request:        `{ "name":"test" }`,
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "invalid payload",
			handler:        &CategoryHandler{service: &fakeCategoryService{}},
			id:             uuid.NewString(),
			request:        `{ "name":"t" }`,
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  ErrorCodeInvalidEntity,
		},
		{
			name:           "invalid JSON",
			handler:        &CategoryHandler{service: &fakeCategoryService{}},
			id:             uuid.NewString(),
			request:        `{ "na:"test" }`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidJSON,
		},
		{
			name:           "invalid uuid",
			handler:        &CategoryHandler{service: &fakeCategoryService{}},
			id:             "invalid",
			request:        `{ "name":"test" }`,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.NewWithConfig(echo.Config{
				HTTPErrorHandler: ErrorHandler,
			})
			e.PATCH("/categories/:id", tt.handler.UpdateCategory)

			req := httptest.NewRequest(http.MethodPatch, "/categories/"+tt.id, strings.NewReader(tt.request))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantHttpStatus {
				t.Fatalf("Want http status code %d, got %d", tt.wantHttpStatus, rec.Code)
			}
			if rec.Code == http.StatusNoContent {
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

func TestUpdateCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := database.NewPostgresCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})

	e.PATCH("/categories/:id", handler.UpdateCategory)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating categories: %v", err)
		}

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at)
			VALUES ($1, 'Old name', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		const newName = "New Name"
		req := httptest.NewRequest(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{ "name":"New Name" }`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("Want http status code %d, got %d", http.StatusNoContent, rec.Code)
		}

		// Validate its updated
		row := testDb.QueryRow(`SELECT name FROM categories WHERE id = $1`, id)
		var name string
		if err := row.Scan(&name); err != nil {
			t.Fatalf("Error getting category name: %v", err)
		}

		if name != newName {
			t.Fatalf("Want name %s, got %s", newName, name)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating categories: %v", err)
		}

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'Test1', NOW(), NOW()),
			(gen_random_uuid(), 'Test2', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{ "name":"Test2" }`))
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
		if res.Code != domain.ErrorCodeCategoryNameConflict.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNameConflict.String(), res.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()
		req := httptest.NewRequest(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{"name" : "test"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}
		if res.Code != domain.ErrorCodeCategoryNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNotFound.String(), res.Code)
		}
	})
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	tests := []struct {
		name           string
		handler        *CategoryHandler
		id             string
		wantHttpStatus int
		wantErrorCode  string
	}{
		{
			name: "success",
			handler: &CategoryHandler{
				service: &fakeCategoryService{
					onDelete: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
				},
			},
			id:             uuid.New().String(),
			wantHttpStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			handler:        &CategoryHandler{service: &fakeCategoryService{}},
			id:             "invalid",
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.NewWithConfig(echo.Config{
				HTTPErrorHandler: ErrorHandler,
			})
			e.DELETE("/categories/:id", tt.handler.DeleteCategory)

			req := httptest.NewRequest(http.MethodDelete, "/categories/"+tt.id, nil)
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

func TestDeleteCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	repository := database.NewPostgresCategoryRepository(testDb)
	service := domain.NewDefaultCategoryService(repository)
	handler := NewCategoryHandler(service)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.DELETE("/categories/:id", handler.DeleteCategory)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE categories CASCADE `); err != nil {
			t.Fatalf("Error truncating categories: %v", err)
		}

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'test', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/categories/"+id.String(), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Want http status code %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.NewString()
		req := httptest.NewRequest(http.MethodDelete, "/categories/"+id, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Want http status code %d, got %d", http.StatusNotFound, rec.Code)
		}

		var res ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Error decoding response body: %v", err)
		}
		if res.Code != domain.ErrorCodeCategoryNotFound.String() {
			t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNotFound.String(), res.Code)
		}
	})
}
