package rest

import (
	"database/sql"
	"flag"
	"log"
	"menu/internal/postgres"
	"os"
	"testing"
)

// testDb holds connection to the test database.
// Connection is only established if the short flag is not provided.
var testDb *sql.DB

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(m.Run())
	}
	dbUrl, ok := os.LookupEnv("TEST_DATABASE_URL")
	if !ok {
		log.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	database, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err = database.Ping(); err != nil {
		log.Fatalf("Error pinging database connection: %v", err)
	}

	if err = postgres.ApplyMigrations(database, "file://./../../migrations"); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	testDb = database
	os.Exit(m.Run())
}
