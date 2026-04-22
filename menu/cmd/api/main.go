package main

import (
	"log"
	"menu/internal/config"
	"menu/internal/db"
	"menu/internal/domain"
	"menu/internal/rest"
	"os/signal"
	"syscall"
	"time"

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

	categoryRepository := db.NewCategoryRepository(database)
	categoryService := domain.NewDefaultCategoryService(categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryService)

	router := rest.NewRouter(&conf.App, categoryHandler)

	go func() {
		_ = router.Start()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()
	<-ctx.Done()

	log.Println("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = router.Stop(shutdownCtx); err != nil {
		log.Fatalf("Error shutting down router: %v", err)
	}

	if err = database.Close(); err != nil {
		log.Fatalf("Error closing database connection: %v", err)
	}
}
