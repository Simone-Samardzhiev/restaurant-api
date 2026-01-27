package postgres_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

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
