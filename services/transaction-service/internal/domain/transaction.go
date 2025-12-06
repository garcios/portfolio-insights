// Package domain defines the business entities and interfaces for the transaction service.
package domain

import (
	"context"
	"time"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Symbol            string    `json:"symbol"`
	Type              string    `json:"type"` // BUY or SELL
	Quantity          float64   `json:"quantity"`
	PricePerShare     float64   `json:"price_per_share"`
	ExecutedAt        time.Time `json:"executed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Brokerage         float64   `json:"brokerage"`
	Notes             string    `json:"notes"`
	PriceCurrency     string    `json:"price_currency"`
	BrokerageCurrency string    `json:"brokerage_currency"`
}

// TransactionFilter defines filters for listing transactions.
type TransactionFilter struct {
	Symbol         string
	Type           string
	FromExecutedAt time.Time
	ToExecutedAt   time.Time
}

// TransactionRepository defines the interface for transaction storage.
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	BulkCreate(ctx context.Context, transactions []*Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	ListByUserID(ctx context.Context, userID string, filter TransactionFilter, limit, offset int) ([]*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id string) error
	Count() (int, error)
}

// UserGateway defines the interface for communicating with the user service.
type UserGateway interface {
	Exists(ctx context.Context, userID string) (bool, error)
}

// MarketDataGateway defines the interface for communicating with the market data service.
type MarketDataGateway interface {
	Exists(ctx context.Context, symbol string) (bool, error)
}

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	PublishTransactionCreated(ctx context.Context, transaction *Transaction) error
}

// TransactionUsecase defines the interface for transaction business logic.
type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, txn *Transaction) error
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	ListTransactions(ctx context.Context, userID string, filter TransactionFilter, limit, offset int) ([]*Transaction, error)
	UpdateTransaction(ctx context.Context, txn *Transaction) error
	DeleteTransaction(ctx context.Context, id string) error
}
