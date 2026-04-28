package db

import (
	"errors"
	"menu/internal/domain"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

func TestCategoryRepositorySave(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repository := NewCategoryRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.ExecContext(context.Background(), `TRUNCATE TABLE categories CASCADE`); err != nil {
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
	})

	t.Run("name conflict", func(t *testing.T) {
		if _, err := testDb.ExecContext(context.Background(), `TRUNCATE TABLE categories CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		if _, err := testDb.ExecContext(
			context.Background(),
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
		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNameConflict {
				t.Fatalf("Want error code: %s, got: %s", domain.ErrorCodeCategoryNameConflict, domainErr.Code)
			}

			return
		}

		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}

func TestCategoryRepositoryGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := testDb.ExecContext(context.Background(), `TRUNCATE TABLE categories CASCADE`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	if _, err := testDb.ExecContext(
		context.Background(),
		`INSERT INTO categories(id, name, created_at, updated_at) 
		VALUES (gen_random_uuid(), 'Category 1', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 2', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 3', NOW(), NOW()),
		       (gen_random_uuid(), 'Category 4', NOW(), NOW())`,
	); err != nil {
		t.Fatalf("Error seeding data: %v", err)
	}

	repository := NewCategoryRepository(testDb)
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
