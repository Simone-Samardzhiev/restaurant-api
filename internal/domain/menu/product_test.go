package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestParseProductName(t *testing.T) {
	tests := []struct {
		name          string
		productName   string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:        "success",
			productName: "Valid product name",
		},
		{
			name:          "short name",
			productName:   "na                          ",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeProductNameTooShort,
		},
		{
			name:          "long name",
			productName:   "ProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductName",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeProductNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseProductName(tt.productName)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}

			if tt.productName != parsed.String() {
				t.Fatalf("want name %s, got %s", tt.productName, parsed.String())
			}
		})
	}
}

func TestParseProductDescription(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:        "success",
			description: "Valid product description",
		}, {
			name:          "short description",
			description:   "short                     ",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeProductDescriptionTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := menu.ParseProductDescription(tt.description)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			if tt.description != parsed.String() {
				t.Fatalf("want description %s, got %s", tt.description, parsed.String())
			}
		})
	}
}

func TestParseProductPrice(t *testing.T) {
	tests := []struct {
		name          string
		price         decimal.Decimal
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:  "success",
			price: decimal.NewFromFloat(10.10),
		},
		{
			name:          "invalid price",
			price:         decimal.NewFromFloat(-10),
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeProductPriceLessThanZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := menu.ParseProductPrice(tt.price)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			if !tt.price.Equal(parsed.Value()) {
				t.Fatalf("want price %s, got %s", tt.price, parsed.Value().String())
			}
		})
	}
}

func checkProduct(t *testing.T, product *menu.Product, wantName, wantDescription string, wantPrice decimal.Decimal) {
	t.Helper()

	if wantName != product.Name.String() {
		t.Errorf("want name %s, got %s", wantName, product.Name.String())
	}
	if wantDescription != product.Description.String() {
		t.Errorf("want description %s, got %s", wantDescription, product.Description.String())
	}
	if !wantPrice.Equal(product.Price.Value()) {
		t.Errorf("want price %s, got %s", wantPrice, product.Price.Value().String())
	}
}

func TestParseProduct(t *testing.T) {
	tests := []struct {
		name             string
		productName      string
		description      string
		price            decimal.Decimal
		wantErr          bool
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:        "success",
			productName: "Valid product name",
			description: "Valid product description",
			price:       decimal.NewFromFloat(10.10),
		},
		{
			name:          "short description and name",
			productName:   "Na",
			description:   "Desc",
			price:         decimal.NewFromFloat(10.10),
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooShort,
				domain.ErrorCodeProductDescriptionTooShort,
			},
		},
		{
			name:          "long name and invalid price",
			productName:   "ProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductName",
			description:   "Valid product description",
			price:         decimal.NewFromFloat(-10.10),
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooLong,
				domain.ErrorCodeProductPriceLessThanZero,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseProduct(uuid.New(), tt.productName, tt.description, tt.price, uuid.New(), "image/path")
			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			checkProduct(t, parsed, tt.productName, tt.description, tt.price)
		})
	}
}

func checkAddProductRequest(
	t *testing.T,
	product *menu.AddProductRequest,
	wantName,
	wantDescription string,
	wantPrice decimal.Decimal,
	wantImageType string,
) {
	t.Helper()

	if wantName != product.Name.String() {
		t.Errorf("want name %s, got %s", wantName, product.Name.String())
	}
	if wantDescription != product.Description.String() {
		t.Errorf("want description %s, got %s", wantDescription, product.Description.String())
	}
	if wantPrice != product.Price.Value() {
		t.Errorf("want price %s, got %s", wantPrice, product.Price.Value().String())
	}
	if wantImageType != product.ImageType.String() {
		t.Errorf("want image type %s, got %s", wantImageType, product.ImageType.String())
	}
}

func TestParseAddProductRequest(t *testing.T) {
	tests := []struct {
		name             string
		productName      string
		description      string
		price            decimal.Decimal
		imageType        string
		wantErr          bool
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:        "success",
			productName: "Valid product name",
			description: "Valid product description",
			price:       decimal.NewFromFloat(10.10),
			imageType:   "png",
		},
		{
			name:          "short description and name",
			productName:   "Na",
			description:   "Desc",
			price:         decimal.NewFromFloat(10.10),
			imageType:     "png",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooShort,
				domain.ErrorCodeProductDescriptionTooShort,
			},
		},
		{
			name:          "long name and invalid price",
			productName:   "ProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductNameProductName",
			description:   "Valid product description",
			price:         decimal.NewFromFloat(-10.10),
			imageType:     "png",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooLong,
				domain.ErrorCodeProductPriceLessThanZero,
			},
		},
		{
			name:          "invalid image type",
			productName:   "Valid product name",
			description:   "Valid product description",
			price:         decimal.NewFromFloat(10.10),
			imageType:     "invalid",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseAddProductRequest(
				tt.productName,
				tt.description,
				tt.price,
				uuid.New(),
				strings.NewReader("image data"),
				tt.imageType,
			)

			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			checkAddProductRequest(t, parsed, tt.productName, tt.description, tt.price, tt.imageType)
		})
	}
}

func TestParseUpdateProductRequest(t *testing.T) {
	tests := []struct {
		name             string
		productName      *string
		description      *string
		price            *decimal.Decimal
		categoryId       *uuid.UUID
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:        "success",
			productName: new("valid product name"),
			description: new("valid description"),
			price:       new(decimal.NewFromFloat(10.10)),
			categoryId:  new(uuid.New()),
		},
		{
			name:          "no data",
			productName:   nil,
			description:   nil,
			price:         nil,
			categoryId:    nil,
			wantErr:       true,
			wantErrorKind: domain.ErrorKindBadRequest,
			wantErrorCode: domain.ErrorCodeNoData,
		},
		{
			name:          "invalid product name",
			productName:   new("na"),
			description:   nil,
			price:         nil,
			categoryId:    nil,
			wantErr:       true,
			wantErrorKind: domain.ErrorKindValidation,
			wantErrorCode: domain.ErrorCodeInvalidProductUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := menu.ParseUpdateProductRequest(uuid.New(), tt.productName, tt.description, tt.price, tt.categoryId)
			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
		})
	}
}
