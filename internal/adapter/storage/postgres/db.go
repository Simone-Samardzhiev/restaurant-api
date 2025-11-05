package postgres

import (
	"database/sql"
	"restaurant/internal/adapter/config"
)

// New establishes connection to postgres database.
func New(dbConfig *config.StorageConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbConfig.DbUrl)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
