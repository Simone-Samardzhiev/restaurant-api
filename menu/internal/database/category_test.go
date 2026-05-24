package database

import (
	"errors"
	"menu/internal/domain"
	"slices"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPostgresCategoryRepositorySave(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	repository := NewPostgresCategoryRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New category",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := repository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		row := testDb.QueryRow(`SELECT name FROM categories WHERE id = $1`, category.Id)
		var name string
		if err := row.Scan(&name); err != nil {
			t.Fatalf("Error getting category name: %v", err)
		}
		if name != category.Name {
			t.Fatalf("Category name mismatch: got %s, want %s", name, category.Name)
		}
	})

	t.Run("name conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Conflict name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := repository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		// change the id so there is no primary key conflict
		category.Id = uuid.New()
		err := repository.Save(context.Background(), category)
		if err == nil {
			t.Fatalf("Want conflict error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNameConflict {
				t.Fatalf("Want error code: %s, got: %s", domain.ErrorCodeCategoryNameConflict, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}

func TestPostgresCategoryRepositoryGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	repository := NewPostgresCategoryRepository(testDb)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	names := []string{"Category 1", "Category 2", "Category 3", "Category 4"}

	for _, name := range names {
		if err := repository.Save(context.Background(), &domain.Category{
			Id:        uuid.New(),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}
	}

	categories, err := repository.GetAll(context.Background())
	if err != nil {
		t.Fatalf("Error getting all categories: %v", err)
	}
	slices.SortFunc(categories, func(a, b domain.Category) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.Sort(names)

	if len(names) != len(categories) {
		t.Fatalf("Want %d categories, got %d", len(names), len(categories))
	}

	for i := 0; i < len(names); i++ {
		if categories[i].Name != names[i] {
			t.Fatalf("Want category %s, got %s", names[i], categories[i].Name)
		}
	}
}

func TestPostgresCategoryRepositoryUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	repository := NewPostgresCategoryRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New category",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		const newName = "New Name"
		if err := repository.Update(context.Background(), category.Id, newName); err != nil {
			t.Fatalf("Error updating category: %v", err)
		}

		row := testDb.QueryRowContext(context.Background(), `SELECT name FROM categories WHERE id = $1`, category.Id)
		var name string
		if err := row.Scan(&name); err != nil {
			t.Fatalf("Error getting category name: %v", err)
		}

		if name != newName {
			t.Fatalf("Category name mismatch: want %s, got: %s", newName, name)
		}
	})

	t.Run("name conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category1 := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repository.Save(context.Background(), category1); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		category2 := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repository.Save(context.Background(), category2); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		err := repository.Update(context.Background(), category2.Id, category1.Name)
		if err == nil {
			t.Fatalf("Want conflict error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNameConflict {
				t.Fatalf("Want error code: %s, got: %s", domain.ErrorCodeCategoryNameConflict, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()
		err := repository.Update(context.Background(), id, "Random name")
		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNotFound {
				t.Fatalf("Want error code: %s, got: %s", domain.ErrorCodeCategoryNotFound, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}

func TestPostgresCategoryRepositoryDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New category",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		if err := categoryRepository.Delete(context.Background(), category.Id); err != nil {
			t.Fatalf("Error deleting category: %v", err)
		}

		var exists bool
		row := testDb.QueryRow(`SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)`, category.Id)
		if err := row.Scan(&exists); err != nil {
			t.Fatalf("Error checking if category exists: %v", err)
		}
		if exists {
			t.Fatalf("Category exists after deletion")
		}
	})

	t.Run("category not found", func(t *testing.T) {
		id := uuid.New()
		err := categoryRepository.Delete(context.Background(), id)

		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNotFound {
				t.Fatalf("Want error code: %s, got: %s", domainErr.Code, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})

	t.Run("category has products", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "New category",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		if err := productRepository.Save(context.Background(), &domain.Product{
			Id:               uuid.New(),
			Name:             "Test product name",
			Description:      "Test product description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "image/key/",
			ImageContentType: "image/png",
			Status:           "ready",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		err := categoryRepository.Delete(context.Background(), category.Id)
		if err == nil {
			t.Fatalf("Want category has products error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryHasProducts {
				t.Fatalf("Want error code: %s, got: %s", domain.ErrorCodeCategoryHasProducts, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}
