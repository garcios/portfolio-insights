package entity

import "time"

// Portfolio represents a user's investment portfolio
type Portfolio struct {
	ID       string
	UserID   string
	Name     string
	Summary  *PortfolioSummary
	Holdings []*Holding
}

// PortfolioSummary contains aggregated portfolio metrics
type PortfolioSummary struct {
	TotalValue              float64
	TotalGainLoss           float64
	TotalGainLossPercentage float64
	DayChange               float64
	DayChangePercentage     float64
	Currency                string
	LastUpdated             time.Time
	StartDate               *time.Time
	EndDate                 *time.Time
	CapitalGain             float64
	CapitalGainPercentage   float64
	CurrencyGain            float64
	CurrencyGainPercentage  float64
	Dividends               float64
	DividendsPercentage     float64
}

// Holding represents a single asset holding in the portfolio
type Holding struct {
	Symbol             string
	Quantity           float64
	AveragePrice       float64
	CurrentPrice       float64
	CurrentValue       float64
	GainLoss           float64
	GainLossPercentage float64
	Currency           string
	AssetName          string
	UserID             string
}

// PortfolioPerformancePoint represents a single data point in portfolio performance history
type PortfolioPerformancePoint struct {
	Timestamp time.Time
	Value     float64
}

// NewPortfolio creates a new Portfolio entity
func NewPortfolio(id, userID, name string) *Portfolio {
	return &Portfolio{
		ID:       id,
		UserID:   userID,
		Name:     name,
		Holdings: make([]*Holding, 0),
	}
}
