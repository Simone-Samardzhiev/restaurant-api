package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/testutils"
	"testing"

	"github.com/gin-gonic/gin"
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

// createAddCategoryRouter creates a gin router with /category path for [CategoryHandler.AddCategory].
func createAddCategoryRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/category", handler.AddCategory)
	return router
}

// createAddCategoryRequest creates an [http.Request] for adding a category.
func createAddCategoryRequest(t *testing.T, request *rest.AddCategoryRequest) *http.Request {
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	return httptest.NewRequest(http.MethodPost, "/category", bytes.NewReader(body))
}

// checkAddCategoryResponse checks if response body is [rest.CategoryResponse] and validates the name.
func checkAddCategoryResponse(t *testing.T, body []byte, expectedName string) {
	t.Helper()

	var response rest.CategoryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Name != expectedName {
		t.Fatalf("want name %s, got %s", expectedName, response.Name)
	}
}

func TestCategoryHandlerAddCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
					return menu.MustParseCategory(uuid.New(), "New Category"), nil
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createAddCategoryRouter(test.service)
			request := createAddCategoryRequest(t, &test.request)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if test.wantHttpStatus != recorder.Code {
				t.Fatalf("want http status %d, got %d", test.wantHttpStatus, recorder.Code)
			}

			if test.wantHttpStatus == http.StatusCreated {
				checkAddCategoryResponse(t, recorder.Body.Bytes(), test.request.Name)
			} else {
				testutils.CheckErrorResponse(t, recorder.Body, test.wantErrorCode, test.wantDetailCodes...)
			}
		})
	}
}

// createUpdateCategoryRouter creates a gin router with /category/:id for [CategoryHandler.UpdateCategory].
func createUpdateCategoryRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.PATCH("/category/:id", handler.UpdateCategory)
	return router
}

// createUpdateCategoryRequest creates a request for updating a category.
func createUpdateCategoryRequest(t *testing.T, request *rest.UpdateCategoryRequest, id uuid.UUID) *http.Request {
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}
	return httptest.NewRequest(http.MethodPatch, "/category/"+id.String(), bytes.NewBuffer(body))
}

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name             string
		service          *fakeCategoryService
		id               uuid.UUID
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
			id: uuid.New(),
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
			id:               uuid.New(),
			wantHttpStatus:   http.StatusUnprocessableEntity,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name: "name too long",
			request: rest.UpdateCategoryRequest{
				Name: new("CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"),
			},
			id:               uuid.New(),
			wantHttpStatus:   http.StatusUnprocessableEntity,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:           "update has not data",
			request:        rest.UpdateCategoryRequest{},
			id:             uuid.New(),
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeNoData,
		},
		{
			name: "category already exists",
			service: &fakeCategoryService{
				onUpdateCategory: func(ctx context.Context, request *menu.UpdateCategoryRequest) error {
					return domain.NewConflictError("category name already used", domain.ErrorCodeCategoryNameConflict, nil)
				},
			},
			request: rest.UpdateCategoryRequest{
				Name: new("Used name"),
			},
			id:             uuid.New(),
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
			id: uuid.New(),
			request: rest.UpdateCategoryRequest{
				Name: new("New Name"),
			},
			wantHttpStatus:   http.StatusNotFound,
			wantErrorCode:    domain.ErrorCodeCategoryNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createUpdateCategoryRouter(test.service)
			request := createUpdateCategoryRequest(t, &test.request, test.id)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if test.wantHttpStatus != recorder.Code {
				t.Fatalf("want http status %d, got %d", test.wantHttpStatus, recorder.Code)
			}
			if test.wantHttpStatus == http.StatusNoContent {
				return
			}

			testutils.CheckErrorResponse(t, recorder.Body, test.wantErrorCode, test.wantDetailsCodes...)
		})
	}
}

// createUpdateCategoryRouter creates a gin router with /category/:id for [CategoryHandler.DeleteCategory].
func createDeleteCategoryRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.DELETE("/category/:id", handler.DeleteCategory)
	return router
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name             string
		service          *fakeCategoryService
		id               uuid.UUID
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
			id:             uuid.New(),
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
			id:               uuid.New(),
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

			id:             uuid.New(),
			wantHttpStatus: http.StatusConflict,
			wantErrorCode:  domain.ErrorCodeCategoryHasLinkedProducts,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createDeleteCategoryRouter(test.service)
			request := httptest.NewRequest(http.MethodDelete, "/category/"+test.id.String(), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if test.wantHttpStatus != recorder.Code {
				t.Fatalf("expected http status %d, got %d", test.wantHttpStatus, recorder.Code)
			}

			if test.wantHttpStatus == http.StatusOK {
				return
			}

			testutils.CheckErrorResponse(t, recorder.Body, test.wantErrorCode, test.wantDetailsCodes...)
		})
	}
}

func createGetCategoriesRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.GET("/categories", handler.GetCategories)
	return router
}

func checkGetCategoriesResponse(t *testing.T, body []byte, expectedIds []uuid.UUID) {
	var response []rest.CategoryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("error decoding response body: %s", err)
	}

	if len(response) != len(expectedIds) {
		t.Fatalf("expected response to have %d items, got %d", len(expectedIds), len(response))
	}

	counter := make(map[uuid.UUID]int)
	for _, id := range expectedIds {
		counter[id]++
	}

	for _, category := range response {
		if counter[category.Id] == 0 {
			t.Errorf("unexpected category id: %s", category.Id)
			continue
		}

		counter[category.Id]--
	}

	for id, count := range counter {
		if count != 0 {
			t.Errorf("missing category with id: %s", id)
		}
	}

}

func TestCategoryHandlerGetCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
						*menu.MustParseCategory(uuid.MustParse("11111111-1111-1111-1111-111111111111"), "Category 1"),
						*menu.MustParseCategory(uuid.MustParse("22222222-2222-2222-2222-222222222222"), "Category 2"),
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
						*menu.MustParseCategory(uuid.MustParse("11111111-1111-1111-1111-111111111111"), "Category 1"),
					}, nil
				},
			},
			id: new(uuid.MustParse("11111111-1111-1111-1111-111111111111")),
			wantIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createGetCategoriesRouter(test.service)
			query := url.Values{}
			if test.id != nil {
				query.Set("id", test.id.String())
			}
			req := httptest.NewRequest(http.MethodGet, "/categories?"+query.Encode(), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("want http status %d, got %d", http.StatusOK, recorder.Code)
			}

			checkGetCategoriesResponse(t, recorder.Body.Bytes(), test.wantIds)
		})
	}
}
