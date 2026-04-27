package db

import (
	"errors"
	"menu/internal/domain"
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
