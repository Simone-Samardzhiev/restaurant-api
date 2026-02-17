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
					t.Fatal("expected error, got none")
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
		name               string
		request            *menu.UpdateCategoryRequest
		wantErr            bool
		expectedErrorKind  domain.ErrorKind
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
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
			name:               "not found",
			request:            menu.MustParseUpdateCategoryRequest(uuid.New(), new("New name")),
			wantErr:            true,
			expectedErrorKind:  domain.ErrorKindNotFound,
			expectedErrorCode:  domain.ErrorCodeCategoryNotFound,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewCategoryRepository(database)
			err := repo.UpdateCategory(context.Background(), test.request)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error, got none")
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

				if test.expectedErrorCodes != nil {
					matchErrorCodes(t, test.expectedErrorCodes, domainErr.Details)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
			}
		})
	}
}
