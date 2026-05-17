package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type (
	// Database represents database config.
	Database struct {
		Url            string
		MigrationsPath string
		MaxIdleConns   int
		MaxOpenConns   int
		MaxIdleTime    time.Duration
		MaxLifetime    time.Duration
	}

	// Valkey represents valkey config.
	Valkey struct {
		Url string
	}

	// RateLimit rate limiting config.
	RateLimit struct {
		Limit  int
		Window time.Duration
	}

	// Bucket represents s3 bucket config.
	Bucket struct {
		BaseEndpoint    string
		UploadUrlExpiry time.Duration
		Name            string
	}

	// Environment represents app environment.
	Environment string

	// App represents the application config.
	App struct {
		Addr string
		Env  Environment
	}

	// Config combines all configurations.
	Config struct {
		Database
		Valkey
		RateLimit
		Bucket
		App
	}
)

// newDatabase loads and return [Database] from environment variables.
func newDatabase() (Database, error) {
	var database Database
	if url, ok := os.LookupEnv("DATABASE_URL"); ok {
		database.Url = url
	} else {
		return Database{}, errors.New("DATABASE_URL environment variable not defined")
	}

	if path, ok := os.LookupEnv("MIGRATIONS_PATH"); ok {
		database.MigrationsPath = path
	} else {
		return Database{}, errors.New("MIGRATIONS_PATH environment variable not defined")
	}

	if maxIdleConns, err := strconv.Atoi(os.Getenv("DATABASE_MAX_IDLE_CONNS")); err == nil {
		database.MaxIdleConns = maxIdleConns
	} else {
		return Database{}, fmt.Errorf("DATABASE_MAX_IDLE_CONNS environment variable not defined: %v", err)
	}

	if maxOpenConns, err := strconv.Atoi(os.Getenv("DATABASE_MAX_OPEN_CONNS")); err == nil {
		database.MaxOpenConns = maxOpenConns
	} else {
		return Database{}, fmt.Errorf("DATABASE_MAX_OPEN_CONNS environment variable not defined: %v", err)
	}

	if maxIdleTime, err := time.ParseDuration(os.Getenv("DATABASE_MAX_IDLE_TIME")); err == nil {
		database.MaxIdleTime = maxIdleTime
	} else {
		return Database{}, fmt.Errorf("DATABASE_MAX_IDLE_TIME environment variable not defined: %v", err)
	}

	if maxLifetime, err := time.ParseDuration(os.Getenv("DATABASE_MAX_LIFETIME")); err == nil {
		database.MaxLifetime = maxLifetime
	} else {
		return Database{}, fmt.Errorf("DATABASE_MAX_LIFETIME environment variable not defined: %v", err)
	}

	return database, nil
}

func newValkey() (Valkey, error) {
	var valkey Valkey
	if url, ok := os.LookupEnv("VALKEY_URL"); ok {
		valkey.Url = url
	} else {
		return Valkey{}, errors.New("VALKEY_URL environment variable not defined")
	}

	return valkey, nil
}

func newRateLimit() (RateLimit, error) {
	var ratelimit RateLimit
	if limit, err := strconv.Atoi(os.Getenv("RATE_LIMIT_COUNT")); err == nil {
		ratelimit.Limit = limit
	} else {
		return RateLimit{}, fmt.Errorf("RATE_LIMIT_COUNT environment variable not defined: %v", err)
	}

	if window, err := time.ParseDuration(os.Getenv("RATE_LIMIT_WINDOW")); err == nil {
		ratelimit.Window = window
	} else {
		return RateLimit{}, fmt.Errorf("RATE_LIMIT_WINDOW environment variable not defined: %v", err)
	}

	return ratelimit, nil
}

func newBucket() (Bucket, error) {
	var bucket Bucket
	if url, ok := os.LookupEnv("BUCKET_URL"); ok {
		bucket.BaseEndpoint = url
	} else {
		return Bucket{}, errors.New("BUCKET_URL environment variable not defined")
	}

	if expiry, err := time.ParseDuration(os.Getenv("BUCKET_EXPIRY")); err == nil {
		bucket.UploadUrlExpiry = expiry
	} else {
		return Bucket{}, fmt.Errorf("BUCKET_EXPIRY environment variable not defined: %v", err)
	}

	if name, ok := os.LookupEnv("BUCKET_NAME"); ok {
		bucket.Name = name
	} else {
		return Bucket{}, errors.New("BUCKET_NAME environment variable not defined")
	}
	return bucket, nil
}

func newApp() (App, error) {
	var app App
	if port, ok := os.LookupEnv("ADDR"); ok {
		app.Addr = port
	} else {
		return App{}, errors.New("ADDR environment variable not defined")
	}

	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		environment := Environment(env)
		switch environment {
		case Production, Development:
			app.Env = environment
		default:
			return App{}, errors.New("ENVIRONMENT environment variable not defined")
		}
	}

	return app, nil
}

// NewConfig loads and returns [Config] from environment variables.
func NewConfig() (*Config, error) {
	database, err := newDatabase()
	if err != nil {
		return nil, err
	}

	valkey, err := newValkey()
	if err != nil {
		return nil, err
	}

	rateLimit, err := newRateLimit()
	if err != nil {
		return nil, err
	}

	bucket, err := newBucket()
	if err != nil {
		return nil, err
	}

	app, err := newApp()
	if err != nil {
		return nil, err
	}

	return &Config{
		database,
		valkey,
		rateLimit,
		bucket,
		app,
	}, nil
}
