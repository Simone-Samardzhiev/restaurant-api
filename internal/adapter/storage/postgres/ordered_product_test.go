package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

func TestOrderedProductRepositorySave(t *testing.T) {
	tests := []struct {
		name             string
		request          *order.AddOrderedProductRequest
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			request: order.NewAddOrderedProductRequest(
				uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				uuid.MustParse("88888888-8888-8888-8888-000000000001"),
			),
		},
		{
			name: "product not found",
			request: order.NewAddOrderedProductRequest(
				uuid.New(),
				uuid.MustParse("88888888-8888-8888-8888-000000000001"),
			),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindNotFound,
			wantErrorCode: domain.ErrorCodeProductNotFound,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNotFoundByID,
			},
		},
		{
			name: "session not found",
			request: order.NewAddOrderedProductRequest(
				uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1"),
				uuid.New(),
			),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindNotFound,
			wantErrorCode: domain.ErrorCodeSessionNotFound,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeSessionNotFoundByID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			test.SeedOrderTables(t, database)
			repository := postgres.NewOrderedProductRepository(database)

			orderedProduct, err := repository.Save(context.Background(), tt.request)
			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Fatalf("want no errors, got %v", err)
			}

			if orderedProduct.ProductId != tt.request.ProductId {
				t.Errorf("got product id %v, want %v", orderedProduct.ProductId, tt.request.ProductId)
			}
			if orderedProduct.SessionId != tt.request.SessionId {
				t.Errorf("got session id %v, want %v", orderedProduct.SessionId, tt.request.SessionId)
			}
		})
	}
}
