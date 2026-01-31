package postgres_test

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"restaurant/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var database *sql.DB
var productsSeedQuery string

func connectToDb() (*sql.DB, error) {
	if err := godotenv.Load("../../../../.env"); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	dbUrl, ok := os.LookupEnv("TEST_DB_URL")
	if !ok {
		return nil, fmt.Errorf("TEST_DB_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

func readProductsSeedQuery() (string, error) {
	content, err := os.ReadFile("./testdata/seeds/menu.sql")
	if err != nil {
		return "", fmt.Errorf("failed to read products seeds: %w", err)
	}

	return string(content), nil
}

func TestMain(t *testing.M) {
	db, err := connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	database = db

	query, err := readProductsSeedQuery()
	if err != nil {
		log.Fatal(err)
	}
	productsSeedQuery = query
	os.Exit(t.Run())
}

// parseUUID is a helper function for parsing UUIDs.
func parseUUID(t *testing.T, s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("error parsing UUID: %v", err)
	}
	return id
}

// asPointer is a helper function for returning values as a pointer.
func asPointer[T any](val T) *T {
	return &val
}

// seedMenuTables is a helper function for seeding the tables.
func seedMenuTables(t *testing.T) {
	_, err := database.Exec(productsSeedQuery)
	if err != nil {
		t.Fatalf("error seeding menu tables: %v", err)
	}
}

// truncateMenuTables is a helper cleanup function for truncating tables.
func truncateMenuTables(t *testing.T) {
	if _, err := database.Exec(`TRUNCATE TABLE product_categories, products RESTART IDENTITY CASCADE;`); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// assertNoErr is a helper function for asserting no error is returned.
func assertNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// assertConflictErr is a helper function for asserting an error is domain.Error
// and the type is domain.Conflict.
func assertConflictErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if domainErr.Type != domain.Conflict {
		t.Fatalf("expected conflict, got %v", domainErr.Type)
	}
}

// assertBadRequestErr is a helper function for asserting an error is domain.Error
// and the type is domain.BadRequest.
func assertBadRequestErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if domainErr.Type != domain.BadRequest {
		t.Fatalf("expected bad request, got %v", domainErr.Type)
	}
}

// assertNotFoundErr is a helper function for asserting an error is domain.Error
// and the type is domain.NotFound.
func assertNotFoundErr(t *testing.T, err error) {
	t.Helper()

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}
	if domainErr.Type != domain.NotFound {
		t.Fatalf("expected NotFound, got %v", domainErr.Type)
	}
}
