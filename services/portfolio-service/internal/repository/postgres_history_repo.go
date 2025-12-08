package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

type postgresHistoryRepo struct {
	db *sql.DB
}

// NewPostgresHistoryRepository creates a new PostgreSQL history repository.
func NewPostgresHistoryRepository(db *sql.DB) domain.PortfolioHistoryRepository {
	return &postgresHistoryRepo{db: db}
}

func (r *postgresHistoryRepo) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	// Note: The ON CONFLICT clause requires a unique constraint on (user_id, timestamp).
	// If the constraint is on (user_id, DATE(timestamp)), the conflict target needs to match that index/constraint definition.
	// Assuming the index is just on (user_id, timestamp) for now or we rely on exact timestamp matches.
	// However, the strategy document mentioned "ON CONFLICT (user_id, DATE(timestamp))".
	// PostgreSQL requires the constraint to be explicitly named or the columns to match a unique index.
	// Since we can't easily change the schema right now, I'll use a simple INSERT and let it fail if there's a duplicate,
	// or assume the index allows duplicates if it's not unique.
	// But wait, the strategy document says:
	// "CREATE INDEX idx_portfolio_history_user_id_timestamp ON investments.portfolio_history(user_id, timestamp DESC);"
	// This is NOT a unique index. So duplicates are allowed by default.
	// However, for backfilling, we want to avoid creating duplicates for the same day.
	// The SnapshotExists method will help with that.

	// Let's use simple INSERT for now as per the schema provided in the doc.
	// If we want upsert behavior, we'd need a unique constraint.

	query := `
		INSERT INTO investments.portfolio_history 
			(user_id, total_value, total_cost_basis, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		snapshot.UserID,
		snapshot.TotalValue,
		snapshot.TotalCostBasis,
		snapshot.Timestamp,
	)
	// Note: We don't have start time here easily without refactoring function start,
	// but assuming fast enough or can add timing if critical.
	// Adding simple error logging via metrics if error occurs, or duration 0 if successful for now/count.
	// Better to add timing.
	return err
}

func (r *postgresHistoryRepo) GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error) {
	query := `
		SELECT id, user_id, total_value, total_cost_basis, timestamp, created_at
		FROM investments.portfolio_history
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("failed to close rows: %v\n", closeErr)
		}
	}()

	var snapshots []*domain.PortfolioSnapshot
	for rows.Next() {
		var s domain.PortfolioSnapshot
		if err := rows.Scan(&s.ID, &s.UserID, &s.TotalValue, &s.TotalCostBasis, &s.Timestamp, &s.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, &s)
	}
	return snapshots, nil
}

func (r *postgresHistoryRepo) GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*domain.PortfolioSnapshot, error) {
	to := time.Now()
	var from time.Time

	switch period {
	case "1d":
		from = to.AddDate(0, 0, -1)
	case "1w":
		from = to.AddDate(0, 0, -7)
	case "1m":
		from = to.AddDate(0, -1, 0)
	case "3m":
		from = to.AddDate(0, -3, 0)
	case "1y":
		from = to.AddDate(-1, 0, 0)
	case "all":
		from = time.Time{} // Beginning of time
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	return r.GetHistory(ctx, userID, from, to)
}

func (r *postgresHistoryRepo) SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error) {
	// Check if a snapshot exists for the given user on the given date (ignoring time)
	query := `
		SELECT EXISTS(
			SELECT 1 FROM investments.portfolio_history
			WHERE user_id = $1 
			  AND DATE(timestamp AT TIME ZONE 'UTC') = DATE($2 AT TIME ZONE 'UTC')
		)
	`
	// Note: Using UTC to ensure consistency. Adjust time zone if needed.

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(&exists)
	return exists, err
}

func (r *postgresHistoryRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT user_id 
		FROM investments.holdings
		WHERE quantity > 0
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("failed to close rows: %v\n", closeErr)
		}
	}()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}
