package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSessionRepositorySave(t *testing.T) {
	test.TruncateOrderTables(t, database)
	repository := postgres.NewSessionRepository(database)

	const tableNumber = 10
	const status = "open"

	session, err := repository.Save(context.Background(), test.Must(order.ParseAddSessionRequest(tableNumber, status)))
	if err != nil {
		t.Fatalf("want no error, got: %v", err)
	}

	if session.Table.Number() != tableNumber {
		t.Errorf("want table number %d, got %d", tableNumber, session.Table.Number())
	}

	if session.Status.String() != status {
		t.Errorf("want status %s, got %s", status, session.Status.String())
	}
}

func TestSessionRepositoryGet(t *testing.T) {
	test.SeedOrderTables(t, database)
	repository := postgres.NewSessionRepository(database)

	sessions, err := repository.Get(context.Background())
	if err != nil {
		t.Fatalf("want no error, got: %v", err)
	}

	ids := []uuid.UUID{
		uuid.MustParse("88888888-8888-8888-8888-000000000001"),
		uuid.MustParse("88888888-8888-8888-8888-000000000002"),
		uuid.MustParse("88888888-8888-8888-8888-000000000003"),
		uuid.MustParse("88888888-8888-8888-8888-000000000004"),
		uuid.MustParse("88888888-8888-8888-8888-000000000005"),
	}

	test.CheckEntities(t, ids, sessions, func(session order.Session) uuid.UUID {
		return session.Id
	})
}

func TestSessionRepositoryUpdate(t *testing.T) {
	tests := []struct {
		name             string
		request          *order.UpdateSessionRequest
		wantErr          bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:    "success",
			request: test.Must(order.ParseUpdateSessionRequest(uuid.MustParse("88888888-8888-8888-8888-000000000001"), new(10), new("closed"))),
		},
		{
			name:             "not found",
			request:          test.Must(order.ParseUpdateSessionRequest(uuid.New(), new(10), new("closed"))),
			wantErr:          true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeSessionNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeSessionNotFoundByID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedOrderTables(t, database)
			repository := postgres.NewSessionRepository(database)
			err := repository.Update(context.Background(), tt.request)

			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Errorf("want no error, got: %v", err)
			}
		})
	}
}

func TestSessionRepositoryGetById(t *testing.T) {
	tests := []struct {
		name             string
		id               uuid.UUID
		wantError        bool
		wantErrorKind    domain.ErrorKind
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			id:   uuid.MustParse("88888888-8888-8888-8888-000000000001"),
		},
		{
			name:             "not found",
			id:               uuid.New(),
			wantError:        true,
			wantErrorKind:    domain.ErrorKindNotFound,
			wantErrorCode:    domain.ErrorCodeSessionNotFound,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeSessionNotFoundByID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedOrderTables(t, database)
			repository := postgres.NewSessionRepository(database)

			_, err := repository.GetById(context.Background(), tt.id)
			if tt.wantError {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Errorf("want no error, got: %v", err)
			}
		})
	}
}

func TestSessionRepositoryGetBill(t *testing.T) {
	tests := []struct {
		name      string
		id        uuid.UUID
		wantPrice decimal.Decimal
		wantIds   []uuid.UUID

		wantErr       bool
		wantErrorKind domain.ErrorKind
		wantErrorCode domain.ErrorCode
	}{
		{
			name:      "success",
			id:        uuid.MustParse("88888888-8888-8888-8888-000000000001"),
			wantPrice: decimal.NewFromFloat(6),
			wantIds: []uuid.UUID{
				uuid.MustParse("d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4"),
				uuid.MustParse("d5d5d5d5-d5d5-d5d5-d5d5-d5d5d5d5d5d5"),
			},
		},
		{
			name:          "error not finished orderes",
			id:            uuid.MustParse("88888888-8888-8888-8888-000000000002"),
			wantErr:       true,
			wantErrorKind: domain.ErrorKindBadRequest,
			wantErrorCode: domain.ErrorCodeOrderNotFinished,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SeedMenuTables(t, database)
			test.SeedOrderTables(t, database)

			repository := postgres.NewSessionRepository(database)
			bill, err := repository.GetBill(context.Background(), tt.id)
			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}

			if !tt.wantPrice.Equal(bill.Price) {
				t.Errorf("want bill price: %v, got %v", tt.wantPrice, bill.Price)
			}

			test.CheckEntities(t, tt.wantIds, bill.Items, func(item order.BillItem) uuid.UUID {
				return item.ProductId
			})
		})
	}
}
