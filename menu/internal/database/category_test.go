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

		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES (gen_random_uuid(), 'Conflicting name', NOW(), NOW())`,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Conflicting name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

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

	if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	if _, err := testDb.Exec(
		`INSERT INTO categories(id, name, created_at, updated_at) 
		VALUES (gen_random_uuid(), 'Category 1', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 2', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 3', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 4', NOW(), NOW())`,
	); err != nil {
		t.Fatalf("Error seeding data: %v", err)
	}

	repository := NewPostgresCategoryRepository(testDb)
	categories, err := repository.GetAll(context.Background())
	if err != nil {
		t.Fatalf("Error getting all categories: %v", err)
	}

	wantNames := []string{"Category 1", "Category 2", "Category 3", "Category 4"}

	if len(categories) != len(wantNames) {
		t.Fatalf("Unexpected category length: %d", len(categories))
	}

	slices.SortFunc(categories, func(a, b domain.Category) int {
		return strings.Compare(a.Name, b.Name)
	})

	slices.Sort(wantNames)

	for i := 0; i < len(categories); i++ {
		if categories[i].Name != wantNames[i] {
			t.Fatalf("Unexpected category name: %s", categories[i].Name)
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

		id := uuid.New()

		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at)
			VALUES ($1, 'test', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		const newName = "New Name"

		if err := repository.Update(context.Background(), id, newName); err != nil {
			t.Logf("Cause: %v", errors.Unwrap(err))
			t.Fatalf("Error updating category: %v", err)
		}

		row := testDb.QueryRowContext(context.Background(), `SELECT name FROM categories WHERE id = $1`, id)
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

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'Test1', NOW(), NOW()),
			(gen_random_uuid(), 'Test2', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		const newName = "Test2"
		err := repository.Update(context.Background(), id, newName)
		if err == nil {
			t.Fatalf("Want conflict error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNameConflict {
				t.Fatalf("Want error code: %s, got: %s", domainErr.Code, domainErr.Code)
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
				t.Fatalf("Want error code: %s, got: %s", domainErr.Code, domainErr.Code)
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

	repository := NewPostgresCategoryRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at)
			VALUES ($1, 'test', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error seeding data: %v", err)
		}

		if err := repository.Delete(context.Background(), id); err != nil {
			t.Fatalf("Error deleting category: %v", err)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		id := uuid.New()
		err := repository.Delete(context.Background(), id)

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

		id := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'test', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error inserting category: %v", err)
		}

		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key,image_content_type, status, created_at, updated_at)
			VALUES (gen_random_uuid(), 'name', 'Some test description for product', 10, $1, 'testImageKey', 'image/png', 'ready', NOW(), NOW())`,
			id,
		); err != nil {
			t.Fatalf("Error inserting product: %v", err)
		}

		err := repository.Delete(context.Background(), id)
		if err == nil {
			t.Fatalf("Want category has products error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryHasProducts {
				t.Fatalf("Want error code: %s, got: %s", domainErr.Code, domainErr.Code)
			}

			return
		}

		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}
