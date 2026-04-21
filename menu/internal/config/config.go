package config

import "time"

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
