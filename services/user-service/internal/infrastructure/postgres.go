package infrastructure

import (
	"database/sql"
)

func NewPostgresDB() *sql.DB {
	// Return mock DB or connect to real one
	return &sql.DB{}
}
