package rest

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"menu/internal/database"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// testDb holds connection to the test database.
// Connection is only established if the short flag is not provided.
var testDb *sql.DB

// testS3Client connects to local mock of s3 for testing.
// Connected only if the short flag is not provided.
var testS3Client *s3.Client

// testS3BucketName
var testS3BucketName string

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(m.Run())
	}
	dbUrl, ok := os.LookupEnv("TEST_DATABASE_URL")
	if !ok {
		log.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database connection: %v", err)
	}

	if err = database.ApplyMigrations(db, "file://./../../migrations"); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}
	testDb = db

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Error loading AWS config: %v", err)
	}
	testS3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	if bucketName, ok := os.LookupEnv("TEST_AWS_S3_BUCKET_NAME"); ok {
		testS3BucketName = bucketName
	} else {
		log.Fatal("TEST_AWS_S3_BUCKET_NAME environment variable not set")
	}

	os.Exit(m.Run())
}
