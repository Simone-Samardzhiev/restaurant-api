package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"context"

	"github.com/google/uuid"
)

func TestSessionRepositorySave(t *testing.T) {
	t.Cleanup(func() {
		test.TruncateOrderTables(t, database)
	})
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
