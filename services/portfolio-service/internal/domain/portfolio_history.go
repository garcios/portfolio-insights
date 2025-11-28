package domain

import (
	"context"
	"time"
)

// PortfolioSnapshot represents a historical snapshot of a user's portfolio value
type PortfolioSnapshot struct {
	ID             string
	UserID         string
	TotalValue     float64
	TotalCostBasis float64
	Timestamp      time.Time
	CreatedAt      time.Time
}

// PortfolioHistoryRepository defines the interface for portfolio history persistence
type PortfolioHistoryRepository interface {
	CreateSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error
	GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*PortfolioSnapshot, error)
	GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*PortfolioSnapshot, error)
	SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error)
	GetAllUserIDs(ctx context.Context) ([]string, error)
}
