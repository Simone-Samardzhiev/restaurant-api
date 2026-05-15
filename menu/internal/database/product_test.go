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

		var fetchedProduct domain.Product
		row := testDb.QueryRow(`SELECT id, name, description, price, category_id, image_key, status, created_at, updated_at FROM products WHERE id = $1`, product.Id)
		if err := row.Scan(&fetchedProduct.Id, &fetchedProduct.Name, &fetchedProduct.Description, &fetchedProduct.Price, &fetchedProduct.CategoryId, &fetchedProduct.ImageKey, &fetchedProduct.Status, &fetchedProduct.CreatedAt, &fetchedProduct.UpdatedAt); err != nil {
			t.Fatalf("Error scanning product: %v", err)
		}

		if fetchedProduct.Name != product.Name {
			t.Errorf("Product name mismatch: got %s, want %s", fetchedProduct.Name, product.Name)
		}
		if fetchedProduct.Description != product.Description {
			t.Errorf("Product description mismatch: got %s, want %s", fetchedProduct.Description, product.Description)
		}
		if !fetchedProduct.Price.Equal(fetchedProduct.Price) {
			t.Errorf("Product price mismatch: got %v, want %v", fetchedProduct.Price, product.Price)
		}
		if fetchedProduct.CategoryId != product.CategoryId {
			t.Errorf("Product category id mismatch: got %v, want %v", fetchedProduct.CategoryId, product.CategoryId)
		}
		if fetchedProduct.ImageKey != product.ImageKey {
			t.Errorf("Product image key mismatch: got %v, want %v", fetchedProduct.ImageKey, product.ImageKey)
		}
		if fetchedProduct.Status != product.Status {
			t.Errorf("Product status mismatch: got %v, want %v", fetchedProduct.Status, product.Status)
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
