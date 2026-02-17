package menu_test

import (
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"github.com/google/uuid"
)

func TestParseCategoryName(t *testing.T) {
	tests := []struct {
		name              string
		categoryName      string
		wantErr           bool
		expectedErrorCode domain.ErrorCode
	}{
		{
			name:         "valid name",
			categoryName: "New Category",
		},
		{
			name:              "short name",
			categoryName:      "ca",
			wantErr:           true,
			expectedErrorCode: domain.ErrorCodeCategoryNameTooShort,
		},
		{
			name:              "long name",
			categoryName:      "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:           true,
			expectedErrorCode: domain.ErrorCodeCategoryNameTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseCategoryName(test.categoryName)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}

				domainErr, ok := errors.AsType[*domain.ErrorDetail](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %v, got %v", test.expectedErrorCode, domainErr.Code)
				}
				return
			}

			if err != nil {
				return
			}
			if parsed.String() != test.categoryName {
				t.Fatalf("want name %s, got %s", test.categoryName, parsed.String())
			}
		})
	}
}

func TestParseAddCategoryRequest(t *testing.T) {
	tests := []struct {
		name               string
		categoryName       string
		wantErr            bool
		expectedErrorKind  domain.ErrorKind
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
	}{
		{
			name:         "valid request",
			categoryName: "New Category",
		},
		{
			name:               "short name",
			categoryName:       "ca",
			wantErr:            true,
			expectedErrorKind:  domain.ErrorKindValidation,
			expectedErrorCode:  domain.ErrorCodeInvalidCategory,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:               "long name",
			categoryName:       "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:            true,
			expectedErrorKind:  domain.ErrorKindValidation,
			expectedErrorCode:  domain.ErrorCodeInvalidCategory,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseAddCategoryRequest(test.categoryName)

			if test.wantErr {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}

				domainErr, ok := errors.AsType[*domain.Error](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if test.expectedErrorKind != domainErr.Kind {
					t.Errorf("expected kind %s, got %s", test.expectedErrorKind, domainErr.Kind)
				}
				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %s, got %s", test.expectedErrorCode, domainErr.Code)
				}

				if test.expectedErrorCodes != nil {
					matchErrorCodes(t, test.expectedErrorCodes, domainErr.Details)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
			}

			if parsed.Name.String() != test.categoryName {
				t.Fatalf("expected name %s, got %s", test.categoryName, parsed.Name.String())
			}
		})
	}
}

func TestParseUpdateCategoryRequest(t *testing.T) {
	tests := []struct {
		name               string
		categoryName       *string
		wantErr            bool
		expectedErrorKind  domain.ErrorKind
		expectedErrorCode  domain.ErrorCode
		expectedErrorCodes []domain.ErrorCode
	}{
		{
			name:         "valid request",
			categoryName: new("New Category"),
		},
		{
			name:               "short name",
			categoryName:       new("na"),
			wantErr:            true,
			expectedErrorKind:  domain.ErrorKindValidation,
			expectedErrorCode:  domain.ErrorCodeInvalidCategoryUpdate,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:               "long name",
			categoryName:       new("CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"),
			wantErr:            true,
			expectedErrorKind:  domain.ErrorKindValidation,
			expectedErrorCode:  domain.ErrorCodeInvalidCategoryUpdate,
			expectedErrorCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:              "no data in update",
			categoryName:      nil,
			wantErr:           true,
			expectedErrorKind: domain.ErrorKindBadRequest,
			expectedErrorCode: domain.ErrorCodeNoData,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := menu.ParseUpdateCategoryRequest(uuid.New(), test.categoryName)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected err, got none")
				}

				domainErr, ok := errors.AsType[*domain.Error](err)
				if !ok {
					t.Fatalf("expected domain error, got %T", err)
				}

				if test.expectedErrorKind != domainErr.Kind {
					t.Errorf("expected kind %s, got %s", test.expectedErrorKind, domainErr.Kind)
				}
				if test.expectedErrorCode != domainErr.Code {
					t.Errorf("expected error code %s, got %s", test.expectedErrorCode, domainErr.Code)
				}

				if test.expectedErrorCodes != nil {
					matchErrorCodes(t, test.expectedErrorCodes, domainErr.Details)

				}
				return
			}

			if err != nil {
				t.Fatalf("expected no err, got : %v", err)
			}

			if *test.categoryName != result.Name.String() {
				t.Fatalf("expected name %s, got %s", *test.categoryName, result.Name.String())
			}
		})
	}
}
