package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type (
	// Environment for different app environments.
	Environment string

	// Container holds all environment variables.
	Container struct {
		AppConfig  AppConfig
		DbConfig   StorageConfig
		AuthConfig AuthConfig
	}

	// AppConfig holds all environment variable for the application.
	AppConfig struct {
		Environment      Environment `envconfig:"ENVIRONMENT" required:"true"`
		ImageSavePath    string      `envconfig:"IMAGE_SAVE_PATH" required:"true"`
		ImageServingPath string      `envconfig:"IMAGE_SERVING_PATH" required:"true"`
		Port             string      `envconfig:"PORT" default:":8080"`
	}

	// StorageConfig holds all environment variable for the database.
	StorageConfig struct {
		DbUrl                string `envconfig:"DB_URL" required:"true"`
		DbMaxIdleConnections int    `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"5"`
		DbMaxOpenConnections int    `envconfig:"DB_MAX_OPEN_CONNECTIONS" default:"5"`
	}

	// AuthConfig holds all environment variable for the authentication.
	AuthConfig struct {
		Username string `envconfig:"USERNAME" required:"true"`
		Password string `envconfig:"PASSWORD" required:"true"`
	}
)

const (
	Development Environment = "development"
	Production  Environment = "production"
)

func (e *Environment) Decode(value string) error {
	switch strings.ToLower(value) {
	case "development":
		*e = Development
	case "production":
		*e = Production
	default:
		return fmt.Errorf("invalid environment value: %s", value)
	}
	return nil
}

func New() (*Container, error) {
	if err := godotenv.Load(); err != nil {
		zap.L().Warn("error loading .env file", zap.Error(err))
	}

	var container Container

	if err := envconfig.Process("", &container); err != nil {
		return nil, fmt.Errorf("error mapping environemnt variables: %w", err)
	}

	return &container, nil
}
