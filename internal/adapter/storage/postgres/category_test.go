package postgres_test

import (
	"errors"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"
	_ "embed"

	"github.com/google/uuid"
)

//go:embed testdata/seeds/menu.sql
var seedMenuTablesQuery string

func seedMenuTables(t *testing.T) {
	t.Helper()

	if _, err := database.Exec(seedMenuTablesQuery); err != nil {
		t.Fatalf("error seeding menu tables: %v", err)
	}

	t.Cleanup(func() {
		if _, err := database.Exec(`TRUNCATE TABLE products, product_categories RESTART IDENTITY CASCADE `); err != nil {
			t.Fatalf("error truncating tables: %v", err)
		}
	})
}

func TestCategoryRepositoryAddCategory(t *testing.T) {
	tests := []struct {
		name              string
		request           *menu.AddCategoryRequest
		wantErr           bool
		expectedErrorKind domain.ErrorKind
		expectedErrorCode domain.ErrorCode
	}{
		{
			name:    "success",
			request: menu.MustParseAddCategoryRequest("New category"),
		},
		{
			name:              "name already exists",
			request:           menu.MustParseAddCategoryRequest("Appetizers"),
			wantErr:           true,
			expectedErrorKind: domain.ErrorKindConflict,
			expectedErrorCode: domain.ErrorCodeCategoryNameConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			result, err := repo.AddCategory(context.Background(), test.request)

			if test.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				domainErr, ok := errors.AsType[*domain.Error](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if domainErr.Kind != test.expectedErrorKind {
					t.Errorf("expected error kind %v, got %v", test.expectedErrorKind, domainErr.Kind)
				}

				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %v, got %v", test.expectedErrorCode, domainErr.Code)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
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
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				domainErr, ok := errors.AsType[*domain.Error](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if domainErr.Kind != test.expectedErrorKind {
					t.Errorf("expected error kind %v, got %v", test.expectedErrorKind, domainErr.Kind)
				}
				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %v, got %v", test.expectedErrorCode, domainErr.Code)
				}

				if test.expectedDetailsCodes != nil {
					matchErrorCodes(t, test.expectedDetailsCodes, domainErr.Details)
				}
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
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				domainErr, ok := errors.AsType[*domain.Error](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if domainErr.Kind != test.expectedErrorKind {
					t.Errorf("expected error kind %v, got %v", test.expectedErrorKind, domainErr.Kind)
				}

				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %v, got %v", test.expectedErrorCode, domainErr.Code)
				}

				if test.expectedDetailsCodes != nil {
					matchErrorCodes(t, test.expectedDetailsCodes, domainErr.Details)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
			}
		})
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
				t.Fatalf("expected no err, got : %v", err)
			}

			if len(categories) != len(test.expectedIds) {
				t.Errorf("expected %d categories, got %d", len(test.expectedIds), len(categories))
			}

			counter := map[uuid.UUID]int{}
			for _, category := range test.expectedIds {
				counter[category]++
			}

			for _, category := range categories {
				if counter[category.Id] == 0 {
					t.Errorf("unexpected category id: %s", category.Id)
					continue
				}
				counter[category.Id]--
			}

			for id, count := range counter {
				if count != 0 {
					t.Errorf("missing category id: %s", id)
				}
			}
		})
	}
}
