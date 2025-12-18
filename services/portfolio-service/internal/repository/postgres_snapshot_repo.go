package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

type postgresSnapshotRepo struct {
	db *sql.DB
}

// NewPostgresSnapshotRepository creates a new PostgreSQL snapshot repository.
func NewPostgresSnapshotRepository(db *sql.DB) domain.DetailedSnapshotRepository {
	return &postgresSnapshotRepo{db: db}
}

func (r *postgresSnapshotRepo) GetLatestSnapshot(ctx context.Context, userID string, before time.Time) (*domain.PortfolioSnapshot, error) {
	query := `
		SELECT id, user_id, timestamp, holdings_snapshot, cash_snapshot, realized_gains_snapshot, net_invested, transaction_count, created_at
		FROM investments.portfolio_snapshots
		WHERE user_id = $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var s domain.PortfolioSnapshot
	var holdingsJSON, cashJSON, gainsJSON []byte

	fmt.Printf("GetLatestSnapshot: Querying for userID=%s before=%s\n", userID, before.Format(time.RFC3339Nano))

	err := r.db.QueryRowContext(ctx, query, userID, before).Scan(
		&s.ID,
		&s.UserID,
		&s.Timestamp,
		&holdingsJSON,
		&cashJSON,
		&gainsJSON,
		&s.State.NetInvested, // Scan NetInvested directly into State
		&s.TransactionCount,
		&s.CreatedAt,
	)

	if err == sql.ErrNoRows {
		fmt.Printf("GetLatestSnapshot: No rows found for userID=%s\n", userID)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}

	// Unmarshal JSONB fields
	if len(holdingsJSON) > 0 {
		if err := json.Unmarshal(holdingsJSON, &s.State.Holdings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal holdings: %w", err)
		}
	}
	if len(cashJSON) > 0 {
		if err := json.Unmarshal(cashJSON, &s.State.Cash); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cash: %w", err)
		}
	}
	if len(gainsJSON) > 0 {
		if err := json.Unmarshal(gainsJSON, &s.State.RealizedGains); err != nil {
			return nil, fmt.Errorf("failed to unmarshal realized gains: %w", err)
		}
	}

	return &s, nil
}

func (r *postgresSnapshotRepo) UpsertSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	holdingsJSON, err := json.Marshal(snapshot.State.Holdings)
	if err != nil {
		return fmt.Errorf("failed to marshal holdings: %w", err)
	}
	cashJSON, err := json.Marshal(snapshot.State.Cash)
	if err != nil {
		return fmt.Errorf("failed to marshal cash: %w", err)
	}
	gainsJSON, err := json.Marshal(snapshot.State.RealizedGains)
	if err != nil {
		return fmt.Errorf("failed to marshal gains: %w", err)
	}

	query := `
		INSERT INTO investments.portfolio_snapshots (
			user_id, timestamp, holdings_snapshot, cash_snapshot, realized_gains_snapshot, net_invested, transaction_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, timestamp) DO UPDATE SET
			holdings_snapshot = EXCLUDED.holdings_snapshot,
			cash_snapshot = EXCLUDED.cash_snapshot,
			realized_gains_snapshot = EXCLUDED.realized_gains_snapshot,
			net_invested = EXCLUDED.net_invested,
			transaction_count = EXCLUDED.transaction_count,
			created_at = EXCLUDED.created_at
	`

	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now()
	}

	_, err = r.db.ExecContext(ctx, query,
		snapshot.UserID,
		snapshot.Timestamp,
		holdingsJSON,
		cashJSON,
		gainsJSON,
		snapshot.State.NetInvested, // Pass NetInvested string
		snapshot.TransactionCount,
		snapshot.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert snapshot: %w", err)
	}
	return nil
}

func (r *postgresSnapshotRepo) InvalidateSnapshots(ctx context.Context, userID string, after time.Time) error {
	query := `
		DELETE FROM investments.portfolio_snapshots
		WHERE user_id = $1 AND timestamp >= $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, after)
	if err != nil {
		return fmt.Errorf("failed to invalidate snapshots: %w", err)
	}
	return nil
}
