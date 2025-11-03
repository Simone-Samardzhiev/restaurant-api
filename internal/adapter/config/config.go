package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/joho/godotenv"
)

// getEnv is a helper function for getting environment variable,
// if the variable doesn't exist fallback is returned.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt is a helper function for getting environment variable parsed as int,
// if the variable doesn't exist or is not a valid int, fallback is returned.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		parsedValue, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return parsedValue
	}
	return fallback
}

type (
	// Environment for different app environments.
	Environment string

	// Container holds all environment variables.
	Container struct {
		AppConfig  AppConfig
		DbConfig   DBConfig
		AuthConfig AuthConfig
	}

	// AppConfig holds all environment variable for the application.
	AppConfig struct {
		Environment Environment
		Port        string
	}

	// DBConfig holds all environment variable for the database.
	DBConfig struct {
		URL                string
		MaxIdleConnections int
		MaxOpenConnections int
	}

	// AuthConfig holds all environment variable for the authentication.
	AuthConfig struct {
		Username string
		Password string
	}
)

const (
	Development Environment = "development"
	Production  Environment = "production"
)

func newAppConfig() (AppConfig, error) {
	port := getEnv("PORT", ":8080")
	environment := Environment(getEnv("ENVIRONMENT", "development"))
	if environment != Production && environment != Development {
		return AppConfig{}, fmt.Errorf("invalid environment: %s", environment)
	}

	return AppConfig{
		Port:        port,
		Environment: environment,
	}, nil
}

func newDBConfig() (DBConfig, error) {
	url := os.Getenv("DB_URL")
	if url == "" {
		return DBConfig{}, fmt.Errorf("DB_URL environment variable not set")
	}

	maxIdleConnections := getEnvInt("MAX_IDLE_CONNECTIONS", 10)
	if maxIdleConnections <= 0 {
		return DBConfig{}, fmt.Errorf("max idle connections must be greater than zero: %d", maxIdleConnections)
	}
	maxOpenConnections := getEnvInt("MAX_OPEN_CONNECTIONS", 10)
	if maxOpenConnections <= 0 {
		return DBConfig{}, fmt.Errorf("max open connections must be greater than zero: %d", maxOpenConnections)
	}

	return DBConfig{
		URL:                url,
		MaxIdleConnections: maxIdleConnections,
		MaxOpenConnections: maxOpenConnections,
	}, nil
}

func newAuthConfig() (AuthConfig, error) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	if utf8.RuneCountInString(username) < 8 {
		return AuthConfig{}, fmt.Errorf("username must be at least 8 characters")
	}
	if utf8.RuneCountInString(password) < 8 {
		return AuthConfig{}, fmt.Errorf("password must be at least 8 characters")
	}
	return AuthConfig{
		Username: username,
		Password: password,
	}, nil
}

func New() (*Container, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	appConfig, err := newAppConfig()
	if err != nil {
		return nil, err
	}

	dbConfig, err := newDBConfig()
	if err != nil {
		return nil, err
	}

	authConfig, err := newAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Container{
		AppConfig:  appConfig,
		DbConfig:   dbConfig,
		AuthConfig: authConfig,
	}, nil
}
