package rest

import (
	"context"
	"encoding/json"
	"menu/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type fakeCategoryService struct {
	onAdd func(ctx context.Context, name string) (*domain.Category, error)
}

var _ domain.CategoryService = (*fakeCategoryService)(nil)

func (f fakeCategoryService) Add(ctx context.Context, name string) (*domain.Category, error) {
	if f.onAdd == nil {
		panic("onAdd not implemented")
	}
	return f.onAdd(ctx, name)
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
