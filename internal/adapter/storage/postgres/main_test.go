package postgres_test

import (
	"database/sql"
	_ "embed"
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

//go:embed testdata/seeds/menu.sql
var seedMenuTablesQuery string

// seedMenuTables seeds table for products and product categories with [seedMenuTablesQuery].
func seedMenuTables(t *testing.T) {
	t.Helper()

	if _, err := database.Exec(seedMenuTablesQuery); err != nil {
		t.Fatalf("error seeding menu tables: %v", err)
	}

	t.Cleanup(func() {
		if _, err := database.Exec(`TRUNCATE TABLE products, product_categories RESTART IDENTITY CASCADE `); err != nil {
			t.Fatalf("error truncating tables: %v", err)
		}
	})
}

func TestMain(m *testing.M) {
	url, ok := os.LookupEnv("TEST_DB_URL")
	if !ok {
		log.Fatal("TEST_DB_URL not set")
	}

	connectToTestDb(url)
	os.Exit(m.Run())
}
