package main

import (
	"log"
	"menu/internal/config"
	"menu/internal/database"
	"menu/internal/domain"
	"menu/internal/logger"
	"menu/internal/rate"
	"menu/internal/rest"
	"menu/internal/storage"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/valkey-io/valkey-go"
)

func main() {
	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(&appConfig.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err = database.ApplyMigrations(db, appConfig.Database.MigrationsPath); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Error loading AWS configuration: %v", err)
	}

	s3client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(appConfig.BaseEndpoint)
		if appConfig.Env == config.Production {
			o.UsePathStyle = false
		} else {
			o.UsePathStyle = true
		}
	})

	valkeyOption, err := valkey.ParseURL(appConfig.Valkey.Url)
	if err != nil {
		log.Fatalf("Error parsing valkey url: %v", err)
	}

	valkeyConn, err := valkey.NewClient(valkeyOption)
	if err != nil {
		log.Fatalf("Error connecting to valkey: %v", err)
	}

	appLogger := logger.New(&appConfig.App)
	rateLimitStore := rate.NewValkeyStore(valkeyConn, appConfig.RateLimit)
	heathCheckHandler := rest.NewHealthHandler(db, valkeyConn)

	categoryRepository := database.NewPostgresCategoryRepository(db)
	categoryService := domain.NewDefaultCategoryService(categoryRepository)
	categoryHandler := rest.NewCategoryHandler(categoryService)

	productRepository := database.NewPostgresProductRepository(db)
	imageStorage := storage.NewS3ImageStorage(s3client, appConfig.UploadUrlExpiry, appConfig.Bucket.Name)
	productService := domain.NewDefaultProductService(productRepository, imageStorage, appLogger)
	productHandler := rest.NewProductHandler(appConfig.BaseImagesUrl, productService)

	router := rest.NewRouter(&rest.RouterConfig{
		App:    &appConfig.App,
		Logger: appLogger,
		Store:  rateLimitStore,

		HeathHandler:    heathCheckHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
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
