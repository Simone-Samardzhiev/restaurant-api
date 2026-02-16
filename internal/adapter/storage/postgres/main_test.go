package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
)

// database holds connecting to the test database.
var database *sql.DB

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
	}
	url = defaultURL

	connectToTestDb(url)
	m.Run()
}
