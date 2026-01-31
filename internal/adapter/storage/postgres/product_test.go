package postgres_test

import (
	"errors"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
)

// newProduct is a helper function for creating a valid menu.Product.
func newProduct(
	t *testing.T,
	name, description string,
	categoryId uuid.UUID,
	price decimal.Decimal,
	imagePath string,
) *menu.Product {
	t.Helper()

	product, err := menu.NewProduct(uuid.New(), name, description, categoryId, price, imagePath)
	if err != nil {
		t.Fatalf("error creating product: %v", err)
	}
	return product
}

func TestProductRepositoryAddProduct(t *testing.T) {
	tests := []struct {
		name     string
		product  *menu.Product
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			product: newProduct(
				t,
				"Chicken Breasts",
				"Grilled chicken breasts with house sauce.",
				parseUUID(t, "11111111-1111-1111-1111-111111111111"),
				decimal.NewFromInt(10),
				"path",
			),
			checkErr: assertNoErr,
		},
		{
			name: "name conflict",
			product: newProduct(
				t,
				"Coca Cola",
				"Cold bottle of Coca Cola",
				parseUUID(t, "11111111-1111-1111-1111-111111111111"),
				decimal.NewFromInt(10),
				"path",
			),
			checkErr: assertConflictErr,
		},
		{
			name: "category not found",
			product: newProduct(
				t,
				"Chicken Breasts",
				"Grilled chicken breasts with house sauce.",
				uuid.New(),
				decimal.NewFromInt(10),
				"path",
			),
			checkErr: assertNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewProductRepository(database)
			err := repository.AddProduct(context.Background(), tt.product)
			tt.checkErr(t, err)
		})
	}
}

func newProductUpdate(
	t *testing.T,
	id uuid.UUID,
	newName, newDescription *string,
	newCategory *uuid.UUID,
	newPrice *decimal.Decimal,
) *menu.ProductUpdate {
	t.Helper()

	product, err := menu.NewProductUpdate(id, newName, newDescription, newCategory, newPrice)
	if err != nil {
		t.Fatalf("error creating product: %v", err)
	}
	return product
}

func TestProductRepositoryUpdateProduct(t *testing.T) {
	tests := []struct {
		name     string
		update   *menu.ProductUpdate
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			update: newProductUpdate(
				t,
				parseUUID(t, "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				asPointer("New name"),
				nil,
				nil,
				nil,
			),
			checkErr: assertNoErr,
		}, {
			name: "name conflict",
			update: newProductUpdate(
				t,
				parseUUID(t, "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				asPointer("Garlic Bread"),
				nil,
				nil,
				nil,
			),
			checkErr: assertConflictErr,
		}, {
			name: "category not found",
			update: newProductUpdate(
				t,
				parseUUID(t, "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				asPointer("New name"),
				nil,
				asPointer(uuid.New()),
				nil,
			),
			checkErr: func(t *testing.T, err error) {
				t.Helper()

				var domainErr *domain.Error
				if !errors.As(err, &domainErr) {
					t.Fatalf("expected domain.Error, got %v", err)
				}

				if domainErr.Type != domain.NotFound {
					t.Fatalf("expected not found, got %v", err)
				}

				if !strings.Contains(domainErr.Message, "category") {
					t.Fatalf("expected not found for category, got %v", err)
				}
			},
		},
		{
			name: "product not found",
			update: newProductUpdate(
				t,
				uuid.New(),
				asPointer("New name"),
				nil,
				nil,
				nil,
			),
			checkErr: func(t *testing.T, err error) {
				t.Helper()

				var domainErr *domain.Error
				if !errors.As(err, &domainErr) {
					t.Fatalf("expected domain.Error, got %v", err)
				}

				if domainErr.Type != domain.NotFound {
					t.Fatalf("expected not found, got %v", err)
				}

				if !strings.Contains(domainErr.Message, "product") {
					t.Fatalf("expected not found for category, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewProductRepository(database)
			err := repository.UpdateProduct(context.Background(), tt.update)
			tt.checkErr(t, err)
		})
	}
}

func TestProductRepositoryUpdateProductImagePath(t *testing.T) {
	tests := []struct {
		name      string
		id        uuid.UUID
		imagePath string
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:      "success",
			id:        parseUUID(t, "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
			imagePath: "/new/path",
			checkErr:  assertNoErr,
		},
		{
			name:      "not found",
			id:        uuid.New(),
			imagePath: "/new/path",
			checkErr:  assertNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewProductRepository(database)
			err := repository.UpdateProductImagePath(context.Background(), tt.id, tt.imagePath)
			tt.checkErr(t, err)
		})
	}
}

func TestProductRepositoryGetProducts(t *testing.T) {
	truncateMenuTables(t)
	seedMenuTables(t)

	repository := postgres.NewProductRepository(database)
	products, err := repository.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("error getting products: %v", err)
	}

	productNames := []string{
		"Bruschetta", "Garlic Bread", "Fried Calamari", "Stuffed Mushrooms", "Chicken Wings",
		"Grilled Salmon", "Beef Steak", "Chicken Alfredo", "Margherita Pizza", "BBQ Ribs",
		"Cheesecake", "Chocolate Cake", "Tiramisu", "Ice Cream Sundae", "Apple Pie",
		"Coca Cola", "Orange Juice", "Iced Tea", "Latte", "Mineral Water",
		"French Fries", "Side Salad", "Rice Bowl", "Mashed Potatoes", "Onion Rings",
	}

	slices.Sort(productNames)

	returnedNames := make([]string, 0, len(products))
	for _, product := range products {
		returnedNames = append(returnedNames, product.Name.String())
	}
	slices.Sort(returnedNames)

	if len(returnedNames) != len(productNames) {
		t.Errorf("expected %d products, got %d", len(productNames), len(returnedNames))
	}

	for i := range returnedNames {
		if returnedNames[i] != productNames[i] {
			t.Errorf("expected %s, got %s", productNames[i], returnedNames[i])
		}
	}
}

func TestProductRepositoryGetProductImagePaths(t *testing.T) {
	truncateMenuTables(t)
	seedMenuTables(t)

	repository := postgres.NewProductRepository(database)
	result, err := repository.GetProductImagePaths(context.Background())
	if err != nil {
		t.Fatalf("error getting product image paths: %v", err)
	}

	expectedPaths := make(map[string]struct{}, 25)
	for i := 1; i <= 25; i++ {
		expectedPaths["image"+strconv.Itoa(i)] = struct{}{}
	}

	if len(result) != len(expectedPaths) {
		t.Fatalf("expected %d paths, got %d", len(expectedPaths), len(result))
	}

	for path := range expectedPaths {
		if _, ok := result[path]; !ok {
			t.Errorf("expected path %s not found", path)
		}
	}
}

func TestProductRepositoryDeleteProduct(t *testing.T) {
	tests := []struct {
		name      string
		id        uuid.UUID
		checkErr  func(t *testing.T, err error)
		checkPath func(t *testing.T, path string)
	}{
		{
			name:     "success",
			id:       parseUUID(t, "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
			checkErr: assertNoErr,
			checkPath: func(t *testing.T, path string) {
				t.Helper()
			},
		}, {
			name:     "product not found",
			id:       uuid.New(),
			checkErr: assertNotFoundErr,
			checkPath: func(t *testing.T, path string) {
				if path != "image1" {
					t.Fatalf("expected path to be image1, got %v", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewProductRepository(database)
			_, err := repository.DeleteProduct(context.Background(), tt.id)
			tt.checkErr(t, err)
		})
	}
}
