package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain/menu"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

// newCategory is a helper function for creating valid menu.Category.
func newCategory(t *testing.T, name string) *menu.Category {
	t.Helper()
	category, err := menu.NewCategory(uuid.New(), name)
	if err != nil {
		t.Fatalf("error creating category: %v", err)
	}
	return category
}

func TestCategoryRepositoryAddCategory(t *testing.T) {
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

			repository := postgres.NewCategoryRepository(database)
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

func TestCategoryRepositoryUpdateCategory(t *testing.T) {
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

			repository := postgres.NewCategoryRepository(database)
			err := repository.UpdateCategory(context.Background(), tt.update)
			tt.checkErr(t, err)
		})
	}
}

func TestCategoryRepositoryDeleteCategory(t *testing.T) {
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

			repository := postgres.NewCategoryRepository(database)
			err := repository.DeleteCategory(context.Background(), tt.id)
			tt.checkErr(t, err)
		})
	}
}

func TestCategoryRepositoryGetCategories(t *testing.T) {
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

			repository := postgres.NewCategoryRepository(database)
			result, err := repository.GetCategories(context.Background(), tt.filter)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}
