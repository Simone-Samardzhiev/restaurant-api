package postgres_test

import (
	"database/sql"
	_ "embed"
	"log"
	"os"
	"restaurant/internal/test"
	"testing"
)

// database holds connecting to the test database.
var database *sql.DB

func TestMain(m *testing.M) {
	url, ok := os.LookupEnv("TEST_DB_URL")
	if !ok {
		log.Fatal("TEST_DB_URL not set")
	}

	database = test.ConnectToDb(url)
	os.Exit(m.Run())
}
