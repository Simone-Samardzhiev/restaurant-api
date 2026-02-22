package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/testutils"
	"testing"

	"github.com/google/uuid"
)

func TestParseCategoryName(t *testing.T) {
	tests := []struct {
		name          string
		categoryName  string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:         "valid name",
			categoryName: "New Category",
		},
		{
			name:          "short name",
			categoryName:  "ca",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeCategoryNameTooShort,
		},
		{
			name:          "long name",
			categoryName:  "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeCategoryNameTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseCategoryName(test.categoryName)
			if test.wantErr {
				testutils.AssertErrorDetail(t, err, test.wantErrorCode)
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

func TestParseCategory(t *testing.T) {
	tests := []struct {
		name             string
		categoryName     string
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:         "valid request",
			categoryName: "New Category",
		},
		{
			name:             "short name",
			categoryName:     "ca",
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:             "long name",
			categoryName:     "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseCategory(uuid.New(), test.categoryName)

			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got : %v", err)
			}

			if parsed.Name.String() != test.categoryName {
				t.Fatalf("want category name %s, got %s", test.categoryName, parsed.Name.String())
			}
		})
	}
}

func TestParseAddCategoryRequest(t *testing.T) {
	tests := []struct {
		name             string
		categoryName     string
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:         "valid request",
			categoryName: "New Category",
		},
		{
			name:             "short name",
			categoryName:     "ca",
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:             "long name",
			categoryName:     "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseAddCategoryRequest(test.categoryName)

			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			if parsed.Name.String() != test.categoryName {
				t.Fatalf("want category name %s, got %s", test.categoryName, parsed.Name.String())
			}
		})
	}
}

func TestParseUpdateCategoryRequest(t *testing.T) {
	tests := []struct {
		name             string
		categoryName     *string
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:         "valid request",
			categoryName: new("New Category"),
		},
		{
			name:             "short name",
			categoryName:     new("na"),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:             "long name",
			categoryName:     new("CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory"),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindValidation,
			wantErrorCode:    domain.ErrorCodeInvalidCategoryUpdate,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
		{
			name:          "no data in update",
			categoryName:  nil,
			wantErr:       true,
			wantErrorKind: domain.ErrorKindBadRequest,
			wantErrorCode: domain.ErrorCodeNoData,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := menu.ParseUpdateCategoryRequest(uuid.New(), test.categoryName)
			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got : %v", err)
			}

			if *test.categoryName != result.Name.String() {
				t.Fatalf("want category name %s, got %s", *test.categoryName, result.Name.String())
			}
		})
	}
}
