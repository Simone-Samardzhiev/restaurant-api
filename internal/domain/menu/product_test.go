package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/testutils"
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseProductName(test.productName)
			if test.wantErr {
				testutils.AssertErrorDetail(t, err, test.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}

			if test.productName != parsed.String() {
				t.Fatalf("want name %s, got %s", test.productName, parsed.String())
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := menu.ParseProductDescription(test.description)
			if test.wantErr {
				testutils.AssertErrorDetail(t, err, test.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			if test.description != parsed.String() {
				t.Fatalf("want description %s, got %s", test.description, parsed.String())
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := menu.ParseProductPrice(test.price)
			if test.wantErr {
				testutils.AssertErrorDetail(t, err, test.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			if !test.price.Equal(parsed.Value()) {
				t.Fatalf("want price %s, got %s", test.price, parsed.Value().String())
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseProduct(uuid.New(), test.productName, test.description, test.price, uuid.New(), "image/path")
			if test.wantErr {
				testutils.AssertError(t, err, domain.ErrorKindValidation, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			checkProduct(t, parsed, test.productName, test.description, test.price)
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseAddProductRequest(
				test.productName,
				test.description,
				test.price,
				uuid.New(),
				strings.NewReader("image data"),
				test.imageType,
			)

			if test.wantErr {
				testutils.AssertError(t, err, domain.ErrorKindValidation, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
			checkAddProductRequest(t, parsed, test.productName, test.description, test.price, test.imageType)
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
			wantErrorCode: domain.ErrorCodeNoData,
		},
		{
			name:          "invalid product name",
			productName:   new("na"),
			description:   nil,
			price:         nil,
			categoryId:    nil,
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidProductUpdate,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := menu.ParseUpdateProductRequest(uuid.New(), test.productName, test.description, test.price, test.categoryId, new("new image path"))
			if test.wantErr {
				testutils.AssertError(t, err, domain.ErrorKindValidation, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %d", err)
			}
		})
	}
}
