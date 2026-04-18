package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

func TestGetDetails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	sessionRepository := postgres.NewSessionRepository(database)
	orderedProductRepository := postgres.NewOrderedProductRepository(database)
	service := order.NewDefaultSessionService(sessionRepository, orderedProductRepository)
	handler := rest.NewSessionHandler(service)
	router := test.NewSessionRouterFromHandler(handler)

	t.Run("get", func(t *testing.T) {
		const sessionID = "88888888-8888-8888-8888-000000000001"

		test.SeedMenuTables(t, database)
		test.SeedOrderTables(t, database)

		request := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("got %d, want %d", recorder.Code, http.StatusOK)
		}

		var response rest.SessionDetailsResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}

		if response.Id.String() != sessionID {
			t.Errorf("want id %s, got %s", sessionID, response.Id.String())
		}
		if response.Table != 10 {
			t.Errorf("want table 10, got %d", response.Table)
		}
		if response.Status != "open" {
			t.Errorf("want status open, got %s", response.Status)
		}

		test.CheckEntities(t, []uuid.UUID{
			uuid.MustParse("99999999-9999-9999-9999-000000000001"),
			uuid.MustParse("99999999-9999-9999-9999-000000000002"),
		}, response.OrderedProducts, func(response rest.OrderedProductResponse) uuid.UUID {
			return response.Id
		})
	})

	t.Run("session not found", func(t *testing.T) {
		test.SeedMenuTables(t, database)
		test.SeedOrderTables(t, database)

		request := httptest.NewRequest(http.MethodGet, "/sessions/"+uuid.NewString(), nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("got %d, want %d", recorder.Code, http.StatusNotFound)
		}

		test.CheckErrorResponse(t, recorder.Body, domain.ErrorCodeSessionNotFound, domain.ErrorCodeSessionNotFoundByID)
	})
}
