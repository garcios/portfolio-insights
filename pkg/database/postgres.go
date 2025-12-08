package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresDB creates a new PostgreSQL database connection.
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Password == "" || cfg.DBName == "" || cfg.SSLMode == "" {
		return nil, fmt.Errorf("missing required database configuration")
	}

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	maxOpenConns := 25
	maxIdleConns := 5
	connMaxLifetime := 5 * time.Minute

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL database: %s@%s:%s/%s",
		cfg.User, cfg.Host, cfg.Port, cfg.DBName)

	StartMonitoring(db, cfg.DBName)

	return db, nil
}
