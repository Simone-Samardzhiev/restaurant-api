package database

import (
	"context"
	"errors"
	"menu/internal/domain"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPostgresProductRepositorySave(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}
		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Name != product.Name {
			t.Errorf("Want name: %s, got: %s", product.Name, fetchedProduct.Name)
		}
		if fetchedProduct.Description != product.Description {
			t.Errorf("Want description: %s, got: %s", product.Description, fetchedProduct.Description)
		}
		if !fetchedProduct.Price.Equal(fetchedProduct.Price) {
			t.Errorf("Want price: %s, got: %s", product.Price, fetchedProduct.Price)
		}
		if fetchedProduct.CategoryId != product.CategoryId {
			t.Errorf("Want category id: %s, got: %s", product.CategoryId, fetchedProduct.CategoryId)
		}
		if fetchedProduct.ImageKey != product.ImageKey {
			t.Errorf("Want image key: %s, got: %s", product.ImageKey, fetchedProduct.ImageKey)
		}
		if fetchedProduct.ImageContentType != domain.ImageContentTypePNG {
			t.Errorf("Want image content type: %s, got: %s", product.ImageContentType, fetchedProduct.ImageContentType)
		}
		if fetchedProduct.Status != product.Status {
			t.Errorf("Want status: %s, got: %s", product.Status, fetchedProduct.Status)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Conflicting name",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		// change the id so there is no primary key conflict
		product.Id = uuid.New()
		// change the image key so there is no unique conflict
		product.ImageKey = "imageKey2"

		err := productRepository.Save(context.Background(), product)
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

		product := &domain.Product{
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

		err := productRepository.Save(context.Background(), product)
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
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Name != product.Name {
			t.Errorf("Want name: %s, got: %s", product.Name, fetchedProduct.Name)
		}
		if fetchedProduct.Description != product.Description {
			t.Errorf("Want description: %s, got: %s", product.Description, fetchedProduct.Description)
		}
		if !fetchedProduct.Price.Equal(fetchedProduct.Price) {
			t.Errorf("Want price: %s, got: %s", product.Price, fetchedProduct.Price)
		}
		if fetchedProduct.CategoryId != product.CategoryId {
			t.Errorf("Want category id: %s, got: %s", product.CategoryId, fetchedProduct.CategoryId)
		}
		if fetchedProduct.ImageKey != product.ImageKey {
			t.Errorf("Want image key: %s, got: %s", product.ImageKey, fetchedProduct.ImageKey)
		}
		if fetchedProduct.ImageContentType != product.ImageContentType {
			t.Errorf("Want image content type: %s, got: %s", product.ImageContentType, fetchedProduct.ImageContentType)
		}
		if fetchedProduct.Status != product.Status {
			t.Errorf("Want status: %s, got: %s", product.Status, fetchedProduct.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := productRepository.Get(context.Background(), uuid.New())
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

func TestPostgresProductRepositoryGetAllReady(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {

	}

	category := &domain.Category{
		Id:        uuid.New(),
		Name:      "Test name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := categoryRepository.Save(context.Background(), category); err != nil {
		t.Fatalf("Error saving category: %v", err)
	}

	products := []domain.Product{
		{
			Id:               uuid.New(),
			Name:             "Test 1",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test 2",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey2",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test 3",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey3",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	for _, product := range products {
		if err := productRepository.Save(context.Background(), &product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}
	}
	products = slices.DeleteFunc(products, func(product domain.Product) bool {
		return product.Status == domain.ProductStatusMissingImage
	})

	fetchedProducts, err := productRepository.GetAllReady(context.Background())
	if err != nil {
		t.Fatalf("Error fetching products: %v", err)
	}
	if len(fetchedProducts) != len(products) {
		t.Fatalf("Want %d products, got %d", len(products), len(fetchedProducts))
	}

	slices.SortFunc(products, func(a, b domain.Product) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.SortFunc(fetchedProducts, func(a, b domain.Product) int {
		return strings.Compare(a.Name, b.Name)
	})

	for i := 0; i < len(products); i++ {
		if products[i].Name != fetchedProducts[i].Name {
			t.Errorf("Want product: %s, got: %s", products[i].Name, fetchedProducts[i].Name)
		}
	}
}

func TestPostgresProductRepositoryUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		updateReq := &domain.UpdateProductRequest{
			Id:    product.Id,
			Name:  new("New Name"),
			Price: new(decimal.NewFromInt(100)),
		}
		if err := productRepository.Update(context.Background(), updateReq); err != nil {
			t.Fatalf("Error updating product: %v", err)
		}

		fetchedProduct, err := productRepository.Get(context.Background(), product.Id)
		if err != nil {
			t.Fatalf("Error fetching product: %v", err)
		}

		if fetchedProduct.Name != *updateReq.Name {
			t.Errorf("Want product: %s, got: %s", *updateReq.Name, fetchedProduct.Name)
		}
		if !fetchedProduct.Price.Equal(*updateReq.Price) {
			t.Errorf("Want price: %s, got: %s", *updateReq.Price, fetchedProduct.Price)
		}
	})

	t.Run("name conflict", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product1 := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test1",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product1); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		product2 := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test2",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey2",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product2); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		err := productRepository.Update(context.Background(), &domain.UpdateProductRequest{
			Id:   product1.Id,
			Name: new("Test2"),
		})

		if err == nil {
			t.Errorf("Want conflict error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeProductNameConflict {
				t.Errorf("Want error code: %v, got: %v", domainErr.Code, domainErr.Code)
			}
		} else {
			t.Fatalf("Want *domain.Error, got: %T", err)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		err := productRepository.Update(context.Background(), &domain.UpdateProductRequest{
			Id:         product.Id,
			CategoryId: new(uuid.New()),
		})
		if err == nil {
			t.Errorf("Want not found error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeCategoryNotFound {
				t.Errorf("Want error code: %v, got: %v", domainErr.Code, domainErr.Code)
			}
		} else {
			t.Fatalf("Want *domain.Error, got: %T", err)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		err := productRepository.Update(context.Background(), &domain.UpdateProductRequest{
			Id:   uuid.New(),
			Name: new("New Name"),
		})
		if err == nil {
			t.Fatalf("Want not found error, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeProductNotFound {
				t.Errorf("Want error code: %v, got: %v", domainErr.Code, domainErr.Code)
			}
		} else {
			t.Fatalf("Want error type *domain.Error, got: %T", err)
		}
	})
}

func TestPostgresProductRepositoryDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		if err := productRepository.Delete(context.Background(), product.Id); err != nil {
			t.Fatalf("Error deleting product: %v", err)
		}

		var exists bool
		row := testDb.QueryRow(`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`, product.Id)
		if err := row.Scan(&exists); err != nil {
			t.Fatalf("Error checking if product exists: %v", err)
		}
		if exists {
			t.Fatalf("Product exists after deletion")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := productRepository.Delete(context.Background(), uuid.New())
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

func TestPostgresProductRepositoryUpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	t.Run("success", func(t *testing.T) {
		if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
			t.Fatalf("Error truncating table: %v", err)
		}

		category := &domain.Category{
			Id:        uuid.New(),
			Name:      "Test name",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := categoryRepository.Save(context.Background(), category); err != nil {
			t.Fatalf("Error saving category: %v", err)
		}

		product := &domain.Product{
			Id:               uuid.New(),
			Name:             "Test name",
			Description:      "Some test description for product",
			Price:            decimal.NewFromInt(100),
			CategoryId:       category.Id,
			ImageKey:         "imageKey",
			ImageContentType: domain.ImageContentTypePNG,
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := productRepository.Save(context.Background(), product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}

		if err := productRepository.UpdateStatus(context.Background(), product.Id, domain.ProductStatusReady); err != nil {
			t.Fatalf("Error updating product status: %v", err)
		}

		var fetchedStatus domain.ProductStatus
		row := testDb.QueryRow(`SELECT status FROM products WHERE id = $1`, product.Id)
		if err := row.Scan(&fetchedStatus); err != nil {
			t.Fatalf("Error getting product status: %v", err)
		}
		if fetchedStatus != domain.ProductStatusReady {
			t.Errorf("Want product status %s, got %s", domain.ProductStatusReady, fetchedStatus)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := productRepository.UpdateStatus(context.Background(), uuid.New(), domain.ProductStatusReady)
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

func TestPostgresProductRepositoryDeleteExpiredByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	categoryRepository := NewPostgresCategoryRepository(testDb)
	productRepository := NewPostgresProductRepository(testDb)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories, products`); err != nil {
		t.Fatalf("Error truncating table: %v", err)
	}

	category := &domain.Category{
		Id:        uuid.New(),
		Name:      "Test name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := categoryRepository.Save(context.Background(), category); err != nil {
		t.Fatalf("Error saving category: %v", err)
	}

	products := []domain.Product{
		{
			Id:               uuid.New(),
			Name:             "Test name 1",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey1",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 2",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey2",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusReady,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 3",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey3",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
		{
			Id:               uuid.New(),
			Name:             "Test name 4",
			Description:      "Test name description",
			Price:            decimal.NewFromInt(10),
			CategoryId:       category.Id,
			ImageKey:         "imageKey4",
			ImageContentType: "image/png",
			Status:           domain.ProductStatusMissingImage,
			CreatedAt:        time.Now().AddDate(-1, 0, 0),
			UpdatedAt:        time.Now(),
		},
	}

	for _, product := range products {
		if err := productRepository.Save(context.Background(), &product); err != nil {
			t.Fatalf("Error saving product: %v", err)
		}
	}

	wantKeys := []string{"imageKey3", "imageKey4"}
	slices.Sort(wantKeys)

	keys, err := productRepository.DeleteExpiredByStatus(context.Background(), 15*time.Minute)
	if err != nil {
		t.Fatalf("Error deleting expired products: %v", err)
	}
	slices.Sort(keys)

	if len(wantKeys) != len(keys) {
		t.Fatalf("Want %d keys, got %d", len(wantKeys), len(keys))
	}

	for i := 0; i < len(wantKeys); i++ {
		if wantKeys[i] != keys[i] {
			t.Errorf("Want key %s, got %s", wantKeys[i], keys[i])
		}
	}
}
