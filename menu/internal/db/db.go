package db

import (
	"database/sql"
	"errors"
	"menu/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

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

// ApplyMigrations applies all migrations.
//
// Note: The path should start with *file://*
func ApplyMigrations(db *sql.DB, path string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}
	return nil
}
