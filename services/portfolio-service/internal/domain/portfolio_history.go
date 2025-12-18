package domain

import (
	"context"
	"time"
)

// SnapshotState represents the JSONB payload stored in the snapshot.
// We use strings for monetary values to ensure precision during marshalling/unmarshalling.
type SnapshotState struct {
	Holdings      map[string]HoldingState `json:"holdings"`       // Map[Symbol]HoldingState
	Cash          map[string]string       `json:"cash"`           // Map[Currency]Amount
	RealizedGains map[string]string       `json:"realized_gains"` // Map[Currency]Amount
	NetInvested   string                  `json:"net_invested"`   // Total Net Invested (Deposits - Withdrawals)
}

// HoldingState captures the position of a specific asset at the snapshot time.
type HoldingState struct {
	Quantity  string `json:"quantity"`   // precise decimal string
	CostBasis string `json:"cost_basis"` // precise decimal string
	Currency  string `json:"currency"`   // Currency of the asset
}

// PortfolioSnapshot represents a historical snapshot of a user's portfolio value.
// It acts as the aggregate root for both simple history and detailed state checkpoints.
type PortfolioSnapshot struct {
	ID        string
	UserID    string
	Timestamp time.Time
	CreatedAt time.Time

	// Legacy / Simple Summary fields
	TotalValue     float64
	TotalCostBasis float64

	// Detailed State fields (for Incremental Aggregation)
	State            SnapshotState
	TransactionCount int
}

// AssetPosition tracks the cost basis of an asset for realized gain calculation.
// Moved here from portfolio_usecase.go to be shared across domain.
type AssetPosition struct {
	Quantity           float64
	AverageCost        float64 // In Base Currency (Total Cost / Qty) or CostBasis per unit? Usually Avg Cost per Unit.
	AverageForeignCost float64 // In Foreign Currency (Total Foreign Cost / Qty)
	Currency           string
}

// PortfolioHistoryRepository defines the interface for portfolio history persistence (Simple/Legacy).
type PortfolioHistoryRepository interface {
	CreateSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error
	GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*PortfolioSnapshot, error)
	GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*PortfolioSnapshot, error)
	SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error)
	GetAllUserIDs(ctx context.Context) ([]string, error)
}

// DetailedSnapshotRepository defines the interface for managing detailed snapshots (Incremental Aggregation).
type DetailedSnapshotRepository interface {
	// GetLatestSnapshot retrieves the most recent snapshot before or at the given time.
	GetLatestSnapshot(ctx context.Context, userID string, before time.Time) (*PortfolioSnapshot, error)

	// UpsertSnapshot saves a new snapshot.
	UpsertSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error

	// InvalidateSnapshots deletes or marks as stale all snapshots after a certain time (used on write).
	InvalidateSnapshots(ctx context.Context, userID string, after time.Time) error
}
