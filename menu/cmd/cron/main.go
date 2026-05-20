package main

import (
	"context"
	"database/sql"
	"log"
	"menu/internal/database"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatalf("DATABASE_URL is required but not set")
	}

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	repository := database.NewPostgresProductRepository(db)
	if err = repository.DeleteExpiredByStatus(context.Background(), 15*time.Minute); err != nil {
		log.Fatalf("Error deleting expired by status: %v", err)
	}
}
