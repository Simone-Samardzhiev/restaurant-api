package main

import (
	"log"
	"menu/internal/config"
	"menu/internal/db"
	"os/signal"
	"syscall"

	"context"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	database, err := db.Connect(&conf.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err = db.ApplyMigrations(database, conf.Database.MigrationsPath); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()
	<-ctx.Done()

	log.Println("Shutting down")
	if err = database.Close(); err != nil {
		log.Fatalf("Error closing database connection: %v", err)
	}
}
