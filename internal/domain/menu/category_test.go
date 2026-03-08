package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseCategoryName(tt.categoryName)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}

			if err != nil {
				return
			}
			if parsed.String() != tt.categoryName {
				t.Fatalf("want name %s, got %s", tt.categoryName, parsed.String())
			}
		})
	}
}

func TestParseCategory(t *testing.T) {
	tests := []struct {
		name             string
		categoryName     string
		wantErr          bool
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
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:             "long name",
			categoryName:     "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseCategory(uuid.New(), tt.categoryName)

			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got : %v", err)
			}

			if parsed.Name.String() != tt.categoryName {
				t.Fatalf("want category name %s, got %s", tt.categoryName, parsed.Name.String())
			}
		})
	}
}

func TestParseAddCategoryRequest(t *testing.T) {
	tests := []struct {
		name             string
		categoryName     string
		wantErr          bool
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
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooShort},
		},
		{
			name:             "long name",
			categoryName:     "CategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategoryCategory",
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidCategory,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNameTooLong},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseAddCategoryRequest(tt.categoryName)

			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			if parsed.Name.String() != tt.categoryName {
				t.Fatalf("want category name %s, got %s", tt.categoryName, parsed.Name.String())
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := menu.ParseUpdateCategoryRequest(uuid.New(), tt.categoryName)
			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got : %v", err)
			}

			if *tt.categoryName != result.Name.String() {
				t.Fatalf("want category name %s, got %s", *tt.categoryName, result.Name.String())
			}
		})
	}
}
