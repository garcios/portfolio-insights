package gateway

import (
	"context"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

// CreateTransactionInput contains the data needed to create a transaction
type CreateTransactionInput struct {
	UserID            string
	Symbol            string
	Type              entity.TransactionType
	Quantity          float64
	PricePerShare     float64
	PriceCurrency     string
	ExecutedAt        time.Time
	Notes             string
	Brokerage         float64
	BrokerageCurrency string
}

// TransactionGateway defines the interface for interacting with the transaction service
type TransactionGateway interface {
	// CreateTransaction creates a new transaction
	CreateTransaction(ctx context.Context, input CreateTransactionInput) (*entity.Transaction, error)
}
