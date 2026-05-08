package main

import (
	"log"
	"menu/internal/config"
	"menu/internal/db"
	"menu/internal/domain"
	"menu/internal/logger"
	"menu/internal/rest"
	"os/signal"
	"syscall"
	"time"

	"context"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	database, err := db.Connect(&appConfig.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err = db.ApplyMigrations(database, appConfig.Database.MigrationsPath); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	heathCheckHandler := rest.NewHealthHandler(database)

	categoryRepository := db.NewCategoryRepository(database)
	categoryService := domain.NewDefaultCategoryService(categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryService)

	router := rest.NewRouter(&rest.RouterConfig{
		App:             &appConfig.App,
		Logger:          logger.New(&appConfig.App),
		HeathHandler:    heathCheckHandler,
		CategoryHandler: categoryHandler,
	})

	go func() {
		_ = router.Start()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	log.Println("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = router.Stop(shutdownCtx); err != nil {
		log.Printf("Error shutting down router: %v\n", err)
	}

	if err = database.Close(); err != nil {
		log.Printf("Error closing database connection: %v\n", err)
	}
}
