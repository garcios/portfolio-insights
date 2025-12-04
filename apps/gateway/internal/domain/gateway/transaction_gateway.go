package gateway

import (
	"context"
	"io"
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

// TransactionFilter contains optional filters for listing transactions
type TransactionFilter struct {
	Symbol         *string
	Type           *entity.TransactionType
	FromExecutedAt *time.Time
	ToExecutedAt   *time.Time
}

// ListTransactionsInput contains the data needed to list transactions
type ListTransactionsInput struct {
	UserID    string
	PageSize  int32
	PageToken string
	Filter    *TransactionFilter
}

// ListTransactionsResult contains the result of listing transactions
type ListTransactionsResult struct {
	Transactions  []*entity.Transaction
	NextPageToken string
}

// TransactionGateway defines the interface for interacting with the transaction service
type TransactionGateway interface {
	// CreateTransaction creates a new transaction
	CreateTransaction(ctx context.Context, input CreateTransactionInput) (*entity.Transaction, error)

	// ListTransactions lists transactions for a user with optional filtering and pagination
	ListTransactions(ctx context.Context, input ListTransactionsInput) (*ListTransactionsResult, error)
}

// TransactionFileGateway defines the interface for uploading transaction files
type TransactionFileGateway interface {
	// UploadCSV uploads a CSV file for processing
	UploadCSV(ctx context.Context, userID string, file io.Reader, filename string) error
}
