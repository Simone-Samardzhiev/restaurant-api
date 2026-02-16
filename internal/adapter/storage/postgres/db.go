package postgres

import (
	"database/sql"
	"restaurant/internal/adapter/config"
)

// New establishes and checks connection to postgres database.
func New(storageConfig *config.StorageConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", storageConfig.DbUrl)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(storageConfig.DbMaxIdleConnections)
	db.SetMaxOpenConns(storageConfig.DbMaxOpenConnections)
	return db, nil
}
