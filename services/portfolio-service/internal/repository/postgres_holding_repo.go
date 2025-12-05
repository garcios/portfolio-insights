package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
)

// PostgresHoldingRepository implements a PostgreSQL holding repository.
type PostgresHoldingRepository struct {
	db *sql.DB
}

// NewPostgresHoldingRepository creates a new PostgreSQL holding repository.
func NewPostgresHoldingRepository(db *sql.DB) *PostgresHoldingRepository {
	return &PostgresHoldingRepository{db: db}
}

// Upsert inserts or updates a holding
func (r *PostgresHoldingRepository) Upsert(holding *domain.Holding) error {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("upsert", "holdings", time.Since(start).Seconds(), nil)
	}()

	query := `
		INSERT INTO investments.holdings (user_id, symbol, quantity, average_cost_basis, currency, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, symbol)
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			average_cost_basis = EXCLUDED.average_cost_basis,
			currency = EXCLUDED.currency,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(
		query,
		holding.UserID,
		holding.Symbol,
		holding.Quantity,
		holding.AverageCost,
		holding.Currency,
		holding.LastUpdated,
	)

	if err != nil {
		metrics.RecordDatabaseQuery("upsert", "holdings", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to upsert holding: %w", err)
	}

	return nil
}

// GetByUserAndSymbol retrieves a specific holding for a user and symbol
func (r *PostgresHoldingRepository) GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_by_user_symbol", "holdings", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at
		FROM investments.holdings
		WHERE user_id = $1 AND symbol = $2
	`

	holding := &domain.Holding{}
	err := r.db.QueryRow(query, userID, symbol).Scan(
		&holding.UserID,
		&holding.Symbol,
		&holding.Quantity,
		&holding.AverageCost,
		&holding.Currency,
		&holding.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("holding not found")
	}
	if err != nil {
		metrics.RecordDatabaseQuery("get_by_user_symbol", "holdings", time.Since(start).Seconds(), err)
		return nil, fmt.Errorf("failed to get holding: %w", err)
	}

	return holding, nil
}

// ListByUser retrieves all holdings for a specific user
func (r *PostgresHoldingRepository) ListByUser(userID string) ([]*domain.Holding, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("list_by_user", "holdings", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at
		FROM investments.holdings
		WHERE user_id = $1 AND quantity > 0
		ORDER BY symbol ASC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		metrics.RecordDatabaseQuery("list_by_user", "holdings", time.Since(start).Seconds(), err)
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			metrics.RecordDatabaseQuery("list_by_user_close", "holdings", 0, closeErr)
		}
	}()

	var holdings []*domain.Holding
	for rows.Next() {
		holding := &domain.Holding{}
		err := rows.Scan(
			&holding.UserID,
			&holding.Symbol,
			&holding.Quantity,
			&holding.AverageCost,
			&holding.Currency,
			&holding.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, holding)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating holdings: %w", err)
	}

	return holdings, nil
}

// Count returns the total number of holdings (useful for metrics)
func (r *PostgresHoldingRepository) Count() (int, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("count", "holdings", time.Since(start).Seconds(), nil)
	}()

	query := `SELECT COUNT(*) FROM investments.holdings`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		metrics.RecordDatabaseQuery("count", "holdings", time.Since(start).Seconds(), err)
		return 0, fmt.Errorf("failed to count holdings: %w", err)
	}

	return count, nil
}

// DeleteZeroQuantityHoldings removes holdings with zero or negative quantity
func (r *PostgresHoldingRepository) DeleteZeroQuantityHoldings() error {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("delete_zero_quantity", "holdings", time.Since(start).Seconds(), nil)
	}()

	query := `DELETE FROM investments.holdings WHERE quantity <= 0`

	_, err := r.db.Exec(query)
	if err != nil {
		metrics.RecordDatabaseQuery("delete_zero_quantity", "holdings", time.Since(start).Seconds(), err)
		return fmt.Errorf("failed to delete zero quantity holdings: %w", err)
	}

	return nil
}
