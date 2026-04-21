package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type (
	// Database represents database config.
	Database struct {
		Url          string
		MaxIdleConns int
		MaxOpenConns int
		MaxIdleTime  time.Duration
		MaxLifetime  time.Duration
	}

	// Config combines all configurations.
	Config struct {
		Database
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

// NewConfig loads and returns [Config] from environment variables.
func NewConfig() (*Config, error) {
	database, err := newDatabase()
	if err != nil {
		return nil, err
	}

	return &Config{database}, nil
}
