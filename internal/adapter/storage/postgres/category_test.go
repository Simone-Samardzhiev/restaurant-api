package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"
	_ "embed"

	"restaurant/internal/testutils"

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
			request: menu.MustParseAddCategoryRequest("New category"),
		},
		{
			name:          "name already exists",
			request:       menu.MustParseAddCategoryRequest("Appetizers"),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindConflict,
			wantErrorCode: domain.ErrorCodeCategoryNameConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			result, err := repo.AddCategory(context.Background(), test.request)

			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got : %v", err)
			}

			if test.request.Name.String() != result.Name.String() {
				t.Errorf("expected name %v, got %v", test.request.Name.String(), result.Name.String())
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
			request: menu.MustParseUpdateCategoryRequest(uuid.MustParse("11111111-1111-1111-1111-111111111111"), new("New name")),
			wantErr: false,
		},
		{
			name:              "name already exists",
			request:           menu.MustParseUpdateCategoryRequest(uuid.MustParse("22222222-2222-2222-2222-222222222222"), new("Appetizers")),
			wantErr:           true,
			expectedErrorKind: domain.ErrorKindConflict,
			expectedErrorCode: domain.ErrorCodeCategoryNameConflict,
		},
		{
			name:                 "not found",
			request:              menu.MustParseUpdateCategoryRequest(uuid.New(), new("New name")),
			wantErr:              true,
			expectedErrorKind:    domain.ErrorKindNotFound,
			expectedErrorCode:    domain.ErrorCodeCategoryNotFound,
			expectedDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			err := repo.UpdateCategory(context.Background(), test.request)
			if test.wantErr {
				testutils.AssertError(t, err, test.expectedErrorKind, test.expectedErrorCode, test.expectedDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			err := repo.DeleteCategory(context.Background(), test.id)
			if test.wantErr {
				testutils.AssertError(t, err, test.expectedErrorKind, test.expectedErrorCode, test.expectedDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
			}
		})
	}
}

func checkCategories(t *testing.T, expectedIds []uuid.UUID, categories []menu.Category) {
	t.Helper()

	counter := make(map[uuid.UUID]int)
	for _, id := range expectedIds {
		counter[id]++
	}

	for _, category := range categories {
		if counter[category.Id] == 0 {
			t.Errorf("unexpected category id %s", category.Id)
			continue
		}
		counter[category.Id]--
	}

	for id, count := range counter {
		if count != 0 {
			t.Errorf("missing category id %s", id)
		}
	}
}

func TestCategoryRepositoryGetCategories(t *testing.T) {
	tests := []struct {
		name        string
		filter      *menu.CategoryFilter
		expectedIds []uuid.UUID
	}{
		{
			name:   "success all",
			filter: &menu.CategoryFilter{},
			expectedIds: []uuid.UUID{
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
			expectedIds: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			categories, err := repo.GetCategories(context.Background(), test.filter)
			if err != nil {
				t.Fatalf("expected no error, got : %v", err)
			}

			checkCategories(t, test.expectedIds, categories)
		})
	}
}
