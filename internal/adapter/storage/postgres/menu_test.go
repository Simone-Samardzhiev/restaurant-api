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

// parseUUID is a helper function for parsing UUIDs.
func parseUUID(t *testing.T, s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("error parsing UUID: %v", err)
	}
	return id
}

// asPointer is a helper function for returning values as a pointer.
func asPointer[T any](val T) *T {
	return &val
}

func seedMenuTables(t *testing.T) {
	_, err := database.Exec(productsSeedQuery)
	if err != nil {
		t.Fatalf("error seeding menu tables: %v", err)
	}
}

// truncateMenuTables is a helper cleanup function for truncating tables.
func truncateMenuTables(t *testing.T) {
	if _, err := database.Exec(`TRUNCATE TABLE product_categories, products RESTART IDENTITY CASCADE;`); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// assertNoErr is a helper function for asserting no error is returned.
func assertNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// assertConflictErr is a helper function for asserting an error is domain.Error
// and the type is domain.Conflict.
func assertConflictErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if domainErr.Type != domain.Conflict {
		t.Fatalf("expected conflict, got %v", domainErr.Type)
	}
}

// assertBadRequestErr is a helper function for asserting an error is domain.Error
// and the type is domain.BadRequest.
func assertBadRequestErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if domainErr.Type != domain.BadRequest {
		t.Fatalf("expected bad request, got %v", domainErr.Type)
	}
}

// assertNotFoundErr is a helper function for asserting an error is domain.Error
// and the type is domain.NotFound.
func assertNotFoundErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}
	if domainErr.Type != domain.NotFound {
		t.Fatalf("expected NotFound, got %v", domainErr.Type)
	}
}

// newCategory is a helper function for creating valid menu.Category.
func newCategory(t *testing.T, name string) *menu.Category {
	t.Helper()
	category, err := menu.NewCategory(uuid.New(), name)
	if err != nil {
		t.Fatalf("error creating category: %v", err)
	}
	return category
}

func TestMenuRepositoryAddCategory(t *testing.T) {
	tests := []struct {
		name     string
		category *menu.Category
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "success",
			category: newCategory(t, "newName"),
			checkErr: assertNoErr,
		},
		{
			name:     "duplicate category name",
			category: newCategory(t, "Appetizers"),
			checkErr: assertConflictErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewMenuRepository(database)
			err := repository.AddCategory(context.Background(), tt.category)
			tt.checkErr(t, err)
		})
	}
}

// newCategoryUpdate is a helper function for creating a valid menu.CategoryUpdate.
func newCategoryUpdate(t *testing.T, id uuid.UUID, newName *string) *menu.CategoryUpdate {
	update, err := menu.NewCategoryUpdate(id, newName)
	if err != nil {
		t.Fatalf("error creating category update: %v", err)
	}

	return update
}

func TestMenuRepositoryUpdateCategory(t *testing.T) {
	tests := []struct {
		name     string
		update   *menu.CategoryUpdate
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "success",
			update:   newCategoryUpdate(t, parseUUID(t, "11111111-1111-1111-1111-111111111111"), asPointer("New name")),
			checkErr: assertNoErr,
		},
		{
			name:     "duplicate category name",
			update:   newCategoryUpdate(t, parseUUID(t, "11111111-1111-1111-1111-111111111111"), asPointer("Drinks")),
			checkErr: assertConflictErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewMenuRepository(database)
			err := repository.UpdateCategory(context.Background(), tt.update)
			tt.checkErr(t, err)
		})
	}
}

func TestMenuRepositoryDeleteCategory(t *testing.T) {
	tests := []struct {
		name     string
		id       uuid.UUID
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "success",
			id:       parseUUID(t, "66666666-6666-6666-6666-666666666666"),
			checkErr: assertNoErr,
		},
		{
			name:     "delete used category",
			id:       parseUUID(t, "11111111-1111-1111-1111-111111111111"),
			checkErr: assertBadRequestErr,
		},
		{
			name:     "category not found",
			id:       uuid.New(),
			checkErr: assertNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewMenuRepository(database)
			err := repository.DeleteCategory(context.Background(), tt.id)
			tt.checkErr(t, err)
		})
	}
}

func TestMenuRepositoryGetCategories(t *testing.T) {
	tests := []struct {
		name        string
		filter      *menu.CategoryFilter
		checkErr    func(t *testing.T, err error)
		checkResult func(t *testing.T, categories []menu.Category)
	}{
		{
			name:     "success all",
			filter:   menu.NewCategoryFilter(nil),
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, categories []menu.Category) {
				t.Helper()

				expectedNames := []string{"Appetizers", "Main Dishes", "Desserts", "Drinks", "Sides", "Salads"}
				if len(categories) != len(expectedNames) {
					t.Fatalf("expected %d categories, got %d", len(expectedNames), len(categories))
				}

				returnedNames := make([]string, 0, len(categories))
				for _, category := range categories {
					returnedNames = append(returnedNames, category.Name.String())
				}

				for i := range returnedNames {
					if expectedNames[i] != categories[i].Name.String() {
						t.Fatalf("expected category name %s, got %s", expectedNames[i], categories[i].Name.String())
					}
				}
			},
		},
		{
			name:     "success filter with id",
			filter:   menu.NewCategoryFilter(asPointer(parseUUID(t, "11111111-1111-1111-1111-111111111111"))),
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, categories []menu.Category) {
				t.Helper()

				if len(categories) != 1 {
					t.Fatalf("expected 1 category, got %d", len(categories))
				}
				if categories[0].Name.String() != "Appetizers" {
					t.Fatalf("expected category name %s, got %s", "Appetizers", categories[0].Name.String())
				}
			},
		},
		{
			name:     "success filter with non existing id",
			filter:   menu.NewCategoryFilter(asPointer(uuid.New())),
			checkErr: assertNoErr,
			checkResult: func(t *testing.T, categories []menu.Category) {
				t.Helper()
				if len(categories) != 0 {
					t.Fatalf("expected 0 categories, got %d", len(categories))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateMenuTables(t)
			seedMenuTables(t)

			repository := postgres.NewMenuRepository(database)
			result, err := repository.GetCategories(context.Background(), tt.filter)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

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

func TestMenuRepositoryAddProduct(t *testing.T) {
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

			repository := postgres.NewMenuRepository(database)
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

func TestMenuRepositoryUpdateProduct(t *testing.T) {
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

			repository := postgres.NewMenuRepository(database)
			err := repository.UpdateProduct(context.Background(), tt.update)
			tt.checkErr(t, err)
		})
	}
}

func TestMenuRepositoryUpdateProductImagePath(t *testing.T) {
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

			repository := postgres.NewMenuRepository(database)
			err := repository.UpdateProductImagePath(context.Background(), tt.id, tt.imagePath)
			tt.checkErr(t, err)
		})
	}
}

func TestMenuRepositoryGetProducts(t *testing.T) {
	truncateMenuTables(t)
	seedMenuTables(t)

	repository := postgres.NewMenuRepository(database)
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

func TestMenuRepositoryGetProductImagePaths(t *testing.T) {
	truncateMenuTables(t)
	seedMenuTables(t)

	repository := postgres.NewMenuRepository(database)
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

func TestMenuRepositoryDeleteProduct(t *testing.T) {
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

			repository := postgres.NewMenuRepository(database)
			_, err := repository.DeleteProduct(context.Background(), tt.id)
			tt.checkErr(t, err)
		})
	}
}
