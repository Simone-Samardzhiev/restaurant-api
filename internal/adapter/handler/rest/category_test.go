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
	onAddCategory func(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error)
}

func (s *fakeCategoryService) AddCategory(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
	if s.onAddCategory == nil {
		panic("onAddCategory function is not defined")
	}
	return s.onAddCategory(ctx, request)
}

var _ menu.CategoryService = (*fakeCategoryService)(nil)

func TestCategoryHandlerAddCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		service            *fakeCategoryService
		request            rest.AddCategoryRequest
		expectHttpStatus   int
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
			request:          rest.AddCategoryRequest{Name: "New Category"},
			expectHttpStatus: http.StatusCreated,
		},
		{
			name: "name too short",
			request: rest.AddCategoryRequest{
				Name: "Ne",
			},
			expectHttpStatus:   http.StatusUnprocessableEntity,
			expectedErrorCode:  domain.ErrorCodeInvalidCategory,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:               "name too long",
			request:            rest.AddCategoryRequest{Name: "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"},
			expectHttpStatus:   http.StatusUnprocessableEntity,
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
			expectHttpStatus:  http.StatusConflict,
			expectedErrorCode: domain.ErrorCodeCategoryNameConflict,
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
			if test.expectHttpStatus != recorder.Code {
				t.Fatalf("expect http status %d, got %d", test.expectHttpStatus, recorder.Code)
			}

			if test.expectHttpStatus == http.StatusCreated {
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
