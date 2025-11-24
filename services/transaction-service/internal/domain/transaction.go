package domain

import (
	"context"
	"time"
)

type Transaction struct {
	ID            string
	UserID        string
	Symbol        string
	Type          string // BUY or SELL
	Quantity      float64
	PricePerShare float64
	ExecutedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	BulkCreate(ctx context.Context, transactions []*Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id string) error
}

type UserGateway interface {
	Exists(ctx context.Context, userID string) (bool, error)
}

type MarketDataGateway interface {
	Exists(ctx context.Context, symbol string) (bool, error)
}

type EventPublisher interface {
	PublishTransactionCreated(ctx context.Context, transaction *Transaction) error
}

type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, userID, symbol, txType string, quantity, price float64, executedAt time.Time) (*Transaction, error)
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	ListTransactions(ctx context.Context, userID string, pageSize int, pageToken string) ([]*Transaction, string, error)
	UpdateTransaction(ctx context.Context, id, symbol, txType string, quantity, price float64, executedAt time.Time) (*Transaction, error)
	DeleteTransaction(ctx context.Context, id string) error
}
