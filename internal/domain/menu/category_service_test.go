package menu_test

import (
	"errors"
	"fmt"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"

	"github.com/google/uuid"
)

type FakeRepository struct {
	OnAddCategory    func(ctx context.Context, category *menu.Category) error
	OnUpdateCategory func(ctx context.Context, category *menu.CategoryUpdate) error
	OnDeleteCategory func(ctx context.Context, id uuid.UUID) error
	OnGetCategories  func(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error)
}

var _ menu.CategoryRepository = (*FakeRepository)(nil)

func (r *FakeRepository) AddCategory(ctx context.Context, category *menu.Category) error {
	if r.OnAddCategory != nil {
		return r.OnAddCategory(ctx, category)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeRepository) UpdateCategory(ctx context.Context, update *menu.CategoryUpdate) error {
	if r.OnUpdateCategory != nil {
		return r.OnUpdateCategory(ctx, update)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if r.OnDeleteCategory != nil {
		return r.OnDeleteCategory(ctx, id)
	}

	return fmt.Errorf("not implemented")
}

func (r *FakeRepository) GetCategories(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
	if r.OnGetCategories != nil {
		return r.OnGetCategories(ctx, filter)
	}

	return nil, fmt.Errorf("not implemented")
}

// asPointer is a helper function to pass values a pointers without creating a variable.
func asPointer[T any](v T) *T {
	return &v
}

// assertNoErr is a helper function for asserting no error is returned.
func assertNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// assertValidationErrors is a helper function for asserting an error is domain.ValidationErrors.
func asserValidationsErr(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error")
	}

	var validationErr *domain.ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

// assertBadRequestErr is a helper function for asserting an error is domain.Error
// and the type is domain.BadRequest.
func assertBadRequestErr(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error")
	}

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain error, got: %v", err)
	}

	if domainErr.Type != domain.BadRequest {
		t.Fatalf("expected bad request error, got: %v", err)
	}
}

// assertInternalErr is a helper function for asserting an error is domain>error
// and the type is domain.BadRequest.
func assertInternalErr(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error")
	}

	var internalErr *domain.InternalError
	if !errors.As(err, &internalErr) {
		t.Fatalf("expected internal error, got: %v", err)
	}
}

func TestDefaultCategoryServiceAddCategory(t *testing.T) {
	tests := []struct {
		name           string
		fakeRepository *FakeRepository
		request        *menu.AddCategoryRequest
		checkErr       func(t *testing.T, err error)
		checkResult    func(t *testing.T, category *menu.Category)
	}{
		{
			name: "success",
			fakeRepository: &FakeRepository{
				OnAddCategory: func(ctx context.Context, category *menu.Category) error {
					return nil
				},
			},
			request:  menu.NewAddCategoryRequest("Appetizers"),
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, category *menu.Category) {
				t.Helper()

				if category == nil {
					t.Fatalf("category should not be nil")
				}

				if category.Name.String() != "Appetizers" {
					t.Fatalf("expected name `Appetizers`, got %s ", category.Name.String())
				}
			},
		},
		{
			name:           "invalid category name",
			fakeRepository: &FakeRepository{},
			request:        menu.NewAddCategoryRequest(""),
			checkErr:       asserValidationsErr,
			checkResult: func(t *testing.T, category *menu.Category) {
				t.Helper()

				if category != nil {
					t.Fatalf("category should be nil")
				}
			},
		},
		{
			name: "internal error",
			fakeRepository: &FakeRepository{
				OnAddCategory: func(ctx context.Context, category *menu.Category) error {
					return domain.NewInternalError("error adding category", fmt.Errorf("tt error"))
				},
			},
			request:  menu.NewAddCategoryRequest("Appetizers"),
			checkErr: assertInternalErr,
			checkResult: func(t *testing.T, category *menu.Category) {
				t.Helper()

				if category == nil {
					t.Fatalf("category should be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultCategoryService(tt.fakeRepository)
			result, err := service.AddCategory(context.Background(), tt.request)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func TestDefaultCategoryServiceUpdateCategory(t *testing.T) {

	tests := []struct {
		name           string
		fakeRepository *FakeRepository
		request        *menu.CategoryUpdateRequest
		checkErr       func(t *testing.T, err error)
	}{
		{
			name: "success",
			fakeRepository: &FakeRepository{
				OnUpdateCategory: func(ctx context.Context, category *menu.CategoryUpdate) error {
					return nil
				},
			},
			request:  menu.NewCategoryUpdateRequest(uuid.New(), asPointer("Appetizers")),
			checkErr: assertNoErr,
		},
		{
			name:           "invalid category name",
			fakeRepository: &FakeRepository{},
			request:        menu.NewCategoryUpdateRequest(uuid.New(), new(string)),
			checkErr:       asserValidationsErr,
		},
		{
			name: "internal error",
			fakeRepository: &FakeRepository{
				OnUpdateCategory: func(ctx context.Context, category *menu.CategoryUpdate) error {
					return domain.NewInternalError("error updating category", fmt.Errorf("tt error"))
				},
			},
			request:  menu.NewCategoryUpdateRequest(uuid.New(), asPointer("Appetizers")),
			checkErr: assertInternalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := menu.NewDefaultCategoryService(tt.fakeRepository)
			err := service.UpdateCategory(context.Background(), tt.request)
			tt.checkErr(t, err)
		})
	}
}
