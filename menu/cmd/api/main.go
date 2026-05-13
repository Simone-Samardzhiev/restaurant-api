package main

import (
	"log"
	"menu/internal/config"
	"menu/internal/database"
	"menu/internal/domain"
	"menu/internal/logger"
	"menu/internal/rate"
	"menu/internal/rest"
	"os/signal"
	"syscall"
	"time"

	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/valkey-io/valkey-go"
)

func main() {
	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	db, err := database.Connect(&appConfig.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	valkeyOption, err := valkey.ParseURL(appConfig.Valkey.Url)
	if err != nil {
		log.Fatalf("Error parsing rate url: %v", err)
	}

	valkeyConn, err := valkey.NewClient(valkeyOption)
	if err != nil {
		log.Fatalf("Error connecting to rate: %v", err)
	}

	if err = database.ApplyMigrations(db, appConfig.Database.MigrationsPath); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	heathCheckHandler := rest.NewHealthHandler(db)

	categoryRepository := database.NewPostgresCategoryRepository(db)
	categoryService := domain.NewDefaultCategoryService(categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryService)

	rateLimitStore := rate.NewValkeyStore(valkeyConn, appConfig.RateLimit)

	router := rest.NewRouter(&rest.RouterConfig{
		App:    &appConfig.App,
		Logger: logger.New(&appConfig.App),
		Store:  rateLimitStore,

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

	if err = db.Close(); err != nil {
		log.Printf("Error closing database connection: %v\n", err)
	}
}
