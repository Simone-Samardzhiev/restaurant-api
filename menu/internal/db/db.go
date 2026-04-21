package db

import (
	"database/sql"
	"menu/internal/config"

	_ "github.com/lib/pq"
)

// Connect establishes a new postgres connection.
func Connect(c *config.Database) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.Url)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetConnMaxIdleTime(c.MaxIdleTime)
	db.SetConnMaxLifetime(c.MaxLifetime)

	return db, nil
}
