package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
)

// CreateTransactionInput contains the data needed to create a transaction
type CreateTransactionInput struct {
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

// TransactionUseCase handles transaction-related business logic
type TransactionUseCase struct {
	transactionGateway     gateway.TransactionGateway
	transactionFileGateway gateway.TransactionFileGateway
}

// NewTransactionUseCase creates a new TransactionUseCase
func NewTransactionUseCase(
	transactionGateway gateway.TransactionGateway,
	transactionFileGateway gateway.TransactionFileGateway,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionGateway:     transactionGateway,
		transactionFileGateway: transactionFileGateway,
	}
}

// CreateTransaction creates a new transaction for a user
func (uc *TransactionUseCase) CreateTransaction(ctx context.Context, userID string, input CreateTransactionInput) (*entity.Transaction, error) {
	// Validate input
	if err := uc.validateCreateTransactionInput(input); err != nil {
		return nil, fmt.Errorf("invalid transaction input: %w", err)
	}

	// Create gateway input
	gatewayInput := gateway.CreateTransactionInput{
		UserID:            userID,
		Symbol:            input.Symbol,
		Type:              input.Type,
		Quantity:          input.Quantity,
		PricePerShare:     input.PricePerShare,
		PriceCurrency:     input.PriceCurrency,
		ExecutedAt:        input.ExecutedAt,
		Notes:             input.Notes,
		Brokerage:         input.Brokerage,
		BrokerageCurrency: input.BrokerageCurrency,
	}

	return uc.transactionGateway.CreateTransaction(ctx, gatewayInput)
}

// UploadCSV uploads a CSV file for processing
func (uc *TransactionUseCase) UploadCSV(ctx context.Context, userID string, file io.Reader, filename string) error {
	if file == nil {
		return fmt.Errorf("file is required")
	}
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	return uc.transactionFileGateway.UploadCSV(ctx, userID, file, filename)
}

// validateCreateTransactionInput validates the transaction input
func (uc *TransactionUseCase) validateCreateTransactionInput(input CreateTransactionInput) error {
	if input.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if input.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	if input.PricePerShare < 0 {
		return fmt.Errorf("price per share cannot be negative")
	}

	if input.ExecutedAt.IsZero() {
		return fmt.Errorf("executed at is required")
	}

	if input.ExecutedAt.After(time.Now()) {
		return fmt.Errorf("executed at cannot be in the future")
	}

	// Validate transaction type
	validTypes := map[entity.TransactionType]bool{
		entity.TransactionTypeBuy:      true,
		entity.TransactionTypeSell:     true,
		entity.TransactionTypeSplit:    true,
		entity.TransactionTypeDividend: true,
	}

	if !validTypes[input.Type] {
		return fmt.Errorf("invalid transaction type: %s", input.Type)
	}

	return nil
}

// ListTransactions lists transactions for a user with optional filtering and pagination
func (uc *TransactionUseCase) ListTransactions(ctx context.Context, userID string, pageSize int32, pageToken string, filter *gateway.TransactionFilter) (*gateway.ListTransactionsResult, error) {
	input := gateway.ListTransactionsInput{
		UserID:    userID,
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
	}

	return uc.transactionGateway.ListTransactions(ctx, input)
}
