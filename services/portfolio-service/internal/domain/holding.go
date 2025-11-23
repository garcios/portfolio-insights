package domain

import "time"

// Holding represents a user's current position in an asset
type Holding struct {
	UserID       string
	Symbol       string
	Quantity     float64
	AverageCost  float64
	CurrentPrice float64 // Enriched from market data
	LastUpdated  time.Time
}

// PortfolioSummary represents the overall portfolio metrics
type PortfolioSummary struct {
	UserID      string
	TotalValue  float64
	TotalCost   float64
	GainLoss    float64
	GainLossPct float64
	LastUpdated time.Time
}

// HoldingRepository defines the interface for holding persistence
type HoldingRepository interface {
	Upsert(holding *Holding) error
	GetByUserAndSymbol(userID, symbol string) (*Holding, error)
	ListByUser(userID string) ([]*Holding, error)
}
