package database

import (
	"context"
	"errors"
	"menu/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPostgresProductRepositorySave(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		product := domain.Product{
			Id:          uuid.New(),
			Name:        "Test",
			Description: "Some test description for product",
			Price:       decimal.NewFromInt(100),
			CategoryId:  uuid.New(),
			ImageKey:    "imageKey",
			Status:      domain.ProductStatusReady,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if _, err := testDb.Exec(`INSERT INTO categories(id, name, created_at, updated_at) VALUES ($1, 'Test', NOW(), NOW())`, product.CategoryId); err != nil {
			t.Fatalf("Error inserting category: %v", err)
		}
		if err := repository.Save(context.Background(), &product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		categoryId := uuid.New()

		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'Test', NOW(), NOW())`,
			categoryId,
		); err != nil {
			t.Fatalf("Error inserting category: %v", err)
		}

		product := domain.Product{
			Id:          uuid.New(),
			Name:        "Test",
			Description: "Some test description for product",
			Price:       decimal.NewFromInt(100),
			CategoryId:  categoryId,
			ImageKey:    "imageKey",
			Status:      domain.ProductStatusReady,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, 'Some test description for product', 10, $2, 'testImageKey', 'ready', NOW(), NOW())`,
			product.Name,
			product.CategoryId,
		); err != nil {
			t.Fatalf("Error inserting product: %v", err)
		}

		err := repository.Save(context.Background(), &product)
		if err == nil {
			t.Fatalf("Want conflict error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeProductNameConflict {
				t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNameConflict, domainErr.Code)
			}

			return
		}

		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})

	t.Run("category not found", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		product := domain.Product{
			Id:          uuid.New(),
			Name:        "Test",
			Description: "Some test description for product",
			Price:       decimal.NewFromInt(100),
			CategoryId:  uuid.New(),
			ImageKey:    "imageKey",
			Status:      domain.ProductStatusReady,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repository.Save(context.Background(), &product)
		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNotFound {
				t.Fatalf("Want error code %s, got %s", domain.ErrorCodeCategoryNotFound, domainErr.Code)
			}

			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})

}
