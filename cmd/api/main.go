package main

import (
	"log"
	"os"
	"os/signal"
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/logger"
	"restaurant/internal/adapter/storage/local"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain/menu"
	"syscall"
	"time"

	"context"
)

func main() {
	container, err := config.New()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if err = logger.SetZapLogger(&container.AppConfig); err != nil {
		log.Fatalf("error setting logger: %v", err)
	}

	db, err := postgres.New(&container.DbConfig)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	// Menu
	// Categories
	categoryRepository := postgres.NewCategoryRepository(db)
	categoryService := menu.NewDefaultCategoryService(categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryService)

	// Products
	productRepository := postgres.NewProductRepository(db)
	imageRepository := local.NewImageRepository(container.AppConfig.ImageSavePath)
	productService := menu.NewDefaultProductService(productRepository, imageRepository)
	productHandler := rest.NewProductHandler(productService, container.AppConfig.ImageServingPath)

	// start up tasks
	if err = imageRepository.CreateSavePath(); err != nil {
		log.Fatalf("error creating save path for images: %v", err)
	}

	router := handler.NewRouter(container, handler.Handlers{
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		router.Run()
	}()

	<-signalChan

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = router.Shutdown(ctx); err != nil {
		log.Fatalf("error shutting down server gracefully: %v", err)
	}

	log.Println("server shutdown gracefully")

	if err = db.Close(); err != nil {
		log.Fatalf("error closing database connection: %v", err)
	}
	log.Println("database close gracefully")
}
