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
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       uuid.New(),
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if _, err := testDb.Exec(`INSERT INTO categories(id, name, created_at, updated_at) VALUES ($1, 'Test', NOW(), NOW())`, product.CategoryId); err != nil {
			t.Fatalf("Error inserting category: %v", err)
		}
		if err := repository.Save(context.Background(), &product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		var fetchedProduct domain.Product
		row := testDb.QueryRow(`SELECT id, name, description, price, category_id, image_key, image_content_type, status, created_at, updated_at FROM products WHERE id = $1`, product.Id)
		if err := row.Scan(&fetchedProduct.Id, &fetchedProduct.Name, &fetchedProduct.Description, &fetchedProduct.Price, &fetchedProduct.CategoryId, &fetchedProduct.ImageKey, &fetchedProduct.ImageContentType, &fetchedProduct.Status, &fetchedProduct.CreatedAt, &fetchedProduct.UpdatedAt); err != nil {
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
		if fetchedProduct.ImageContentType != domain.ImageContentTypePNG {
			t.Errorf("Product image content type mismatch: got %v, want %v", fetchedProduct.ImageContentType, domain.ImageContentTypePNG)
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
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       categoryId,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key,image_content_type, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, 'Some test description for product', 10, $2, 'testImageKey', 'image/png', 'ready', NOW(), NOW())`,
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
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       uuid.New(),
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
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

func TestPostgresProductRepositoryGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		product := domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       uuid.New(),
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if _, err := testDb.Exec(
			`INSERT INTO categories(id, name, created_at, updated_at) 
			VALUES ($1, 'test', NOW(), NOW())`,
			product.CategoryId,
		); err != nil {
			t.Fatalf("Error inserting category: %v", err)
		}

		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key, image_content_type, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			product.Id,
			product.Name,
			product.Description,
			product.Price,
			product.CategoryId,
			product.ImageKey,
			product.ImageContentType,
			product.Status,
			product.CreatedAt,
			product.UpdatedAt,
		); err != nil {
			t.Fatalf("Error inserting product: %v", err)
		}

		fetchedProduct, err := repository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Name != product.Name {
			t.Errorf("Want product name %s, got %s", product.Name, fetchedProduct.Name)
		}
		if fetchedProduct.Description != product.Description {
			t.Errorf("Want product description %s, got %s", product.Description, fetchedProduct.Description)
		}
		if !fetchedProduct.Price.Equal(fetchedProduct.Price) {
			t.Errorf("Want product price %s, got %s", product.Price, fetchedProduct.Price)
		}
		if fetchedProduct.CategoryId != product.CategoryId {
			t.Errorf("Want product category id %s, got %s", product.CategoryId, fetchedProduct.CategoryId)
		}
		if fetchedProduct.ImageKey != product.ImageKey {
			t.Errorf("Want product image key %s, got %s", product.ImageKey, fetchedProduct.ImageKey)
		}
		if fetchedProduct.ImageContentType != product.ImageContentType {
			t.Errorf("Want product image content type %s, got %s", product.ImageContentType, fetchedProduct.ImageContentType)
		}
		if fetchedProduct.Status != domain.ProductStatusReady {
			t.Errorf("Want product status %s, got %s", product.Status, fetchedProduct.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repository.Get(context.Background(), uuid.New())
		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}
		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeProductNotFound {
				t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNotFound, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}

func TestPostgresProductRepositoryDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	repository := NewPostgresProductRepository(testDb)
	categoryId := uuid.New()

	if _, err := testDb.Exec(`TRUNCATE TABLE categories CASCADE`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	if _, err := testDb.Exec(
		`INSERT INTO categories(id, name, created_at, updated_at) 
		VALUES ($1, 'test', NOW(), NOW())`,
		categoryId,
	); err != nil {
		t.Fatalf("Error inserting category: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		productId := uuid.New()
		if _, err := testDb.Exec(
			`INSERT INTO products(id, name, description, price, category_id, image_key, image_content_type, status, created_at, updated_at) 
			VALUES ($1, 'Some test name', 'Some test description for product', 10, $2, 'testImageKey', 'image/png', 'ready', NOW(), NOW())`,
			productId,
			categoryId,
		); err != nil {
			t.Fatalf("Error inserting product: %v", err)
		}

		if err := repository.Delete(context.Background(), productId); err != nil {
			t.Fatalf("Error deleting product: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := repository.Delete(context.Background(), uuid.New())
		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}
		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeProductNotFound {
				t.Fatalf("Want error code %s, got %s", domain.ErrorCodeProductNotFound, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type: domain.Error, got: %T", err)
	})
}
