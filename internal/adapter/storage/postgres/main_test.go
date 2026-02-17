package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"restaurant/internal/domain"
	"testing"
)

// database holds connecting to the test database.
var database *sql.DB

// matchErrorCodes verifies details error codes match the codes in details
// ignoring the order.
func matchErrorCodes(t *testing.T, expectedCodes []domain.ErrorCode, details []domain.ErrorDetail) {
	t.Helper()

	if len(details) != len(expectedCodes) {
		t.Fatalf("length mismatch: expected codes %d, got %d ", len(expectedCodes), len(details))
	}

	var counter = map[domain.ErrorCode]int{}
	for _, code := range expectedCodes {
		counter[code]++
	}

	for _, detail := range details {
		if counter[detail.Code] == 0 {
			t.Errorf("unexpected error code: %s", detail.Code)
		}
		counter[detail.Code]--
	}

	for code, count := range counter {
		if count != 0 {
			t.Errorf("missing error code: %s", code)
		}
	}
}

// connectToTestDb establishes and checks connection to postgres database.
func connectToTestDb(url string) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Panicf("error connecting to test database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Panicf("error pinging test database: %v", err)
	}

	database = db
}

// defaultURL holds the default postgres url for testing.
const defaultURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

func TestMain(m *testing.M) {
	var url string
	if val, ok := os.LookupEnv("TEST_DB_URL"); ok {
		url = val
	} else {
		url = defaultURL
	}

	connectToTestDb(url)
	os.Exit(m.Run())
}
