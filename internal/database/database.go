package database

import (
	"database/sql"
)

// DB wraps the database connection
type DB struct {
	DB *sql.DB
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
