package postgres_test

import (
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"context"
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
