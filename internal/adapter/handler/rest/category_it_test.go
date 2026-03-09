package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

func TestAddCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repository := postgres.NewCategoryRepository(database)
	service := menu.NewDefaultCategoryService(repository)
	handler := rest.NewCategoryHandler(service)
	router := test.NewCategoryRouterFromHandler(handler)

	t.Run("add and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const categoryName = "New Category"

		body, err := json.Marshal(rest.AddCategoryRequest{Name: categoryName})
		if err != nil {
			t.Fatalf("error encoding request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("want status %d, got %d", http.StatusCreated, recorder.Code)
		}

		var addResp rest.CategoryResponse
		err = json.NewDecoder(recorder.Body).Decode(&addResp)
		if err != nil {
			t.Fatalf("error decoding response body: %v", err)
		}
		if addResp.Name != "New Category" {
			t.Errorf("want category name New Category, got %s", addResp.Name)
		}

		query := url.Values{}
		query.Add("id", addResp.Id.String())
		request = httptest.NewRequest(http.MethodGet, "/categories?"+query.Encode(), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d, got %d", http.StatusOK, recorder.Code)
		}
		var getResp []rest.CategoryResponse
		err = json.NewDecoder(recorder.Body).Decode(&getResp)
		if err != nil {
			t.Fatalf("error decoding response body: %v", err)
		}

		if len(getResp) != 1 {
			t.Fatalf("want 1 category, got %d", len(getResp))
		}
		if getResp[0].Name != categoryName {
			t.Errorf("want category name %s , got %s", categoryName, getResp[0].Name)
		}
	})

	t.Run("add with conflicting name", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		body, err := json.Marshal(rest.AddCategoryRequest{Name: "Drinks"})
		if err != nil {
			t.Fatalf("error encoding request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d, got %d", http.StatusConflict, recorder.Code)
		}

		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNameConflict)
	})
}

func TestUpdateCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repository := postgres.NewCategoryRepository(database)
	service := menu.NewDefaultCategoryService(repository)
	handler := rest.NewCategoryHandler(service)
	router := test.NewCategoryRouterFromHandler(handler)
	t.Run("update and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const categoryName = "New name"
		const categoryId = "11111111-1111-1111-1111-111111111111"
		body, err := json.Marshal(rest.UpdateCategoryRequest{
			Name: new(categoryName),
		})
		if err != nil {
			t.Fatalf("error encoding request body: %v", err)
		}

		request := httptest.NewRequest(http.MethodPatch, "/categories/"+categoryId, bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("want status %d, got %d", http.StatusNoContent, recorder.Code)
		}

		query := url.Values{}
		query.Add("id", categoryId)
		request = httptest.NewRequest(http.MethodGet, "/categories?"+query.Encode(), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d, got %d", http.StatusOK, recorder.Code)
		}
		var getResp []rest.CategoryResponse
		err = json.NewDecoder(recorder.Body).Decode(&getResp)
		if err != nil {
			t.Fatalf("error decoding response body: %v", err)
		}

		if len(getResp) != 1 {
			t.Fatalf("want 1 category, got %d", len(getResp))
		}
		if getResp[0].Name != categoryName {
			t.Errorf("want category name %s, got %s", categoryName, getResp[0].Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		body, err := json.Marshal(rest.UpdateCategoryRequest{
			Name: new("name"),
		})
		if err != nil {
			t.Fatalf("error encoding request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPatch, "/categories/"+uuid.NewString(), bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d, got %d", http.StatusNotFound, recorder.Code)
		}

		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNotFound, domain.ErrorCodeCategoryNotFoundByID)
	})

	t.Run("conflict name", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const categoryName = "Main Dishes"
		const categoryId = "11111111-1111-1111-1111-111111111111"
		body, err := json.Marshal(rest.UpdateCategoryRequest{
			Name: new(categoryName),
		})
		if err != nil {
			t.Fatalf("error encoding request body: %v", err)
		}

		request := httptest.NewRequest(http.MethodPatch, "/categories/"+categoryId, bytes.NewReader(body))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d, got %d", http.StatusConflict, recorder.Code)
		}
		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNameConflict)
	})
}

func TestDeleteCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repository := postgres.NewCategoryRepository(database)
	service := menu.NewDefaultCategoryService(repository)
	handler := rest.NewCategoryHandler(service)
	router := test.NewCategoryRouterFromHandler(handler)

	t.Run("delete and fetch", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const categoryId = "66666666-6666-6666-6666-666666666666"
		request := httptest.NewRequest(http.MethodDelete, "/categories/"+categoryId, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d, got %d", http.StatusOK, recorder.Code)
		}

		query := url.Values{}
		query.Add("id", categoryId)
		request = httptest.NewRequest(http.MethodGet, "/categories?"+query.Encode(), nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("want status %d, got %d", http.StatusOK, recorder.Code)
		}
		var getResp []rest.CategoryResponse
		if err := json.NewDecoder(recorder.Body).Decode(&getResp); err != nil {
			t.Fatalf("error decoding response body: %v", err)
		}
		if len(getResp) != 0 {
			t.Fatalf("want 0 category, got %d", len(getResp))
		}
	})

	t.Run("not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		request := httptest.NewRequest(http.MethodDelete, "/categories/"+uuid.NewString(), nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("want status %d, got %d", http.StatusNotFound, recorder.Code)
		}

		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryNotFound, domain.ErrorCodeCategoryNotFoundByID)
	})

	t.Run("conflict with products", func(t *testing.T) {
		test.SeedMenuTables(t, database)

		const categoryId = "11111111-1111-1111-1111-111111111111"
		request := httptest.NewRequest(http.MethodDelete, "/categories/"+categoryId, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("want status %d, got %d", http.StatusConflict, recorder.Code)
		}

		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeCategoryHasLinkedProducts)
	})
}
