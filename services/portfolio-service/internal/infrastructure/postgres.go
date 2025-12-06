package infrastructure

import (
	"database/sql"
	"fmt"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	_ "github.com/lib/pq"
)

// NewPostgresDB creates a new PostgreSQL database connection.
func NewPostgresDB(cfg config.Config) (*sql.DB, error) {
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" || cfg.DBSSLMode == "" {
		return nil, fmt.Errorf("missing required database configuration")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
