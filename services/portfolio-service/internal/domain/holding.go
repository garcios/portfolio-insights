package domain

import "time"

// Holding represents a user's current position in an asset
type Holding struct {
	UserID      string
	Symbol      string
	Quantity    float64
	AverageCost float64
	LastUpdated time.Time
}

// HoldingRepository defines the interface for holding persistence
type HoldingRepository interface {
	Upsert(holding *Holding) error
	GetByUserAndSymbol(userID, symbol string) (*Holding, error)
	ListByUser(userID string) ([]*Holding, error)
}
