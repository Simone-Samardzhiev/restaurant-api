package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeCategoryService struct {
	onAddCategory    func(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error)
	onUpdateCategory func(ctx context.Context, request *menu.UpdateCategoryRequest) error
	onDeleteCategory func(ctx context.Context, id uuid.UUID) error
}

var _ menu.CategoryService = (*fakeCategoryService)(nil)

func (s *fakeCategoryService) AddCategory(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
	if s.onAddCategory == nil {
		panic("onAddCategory function is not defined")
	}
	return s.onAddCategory(ctx, request)
}

func (s *fakeCategoryService) UpdateCategory(ctx context.Context, request *menu.UpdateCategoryRequest) error {
	if s.onUpdateCategory == nil {
		panic("onUpdateCategory function is not defined")
	}
	return s.onUpdateCategory(ctx, request)
}

func (s *fakeCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if s.onDeleteCategory == nil {
		panic("onDeleteCategory function is not defined")
	}
	return s.onDeleteCategory(ctx, id)
}

func TestCategoryHandlerAddCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		service            *fakeCategoryService
		request            rest.AddCategoryRequest
		expectedHttpStatus int
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onAddCategory: func(_ context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
					return menu.MustParseCategory(uuid.New(), "New Category"), nil
				},
			},
			request:            rest.AddCategoryRequest{Name: "New Category"},
			expectedHttpStatus: http.StatusCreated,
		},
		{
			name: "name too short",
			request: rest.AddCategoryRequest{
				Name: "Ne",
			},
			expectedHttpStatus: http.StatusUnprocessableEntity,
			expectedErrorCode:  domain.ErrorCodeInvalidCategory,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:               "name too long",
			request:            rest.AddCategoryRequest{Name: "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"},
			expectedHttpStatus: http.StatusUnprocessableEntity,
			expectedErrorCode:  domain.ErrorCodeInvalidCategory,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:    "category already exists",
			request: rest.AddCategoryRequest{Name: "Duplicate name"},
			service: &fakeCategoryService{
				onAddCategory: func(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
					return nil, domain.NewConflictError("category name already used", domain.ErrorCodeCategoryNameConflict, nil)
				},
			},
			expectedHttpStatus: http.StatusConflict,
			expectedErrorCode:  domain.ErrorCodeCategoryNameConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := rest.NewCategoryHandler(test.service)

			router := gin.New()
			router.Use(middleware.Error())
			router.POST("/category", handler.AddCategory)

			body, err := json.Marshal(test.request)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/category", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			router.ServeHTTP(recorder, req)
			if test.expectedHttpStatus != recorder.Code {
				t.Fatalf("expect http status %d, got %d", test.expectedHttpStatus, recorder.Code)
			}

			if test.expectedHttpStatus != recorder.Code {
				t.Fatalf("expect http status %d, got %d", test.expectedHttpStatus, recorder.Code)
			}

			if test.expectedHttpStatus == http.StatusCreated {
				var response rest.AddCategoryResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				if response.Name != test.request.Name {
					t.Fatalf("expect name %s, got %s", test.request.Name, response.Name)
				}
			} else {
				var response middleware.ErrorResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				if test.expectedErrorCode.String() != response.Code {
					t.Fatalf("expect code %s, got %s", test.expectedErrorCode, response.Code)
				}

				if test.expectedErrorCodes != nil {
					matchErrorCodes(t, test.expectedErrorCodes, response.Details)
				}
			}
		})
	}
}

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		service            *fakeCategoryService
		id                 uuid.UUID
		request            rest.UpdateCategoryRequest
		expectedHttpStatus int
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
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
			expectedHttpStatus: http.StatusNoContent,
		},
		{
			name: "name too short",
			request: rest.UpdateCategoryRequest{
				Name: new("na"),
			},
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusUnprocessableEntity,
			expectedErrorCode:  domain.ErrorCodeInvalidCategoryUpdate,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name: "name too long",
			request: rest.UpdateCategoryRequest{
				Name: new("CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"),
			},
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusUnprocessableEntity,
			expectedErrorCode:  domain.ErrorCodeInvalidCategoryUpdate,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:               "update has not data",
			request:            rest.UpdateCategoryRequest{},
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusBadRequest,
			expectedErrorCode:  domain.ErrorCodeNoData,
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
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusConflict,
			expectedErrorCode:  domain.ErrorCodeCategoryNameConflict,
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
			expectedHttpStatus: http.StatusNotFound,
			expectedErrorCode:  domain.ErrorCodeCategoryNotFound,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := rest.NewCategoryHandler(test.service)
			router := gin.New()
			router.Use(middleware.Error())
			router.PATCH("/category/:id", handler.UpdateCategory)

			body, err := json.Marshal(test.request)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPatch, "/category/"+test.id.String(), bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			router.ServeHTTP(recorder, req)

			if test.expectedHttpStatus != recorder.Code {
				t.Fatalf("expect http status %d, got %d", test.expectedHttpStatus, recorder.Code)
			}

			if test.expectedHttpStatus == http.StatusNoContent {
				return
			}

			var response middleware.ErrorResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if test.expectedErrorCode.String() != response.Code {
				t.Fatalf("expect code %s, got %s", test.expectedErrorCode, response.Code)
			}

			if test.expectedErrorCodes != nil {
				matchErrorCodes(t, test.expectedErrorCodes, response.Details)
			}
		})
	}
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		service            *fakeCategoryService
		id                 uuid.UUID
		expectedHttpStatus int
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeCategoryService{
				onDeleteCategory: func(ctx context.Context, categoryID uuid.UUID) error {
					return nil
				},
			},
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusOK,
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
			id:                 uuid.New(),
			expectedHttpStatus: http.StatusNotFound,
			expectedErrorCode:  domain.ErrorCodeCategoryNotFound,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
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

			id:                 uuid.New(),
			expectedHttpStatus: http.StatusConflict,
			expectedErrorCode:  domain.ErrorCodeCategoryHasLinkedProducts,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := rest.NewCategoryHandler(test.service)
			router := gin.New()
			router.Use(middleware.Error())
			router.DELETE("/category/:id", handler.DeleteCategory)

			body, err := json.Marshal(test.id)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodDelete, "/category/"+test.id.String(), bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			router.ServeHTTP(recorder, req)

			if test.expectedHttpStatus != recorder.Code {
				t.Fatalf("expect http status %d, got %d", test.expectedHttpStatus, recorder.Code)
			}

			if test.expectedHttpStatus == http.StatusOK {
				return
			}

			var response middleware.ErrorResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)

			if err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if test.expectedErrorCode.String() != response.Code {
				t.Fatalf("expect code %s, got %s", test.expectedErrorCode, response.Code)
			}

			if test.expectedErrorCodes != nil {
				matchErrorCodes(t, test.expectedErrorCodes, response.Details)
			}
		})
	}
}
