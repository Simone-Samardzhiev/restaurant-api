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

// NewDatabase loads and return [Database] from environment variables.
func NewDatabase() (Database, error) {
	var database Database
	if url, ok := os.LookupEnv("DATABASE_URL"); ok {
		database.Url = url
	} else {
		return Database{}, errors.New("config: DATABASE_URL is required but not set")
	}

	if path, ok := os.LookupEnv("MIGRATIONS_PATH"); ok {
		database.MigrationsPath = path
	} else {
		return Database{}, errors.New("config: MIGRATIONS_PATH is required but not set")
	}

	if maxIdleConns, ok := os.LookupEnv("DATABASE_MAX_IDLE_CONNS"); ok {
		val, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABABE_MAX_IDLE_CONNS must be an integer (got: %s): %w", maxIdleConns, err)
		}
		database.MaxIdleConns = val
	} else {
		return Database{}, errors.New("config: DATABABE_MAX_IDLE_CONNS is required but not set")
	}

	if maxIdleConns, ok := os.LookupEnv("DATABASE_MAX_IDLE_CONNS"); ok {
		val, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABASE_MAX_IDLE_CONNS must be an integer (got: %s): %w", maxIdleConns, err)
		}
		database.MaxIdleConns = val
	} else {
		return Database{}, errors.New("config: DATABASE_MAX_IDLE_CONNS is required but not set")
	}

	if maxOpenConns, ok := os.LookupEnv("DATABASE_MAX_OPEN_CONNS"); ok {
		val, err := strconv.Atoi(maxOpenConns)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABASE_MAX_OPEN_CONNS must be an integer (got: %s): %w", maxOpenConns, err)
		}
		database.MaxOpenConns = val
	} else {
		return Database{}, errors.New("config: DATABASE_MAX_OPEN_CONNS is required but not set")
	}

	if maxIdleTime, ok := os.LookupEnv("DATABASE_MAX_IDLE_TIME"); ok {
		val, err := time.ParseDuration(maxIdleTime)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABASE_MAX_IDLE_TIME must be a duration (got: %s): %w", maxIdleTime, err)
		}
		database.MaxIdleTime = val
	} else {
		return Database{}, errors.New("config: DATABASE_MAX_IDLE_TIME is required but not set")
	}

	if maxIdleTime, ok := os.LookupEnv("DATABASE_MAX_IDLE_TIME"); ok {
		val, err := time.ParseDuration(maxIdleTime)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABASE_MAX_IDLE_TIME must be a duration (got: %s): %w", maxIdleTime, err)

		}
		database.MaxIdleTime = val
	} else {
		return Database{}, errors.New("config: DATABASE_MAX_IDLE_TIME is required but not set")
	}

	if maxLifetime, ok := os.LookupEnv("DATABASE_MAX_LIFETIME"); ok {
		val, err := time.ParseDuration(maxLifetime)
		if err != nil {
			return Database{}, fmt.Errorf("config: DATABASE_MAX_LIFETIME must be a duration (got: %s): %w", maxLifetime, err)
		}
		database.MaxLifetime = val
	} else {
		return Database{}, errors.New("config: DATABASE_MAX_LIFETIME is required but not set")
	}

	return database, nil
}

func NewValkey() (Valkey, error) {
	var valkey Valkey
	if url, ok := os.LookupEnv("VALKEY_URL"); ok {
		valkey.Url = url
	} else {
		return Valkey{}, errors.New("config: VALKEY_URL is required but not set")
	}

	return valkey, nil
}

func NewRateLimit() (RateLimit, error) {
	var rateLimit RateLimit

	if limit, ok := os.LookupEnv("RATE_LIMIT_COUNT"); ok {
		val, err := strconv.Atoi(limit)
		if err != nil {
			return RateLimit{}, fmt.Errorf("config: RATE_LIMIT_COUNT must be an integer (got: %s): %w", limit, err)
		}
		rateLimit.Limit = val
	} else {
		return RateLimit{}, errors.New("config: RATE_LIMIT_COUNT is required but not set")
	}

	if window, ok := os.LookupEnv("RATE_LIMIT_WINDOW"); ok {
		val, err := time.ParseDuration(window)
		if err != nil {
			return RateLimit{}, fmt.Errorf("config: RATE_LIMIT_WINDOW must be a duration (got: %s): %w", window, err)
		}
		rateLimit.Window = val
	} else {
		return RateLimit{}, errors.New("config: RATE_LIMIT_WINDOW is required but not set")
	}

	return rateLimit, nil
}

func NewBucket() (Bucket, error) {
	var bucket Bucket
	if url, ok := os.LookupEnv("AWS_ENDPOINT_URL"); ok {
		bucket.BaseEndpoint = url
	} else {
		return Bucket{}, errors.New("config: AWS_ENDPOINT_URL is required but not set")
	}

	if expiry, ok := os.LookupEnv("AWS_UPLOAD_URL_EXPIRY"); ok {
		val, err := time.ParseDuration(expiry)
		if err != nil {
			return Bucket{}, fmt.Errorf("config: AWS_UPLOAD_URL_EXPIRY must be a duration (got: %s): %w", expiry, err)
		}
		bucket.UploadUrlExpiry = val
	} else {
		return Bucket{}, errors.New("config: AWS_UPLOAD_URL_EXPIRY is required but not set")
	}

	if name, ok := os.LookupEnv("AWS_S3_BUCKET_NAME"); ok {
		bucket.Name = name
	} else {
		return Bucket{}, errors.New("config: AWS_S3_BUCKET_NAME environment variable not defined")
	}
	return bucket, nil
}

func NewApp() (App, error) {
	var app App
	if port, ok := os.LookupEnv("ADDR"); ok {
		app.Addr = port
	} else {
		return App{}, errors.New("config: ADDR is required but not set")
	}

	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		environment := Environment(env)
		switch environment {
		case Production, Development:
			app.Env = environment
		default:
			return App{}, errors.New("config: ENVIRONMENT is required but not set")
		}
	}

	return app, nil
}

// NewConfig loads and returns [Config] from environment variables.
func NewConfig() (*Config, error) {
	database, err := NewDatabase()
	if err != nil {
		return nil, err
	}

	valkey, err := NewValkey()
	if err != nil {
		return nil, err
	}

	rateLimit, err := NewRateLimit()
	if err != nil {
		return nil, err
	}

	bucket, err := NewBucket()
	if err != nil {
		return nil, err
	}

	app, err := NewApp()
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
