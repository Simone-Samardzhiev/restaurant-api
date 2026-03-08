package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

type fakeCategoryService struct {
	onAddCategory    func(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error)
	onUpdateCategory func(ctx context.Context, request *menu.UpdateCategoryRequest) error
	onDeleteCategory func(ctx context.Context, id uuid.UUID) error
	onGetCategories  func(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error)
}

var _ menu.CategoryService = (*fakeCategoryService)(nil)

func (s *fakeCategoryService) AddCategory(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
	if s.onAddCategory == nil {
		panic("onAddCategory function is not implemented")
	}
	return s.onAddCategory(ctx, request)
}
func (s *fakeCategoryService) UpdateCategory(ctx context.Context, request *menu.UpdateCategoryRequest) error {
	if s.onUpdateCategory == nil {
		panic("onUpdateCategory function is not implemented")
	}
	return s.onUpdateCategory(ctx, request)
}

func (s *fakeCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if s.onDeleteCategory == nil {
		panic("onDeleteCategory function is not implemented")
	}
	return s.onDeleteCategory(ctx, id)
}

func (s *fakeCategoryService) GetCategories(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
	if s.onGetCategories == nil {
		panic("onGetCategories function is not implemented")
	}
	return s.onGetCategories(ctx, filter)
}

// checkAddCategoryResponse checks if response body is [rest.CategoryResponse] and validates the name.
func checkAddCategoryResponse(t *testing.T, body io.Reader, expectedName string) {
	t.Helper()

	var res rest.CategoryResponse
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Name != expectedName {
		t.Fatalf("want name %s, got %s", expectedName, res.Name)
	}
}

func TestCategoryHandlerAddCategory(t *testing.T) {
	tests := []struct {
		name            string
		service         *fakeCategoryService
		request         rest.AddCategoryRequest
		wantHttpStatus  int
		wantErrorCode   domain.ErrorCode
		wantDetailCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onAddCategory: func(_ context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
					return test.Must(menu.ParseCategory(uuid.New(), "New Category")), nil
				},
			},
			request:        rest.AddCategoryRequest{Name: "New Category"},
			wantHttpStatus: http.StatusCreated,
		},
		{
			name:            "name too short",
			service:         &fakeCategoryService{},
			request:         rest.AddCategoryRequest{Name: "Ne"},
			wantHttpStatus:  http.StatusUnprocessableEntity,
			wantErrorCode:   domain.ErrorCodeInvalidCategory,
			wantDetailCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:            "name too long",
			service:         &fakeCategoryService{},
			request:         rest.AddCategoryRequest{Name: "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"},
			wantHttpStatus:  http.StatusUnprocessableEntity,
			wantErrorCode:   domain.ErrorCodeInvalidCategory,
			wantDetailCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name: "category already exists",
			service: &fakeCategoryService{
				onAddCategory: func(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
					return nil, domain.NewConflictError("category name already used", domain.ErrorCodeCategoryNameConflict, nil)
				},
			},
			request:        rest.AddCategoryRequest{Name: "Conflict name"},
			wantHttpStatus: http.StatusConflict,
			wantErrorCode:  domain.ErrorCodeCategoryNameConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to encode request body: %v", err)
			}

			request := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
			router := test.CreateCategoryRouter(tt.service)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if tt.wantHttpStatus != recorder.Code {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if tt.wantHttpStatus == http.StatusCreated {
				checkAddCategoryResponse(t, recorder.Body, tt.request.Name)
			} else {
				test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailCodes...)
			}
		})
	}
}

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeCategoryService
		id               string
		request          rest.UpdateCategoryRequest
		wantHttpStatus   int
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onUpdateCategory: func(ctx context.Context, request *menu.UpdateCategoryRequest) error {
					return nil
				},
			},
			id: uuid.NewString(),
			request: rest.UpdateCategoryRequest{
				Name: new("New Name"),
			},
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name: "name too short",
			request: rest.UpdateCategoryRequest{
				Name: new("na"),
			},
			id:               uuid.NewString(),
			wantHttpStatus:   http.StatusUnprocessableEntity,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name: "name too long",
			request: rest.UpdateCategoryRequest{
				Name: new("CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"),
			},
			id:               uuid.NewString(),
			wantHttpStatus:   http.StatusUnprocessableEntity,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:           "update has not data",
			request:        rest.UpdateCategoryRequest{},
			id:             uuid.NewString(),
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeNoData,
		},
		{
			name: "category name already exists",
			service: &fakeCategoryService{
				onUpdateCategory: func(ctx context.Context, request *menu.UpdateCategoryRequest) error {
					return domain.NewConflictError("category name already used", domain.ErrorCodeCategoryNameConflict, nil)
				},
			},
			request: rest.UpdateCategoryRequest{
				Name: new("Used name"),
			},
			id:             uuid.NewString(),
			wantHttpStatus: http.StatusConflict,
			wantErrorCode:  domain.ErrorCodeCategoryNameConflict,
		},
		{
			name: "category not found",
			service: &fakeCategoryService{
				onUpdateCategory: func(ctx context.Context, request *menu.UpdateCategoryRequest) error {
					return domain.NewNotFoundError(
						"category not found",
						domain.ErrorCodeCategoryNotFound,
						domain.ErrorDetail{
							Code:    domain.ErrorCodeCategoryNotFoundByID,
							Message: "category not found by id",
						},
					)
				},
			},
			id: uuid.NewString(),
			request: rest.UpdateCategoryRequest{
				Name: new("New Name"),
			},
			wantHttpStatus:   http.StatusNotFound,
			wantErrorCode:    domain.ErrorCodeCategoryNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
		{
			name:    "invalid id",
			service: &fakeCategoryService{},
			id:      "invalid",
			request: rest.UpdateCategoryRequest{
				Name: new("New Name"),
			},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to encode request body: %v", err)
			}
			router := test.CreateCategoryRouter(tt.service)
			request := httptest.NewRequest(http.MethodPatch, "/categories/"+tt.id, bytes.NewReader(body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if tt.wantHttpStatus != recorder.Code {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}
			if tt.wantHttpStatus == http.StatusNoContent {
				return
			}
			test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailsCodes...)
		})
	}
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeCategoryService
		id               string
		wantHttpStatus   int
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onDeleteCategory: func(ctx context.Context, categoryID uuid.UUID) error {
					return nil
				},
			},
			id:             uuid.NewString(),
			wantHttpStatus: http.StatusOK,
		},
		{
			name: "category not found",
			service: &fakeCategoryService{
				onDeleteCategory: func(ctx context.Context, categoryID uuid.UUID) error {
					return domain.NewNotFoundError(
						"category not found",
						domain.ErrorCodeCategoryNotFound,
						domain.ErrorDetail{
							Code:    domain.ErrorCodeCategoryNotFoundByID,
							Message: "category not found by id",
						},
					)
				},
			},
			id:               uuid.NewString(),
			wantHttpStatus:   http.StatusNotFound,
			wantErrorCode:    domain.ErrorCodeCategoryNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
		{
			name: "category has linked products",
			service: &fakeCategoryService{
				onDeleteCategory: func(ctx context.Context, categoryID uuid.UUID) error {
					return domain.NewConflictError(
						"cannot delete category with linked products",
						domain.ErrorCodeCategoryHasLinkedProducts,
						nil,
					)
				},
			},

			id:             uuid.NewString(),
			wantHttpStatus: http.StatusConflict,
			wantErrorCode:  domain.ErrorCodeCategoryHasLinkedProducts,
		},
		{
			name:           "invalid id",
			service:        &fakeCategoryService{},
			id:             "invalid",
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.CreateCategoryRouter(tt.service)
			request := httptest.NewRequest(http.MethodDelete, "/categories/"+tt.id, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if tt.wantHttpStatus != recorder.Code {
				t.Fatalf("expected http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if tt.wantHttpStatus == http.StatusOK {
				return
			}

			test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailsCodes...)
		})
	}
}

func checkGetCategoriesResponse(t *testing.T, body io.Reader, wantIds []uuid.UUID) {
	var resp []rest.CategoryResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("error decoding response body: %s", err)
	}

	if len(resp) != len(wantIds) {
		t.Fatalf("expected resp to have %d items, got %d", len(wantIds), len(resp))
	}

	test.CheckEntities(t, wantIds, resp, func(response rest.CategoryResponse) uuid.UUID {
		return response.Id
	})
}

func TestCategoryHandlerGetCategories(t *testing.T) {
	tests := []struct {
		name    string
		service *fakeCategoryService
		id      *uuid.UUID
		wantIds []uuid.UUID
	}{
		{
			name: "success all",
			service: &fakeCategoryService{
				onGetCategories: func(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
					return []menu.Category{
						*test.Must(menu.ParseCategory(uuid.MustParse("11111111-1111-1111-1111-111111111111"), "Category 1")),
						*test.Must(menu.ParseCategory(uuid.MustParse("22222222-2222-2222-2222-222222222222"), "Category 2")),
					}, nil
				},
			},
			wantIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			},
		},
		{
			name: "success one",
			service: &fakeCategoryService{
				onGetCategories: func(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
					if filter.Id == nil {
						panic("filter.Id should not be nil")
					}

					if *filter.Id != uuid.MustParse("11111111-1111-1111-1111-111111111111") {
						panic("filter.Id should be: 11111111-1111-1111-1111-111111111111")
					}

					return []menu.Category{
						*test.Must(menu.ParseCategory(uuid.MustParse("11111111-1111-1111-1111-111111111111"), "Category 1")),
					}, nil
				},
			},
			id: new(uuid.MustParse("11111111-1111-1111-1111-111111111111")),
			wantIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.CreateCategoryRouter(tt.service)
			query := url.Values{}
			if tt.id != nil {
				query.Set("id", tt.id.String())
			}
			req := httptest.NewRequest(http.MethodGet, "/categories?"+query.Encode(), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("want http status %d, got %d", http.StatusOK, recorder.Code)
			}

			checkGetCategoriesResponse(t, recorder.Body, tt.wantIds)
		})
	}
}
