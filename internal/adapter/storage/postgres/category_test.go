package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"
	_ "embed"

	"restaurant/internal/test"

	"github.com/google/uuid"
)

func TestCategoryRepositoryAddCategory(t *testing.T) {
	tests := []struct {
		name          string
		request       *menu.AddCategoryRequest
		wantErr       bool
		wantErrorKind domain.ErrorKind
		wantErrorCode domain.ErrorCode
	}{
		{
			name:    "success",
			request: test.Must(menu.ParseAddCategoryRequest("New category")),
		},
		{
			name:          "name already exists",
			request:       test.Must(menu.ParseAddCategoryRequest("Appetizers")),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindConflict,
			wantErrorCode: domain.ErrorCodeCategoryNameConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			repo := postgres.NewCategoryRepository(database)
			result, err := repo.SaveCategory(context.Background(), tt.request)

			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			if tt.request.Name.String() != result.Name.String() {
				t.Errorf("wan name %v, got %v", tt.request.Name.String(), result.Name.String())
			}
		})
	}
}

func TestCategoryRepositoryUpdateCategory(t *testing.T) {
	tests := []struct {
		name                 string
		request              *menu.UpdateCategoryRequest
		wantErr              bool
		expectedErrorKind    domain.ErrorKind
		expectedErrorCode    domain.ErrorCode
		expectedDetailsCodes []domain.ErrorCode
	}{
		{
			name:    "success",
			request: test.Must(menu.ParseUpdateCategoryRequest(uuid.MustParse("11111111-1111-1111-1111-111111111111"), new("New name"))),
			wantErr: false,
		},
		{
			name:              "name already exists",
			request:           test.Must(menu.ParseUpdateCategoryRequest(uuid.MustParse("22222222-2222-2222-2222-222222222222"), new("Appetizers"))),
			wantErr:           true,
			expectedErrorKind: domain.ErrorKindConflict,
			expectedErrorCode: domain.ErrorCodeCategoryNameConflict,
		},
		{
			name:                 "not found",
			request:              test.Must(menu.ParseUpdateCategoryRequest(uuid.New(), new("New name"))),
			wantErr:              true,
			expectedErrorKind:    domain.ErrorKindNotFound,
			expectedErrorCode:    domain.ErrorCodeCategoryNotFound,
			expectedDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			repo := postgres.NewCategoryRepository(database)
			err := repo.UpdateCategory(context.Background(), tt.request)
			if tt.wantErr {
				test.AssertError(t, err, tt.expectedErrorKind, tt.expectedErrorCode, tt.expectedDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}
		})
	}
}

func TestCategoryRepositoryDeleteCategory(t *testing.T) {
	tests := []struct {
		name                 string
		id                   uuid.UUID
		wantErr              bool
		expectedErrorKind    domain.ErrorKind
		expectedErrorCode    domain.ErrorCode
		expectedDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			id:   uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		},
		{
			name:                 "not found",
			id:                   uuid.New(),
			wantErr:              true,
			expectedErrorKind:    domain.ErrorKindNotFound,
			expectedErrorCode:    domain.ErrorCodeCategoryNotFound,
			expectedDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
		{
			name:              "category has linked products",
			id:                uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			wantErr:           true,
			expectedErrorKind: domain.ErrorKindConflict,
			expectedErrorCode: domain.ErrorCodeCategoryHasLinkedProducts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			repo := postgres.NewCategoryRepository(database)
			err := repo.DeleteCategory(context.Background(), tt.id)
			if tt.wantErr {
				test.AssertError(t, err, tt.expectedErrorKind, tt.expectedErrorCode, tt.expectedDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}
		})
	}
}

func TestCategoryRepositoryGetCategories(t *testing.T) {
	tests := []struct {
		name    string
		filter  *menu.CategoryFilter
		wantIds []uuid.UUID
	}{
		{
			name:   "success all",
			filter: &menu.CategoryFilter{},
			wantIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				uuid.MustParse("66666666-6666-6666-6666-666666666666"),
			},
		},
		{
			name: "success filter",
			filter: &menu.CategoryFilter{
				Id: new(uuid.MustParse("11111111-1111-1111-1111-111111111111")),
			},
			wantIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			repo := postgres.NewCategoryRepository(database)
			categories, err := repo.GetCategories(context.Background(), tt.filter)
			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			test.CheckEntities(t, tt.wantIds, categories, func(category menu.Category) uuid.UUID {
				return category.Id
			})
		})
	}
}
