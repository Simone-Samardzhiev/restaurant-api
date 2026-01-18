package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/logger"
	"restaurant/internal/adapter/storage/postgres"
	"restaurant/internal/domain/menu"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/net/context"
)

func main() {
	container, err := config.New()
	if err != nil {
		log.Printf("error loading config: %v", err)
		os.Exit(1)
	}

	if err = logger.SetZapLogger(&container.AppConfig); err != nil {
		log.Printf("error setting zap logger: %v", err)
		os.Exit(1)
	}

	db, err := postgres.New(&container.DbConfig)
	if err != nil {
		log.Printf("error connecting to database: %v", err)
		os.Exit(1)
	}

	// Product
	productRepository := postgres.NewProductRepository(db)
	productService := menu.NewService(productRepository)
	productHandler := rest.NewProductHandler(productService)

	router := handler.NewRouter(container, productHandler)
	go func() {
		err = router.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("error starting http server: %v", err)
			os.Exit(1)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	<-stopChan

	if err = db.Close(); err != nil {
		log.Printf("error closing db: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = router.Stop(ctx); err != nil {
		log.Printf("error shutting down http server: %v", err)
		os.Exit(1)
	}
}
