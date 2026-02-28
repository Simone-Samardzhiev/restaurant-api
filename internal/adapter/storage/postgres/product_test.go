package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/testutils"
	"testing"

	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// mustParseSaveProductRequest parses all fields to create a valid [menu.SaveProductRequest]
// and panics if any error occurred.
func mustParseSaveProductRequest(
	name,
	description string,
	price decimal.Decimal,
	categoryId uuid.UUID,
	imagePath string,
) *menu.SaveProductRequest {
	parsedName, err := menu.ParseProductName(name)
	if err != nil {
		panic(err)
	}

	parsedDescription, err := menu.ParseProductDescription(description)
	if err != nil {
		panic(err)
	}

	parsedPrice, err := menu.ParseProductPrice(price)
	if err != nil {
		panic(err)
	}

	return &menu.SaveProductRequest{
		Name:        parsedName,
		Description: parsedDescription,
		Price:       parsedPrice,
		CategoryId:  categoryId,
		ImagePath:   imagePath,
	}
}

func checkSaveProductResult(t *testing.T,
	product *menu.Product,
	wantName menu.ProductName,
	wantDescription menu.ProductDescription,
	wantPrice menu.ProductPrice,
	wantCategoryId uuid.UUID,
	wantImagePath string,
) {
	t.Helper()

	if wantName.String() != product.Name.String() {
		t.Errorf("want name %s, got %s", wantName, product.Name)
	}
	if wantDescription.String() != product.Description.String() {
		t.Errorf("want description %s, got %s", wantDescription, product.Description)
	}
	if !wantPrice.Value().Equal(product.Price.Value()) {
		t.Errorf("want price %s, got %s", wantPrice.Value(), product.Price.Value())
	}
	if wantCategoryId != product.CategoryId {
		t.Errorf("want category id %s, got %s", wantCategoryId, product.CategoryId)
	}
	if wantImagePath != product.ImagePath {
		t.Errorf("want image path %s, got %s", wantImagePath, product.ImagePath)
	}
}

func TestProductRepositorySaveProduct(t *testing.T) {
	tests := []struct {
		name             string
		request          *menu.SaveProductRequest
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			request: mustParseSaveProductRequest(
				"New product",
				"New product description",
				decimal.NewFromFloat(10),
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				"image/path",
			),
		},
		{
			name: "name already exists",
			request: mustParseSaveProductRequest(
				"Bruschetta",
				"New product description",
				decimal.NewFromFloat(10),
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				"image/path",
			),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindConflict,
			wantErrorCode: domain.ErrorCodeProductNameConflict,
		},
		{
			name: "category not found",
			request: mustParseSaveProductRequest(
				"New product",
				"New product description",
				decimal.NewFromFloat(10),
				uuid.New(),
				"image/path",
			),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeCategoryNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewProductRepository(database)
			result, err := repo.SaveProduct(context.Background(), test.request)
			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got : %v", err)
			}
			checkSaveProductResult(
				t,
				result,
				test.request.Name,
				test.request.Description,
				test.request.Price,
				test.request.CategoryId,
				test.request.ImagePath,
			)
		})
	}
}

func TestProductRepositoryUpdateProduct(t *testing.T) {
	tests := []struct {
		name             string
		request          *menu.UpdateProductRequest
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			request: menu.MustParseProductUpdateRequest(
				uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				new("New product name"),
				nil, nil, nil,
			),
		},
		{
			name: "name already exists",
			request: menu.MustParseProductUpdateRequest(
				uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				new("Garlic Bread"),
				nil, nil, nil,
			),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindConflict,
			wantErrorCode: domain.ErrorCodeProductNameConflict,
		},
		{
			name: "category not found",
			request: menu.MustParseProductUpdateRequest(
				uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				nil, nil, nil,
				new(uuid.New()),
			),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeCategoryNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeCategoryNotFoundByID},
		},
		{
			name: "product not found",
			request: menu.MustParseProductUpdateRequest(
				uuid.New(),
				new("New product name"),
				nil, nil, nil,
			),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeProductNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeProductNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)
			repo := postgres.NewProductRepository(database)
			err := repo.UpdateProduct(context.Background(), test.request)
			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}
		})
	}
}

func TestProductRepositoryUpdateImagePath(t *testing.T) {
	tests := []struct {
		name             string
		id               uuid.UUID
		path             string
		wantPath         string
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:     "success",
			id:       uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
			path:     "new/path/to/image",
			wantPath: "image1",
		},
		{
			name:             "product not found",
			id:               uuid.New(),
			path:             "new/path/to/image",
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeProductNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeProductNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewProductRepository(database)
			result, err := repo.UpdateImagePath(context.Background(), test.id, test.path)
			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}
			if test.wantPath != result {
				t.Fatalf("want %v, got %v", test.wantPath, result)
			}
		})
	}
}

func TestProductRepositoryDeleteProduct(t *testing.T) {
	tests := []struct {
		name             string
		id               uuid.UUID
		wantPath         string
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:     "success",
			id:       uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
			wantPath: "image1",
		},
		{
			name:             "product not found",
			id:               uuid.New(),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeProductNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeProductNotFoundByID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seedMenuTables(t)

			repo := postgres.NewProductRepository(database)
			result, err := repo.DeleteProduct(context.Background(), test.id)
			if test.wantErr {
				testutils.AssertError(t, err, test.wantErrorKind, test.wantErrorCode, test.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			if test.wantPath != result {
				t.Fatalf("want %v, got %v", test.wantPath, result)
			}
		})
	}
}
