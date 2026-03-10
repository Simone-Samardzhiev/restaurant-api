package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/storage/local"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestAddProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	productRepository := postgres.NewProductRepository(database)
	imageRepository := local.NewImageRepository(tempDir)
	service := menu.NewDefaultProductService(productRepository, imageRepository)
	handler := rest.NewProductHandler(service, "")
	router := test.NewProductRouterFromHandler(handler)

	t.Run("add and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		addProductRequest := rest.AddProductRequest{
			Name:        "New product",
			Description: "Description of a new product",
			Price:       decimal.NewFromFloat(10.25),
			CategoryId:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		}

		request := test.NewAddProductRequest(t, &addProductRequest, validProductImage)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		var addResp rest.ProductResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &addResp)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}

		if addProductRequest.Name != addResp.Name {
			t.Errorf("want name %s, got %s", addProductRequest.Name, addResp.Name)
		}
		if addProductRequest.Description != addResp.Description {
			t.Errorf("want description %s, got %s", addProductRequest.Description, addResp.Description)
		}
		if !addProductRequest.Price.Equal(addResp.Price) {
			t.Errorf("want price %s, got %s", addProductRequest.Price, addResp.Price)
		}
		if addProductRequest.CategoryId != addResp.CategoryId {
			t.Errorf("want category id %s, got %s", addProductRequest.CategoryId, addResp.CategoryId)
		}

		request = test.NewGetProductsRequest(new(addResp.Id.String()), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d; got %d", http.StatusOK, recorder.Code)
		}

		var getResp []rest.ProductResponse
		err = json.NewDecoder(recorder.Body).Decode(&getResp)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}

		if len(getResp) != 1 {
			t.Fatalf("want 1 product; got %d", len(getResp))
		}
	})

	t.Run("add with conflict name", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		addProductRequest := rest.AddProductRequest{
			Name:        "Bruschetta",
			Description: "Description of a new product",
			Price:       decimal.NewFromFloat(10.25),
			CategoryId:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		}
		request := test.NewAddProductRequest(t, &addProductRequest, validProductImage)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d; got %d", http.StatusConflict, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductNameConflict)
	})

	t.Run("add product not non existing category", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		addProductRequest := rest.AddProductRequest{
			Name:        "New product",
			Description: "Description of a new product",
			Price:       decimal.NewFromFloat(10.25),
			CategoryId:  uuid.New(),
		}
		request := test.NewAddProductRequest(t, &addProductRequest, validProductImage)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d; got %d", http.StatusNotFound, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNotFound, domain.ErrorCodeCategoryNotFoundByID)
	})
}

func TestUpdateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	productRepository := postgres.NewProductRepository(database)
	imageRepository := local.NewImageRepository(tempDir)
	service := menu.NewDefaultProductService(productRepository, imageRepository)
	handler := rest.NewProductHandler(service, "")
	router := test.NewProductRouterFromHandler(handler)

	t.Run("update and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const productId = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
		const productName = "New product"
		body, err := json.Marshal(rest.UpdateProductRequest{
			Name: new(productName),
		})
		if err != nil {
			t.Fatalf("failed to encode request body: %s", err)
		}

		request := httptest.NewRequest(http.MethodPatch, "/products/"+productId, bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("want status %d; got %d", http.StatusNoContent, recorder.Code)
		}

		request = test.NewGetProductsRequest(new(productId), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d; got %d", http.StatusOK, recorder.Code)
		}

		var getResp []rest.ProductResponse
		err = json.NewDecoder(recorder.Body).Decode(&getResp)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(getResp) != 1 {
			t.Fatalf("want 1 product; got %d", len(getResp))
		}

		if getResp[0].Name != productName {
			t.Errorf("want name %s, got %s", productName, getResp[0].Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)
		body, err := json.Marshal(rest.UpdateProductRequest{
			Name: new("New name"),
		})
		if err != nil {
			t.Fatalf("failed to encode request body: %s", err)
		}
		request := httptest.NewRequest(http.MethodPatch, "/products/"+uuid.NewString(), bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d; got %d", http.StatusNotFound, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductNotFound, domain.ErrorCodeProductNotFoundByID)
	})

	t.Run("update with conflict name", func(t *testing.T) {
		test.SeedMenuTables(t, database)
		const productId = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"

		body, err := json.Marshal(rest.UpdateProductRequest{
			Name: new("Garlic Bread"),
		})
		if err != nil {
			t.Fatalf("failed to encode request body: %s", err)
		}
		request := httptest.NewRequest(http.MethodPatch, "/products/"+productId, bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d; got %d", http.StatusConflict, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductNameConflict)
	})

	t.Run("update with non existing category", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const productId = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
		body, err := json.Marshal(rest.UpdateProductRequest{
			CategoryID: new(uuid.New()),
		})
		if err != nil {
			t.Fatalf("failed to encode request body: %s", err)
		}
		request := httptest.NewRequest(http.MethodPatch, "/products/"+productId, bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d; got %d", http.StatusNotFound, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNotFound, domain.ErrorCodeCategoryNotFoundByID)
	})
}

func TestUpdateProductImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	productRepository := postgres.NewProductRepository(database)
	imageRepository := local.NewImageRepository(tempDir)
	service := menu.NewDefaultProductService(productRepository, imageRepository)
	handler := rest.NewProductHandler(service, "")
	router := test.NewProductRouterFromHandler(handler)

	t.Run("update", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const productId = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
		request := httptest.NewRequest(http.MethodPut, "/products/"+productId+"/image", bytes.NewReader(validProductImage))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("want status %d; got %d", http.StatusCreated, recorder.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)
		request := httptest.NewRequest(http.MethodPut, "/products/"+uuid.NewString()+"/image", bytes.NewReader(validProductImage))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d; got %d", http.StatusNotFound, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductNotFound, domain.ErrorCodeProductNotFoundByID)
	})
}

func TestDeleteProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	productRepository := postgres.NewProductRepository(database)
	imageRepository := local.NewImageRepository(tempDir)
	service := menu.NewDefaultProductService(productRepository, imageRepository)
	handler := rest.NewProductHandler(service, "")
	router := test.NewProductRouterFromHandler(handler)

	t.Run("delete and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const productId = "a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"
		request := httptest.NewRequest(http.MethodDelete, "/products/"+productId, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d; got %d", http.StatusOK, recorder.Code)
		}

		request = test.NewGetProductsRequest(new(productId), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d; got %d", http.StatusOK, recorder.Code)
		}

		var getResp []rest.ProductResponse
		err := json.NewDecoder(recorder.Body).Decode(&getResp)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(getResp) != 0 {
			t.Fatalf("want 0 products; got %d", len(getResp))
		}
	})

	t.Run("not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		request := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.NewString(), nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d; got %d", http.StatusNotFound, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductNotFound, domain.ErrorCodeProductNotFoundByID)
	})

	t.Run("conflict with orders", func(t *testing.T) {
		test.SeedMenuTables(t, database)
		test.SeedOrderTables(t, database)

		const productId = "d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4"

		request := httptest.NewRequest(http.MethodDelete, "/products/"+productId, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d; got %d", http.StatusConflict, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeProductHasLinkedOrders)
	})
}
