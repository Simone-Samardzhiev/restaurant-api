package main

import (
	"context"
	"database/sql"
	"log"
	"menu/internal/config"
	"menu/internal/database"
	"menu/internal/storage"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	bucketConfig, err := config.NewBucket()
	if err != nil {
		log.Fatalf("Error loading bucket config: %v", err)
	}
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Error loading AWS configuration: %v", err)
	}
	s3client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(bucketConfig.BaseEndpoint)
	})
	s3storage := storage.NewS3ImageStorage(s3client, 0, bucketConfig.Name)

	keys, err := repository.DeleteExpiredByStatus(context.Background(), 15*time.Minute)
	if err != nil {
		log.Fatalf("Error deleting expired images: %v", err)
	}
	if err = s3storage.DeleteMultiple(context.Background(), keys); err != nil {
		log.Fatalf("Error deleting images of expired products: %v", err)
	}
}
