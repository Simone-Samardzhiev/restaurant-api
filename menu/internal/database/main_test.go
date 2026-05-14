package database

import (
	"database/sql"
	"flag"
	"log"
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

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database connection: %v", err)
	}

	if err = ApplyMigrations(db, "file://./../../migrations"); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	testDb = db
	os.Exit(m.Run())
}
