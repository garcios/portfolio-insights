// Package usecase implements the business logic for the transaction service.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/metrics"
	"github.com/google/uuid"
)

type transactionUsecase struct {
	repo              domain.TransactionRepository
	userGateway       domain.UserGateway
	marketDataGateway domain.MarketDataGateway
	eventPublisher    domain.EventPublisher
}

// NewTransactionUsecase creates a new transaction usecase.
func NewTransactionUsecase(repo domain.TransactionRepository, userGateway domain.UserGateway, marketDataGateway domain.MarketDataGateway, eventPublisher domain.EventPublisher) domain.TransactionUsecase {
	return &transactionUsecase{
		repo:              repo,
		userGateway:       userGateway,
		marketDataGateway: marketDataGateway,
		eventPublisher:    eventPublisher,
	}
}

func (uc *transactionUsecase) CreateTransaction(ctx context.Context, txn *domain.Transaction) error {
	start := time.Now()
	defer func() {
		metrics.TransactionProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	// Validate input
	if txn.Type != "BUY" && txn.Type != "SELL" {
		return fmt.Errorf("invalid transaction type: %s", txn.Type)
	}
	if txn.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if txn.PricePerShare < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	if txn.PriceCurrency != "" && len(txn.PriceCurrency) != 3 {
		return fmt.Errorf("price_currency must be a 3-letter code")
	}
	if txn.BrokerageCurrency != "" && len(txn.BrokerageCurrency) != 3 {
		return fmt.Errorf("brokerage_currency must be a 3-letter code")
	}

	// Validate User
	userValidationStart := time.Now()
	exists, err := uc.userGateway.Exists(ctx, txn.UserID)
	metrics.UserValidationDuration.Observe(time.Since(userValidationStart).Seconds())
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found: %s", txn.UserID)
	}

	// Validate Asset
	assetValidationStart := time.Now()
	exists, err = uc.marketDataGateway.Exists(ctx, txn.Symbol)
	metrics.AssetValidationDuration.Observe(time.Since(assetValidationStart).Seconds())
	if err != nil {
		return fmt.Errorf("failed to validate asset: %w", err)
	}
	if !exists {
		return fmt.Errorf("asset %s does not exist", txn.Symbol)
	}

	// Set timestamps and ID if not present
	if txn.ID == "" {
		txn.ID = uuid.New().String()
	}
	now := time.Now()
	if txn.CreatedAt.IsZero() {
		txn.CreatedAt = now
	}
	if txn.UpdatedAt.IsZero() {
		txn.UpdatedAt = now
	}

	if err := uc.repo.Create(ctx, txn); err != nil {
		return err
	}

	// Record business metrics
	metrics.TransactionsCreatedTotal.WithLabelValues(txn.Type).Inc()
	metrics.TransactionValueTotal.WithLabelValues(txn.Type).Add(txn.Quantity * txn.PricePerShare)

	// Publish transaction created event
	if err := uc.eventPublisher.PublishTransactionCreated(ctx, txn); err != nil {
		// Log the error but don't fail the transaction creation
		fmt.Printf("failed to publish transaction created event: %v\n", err)
	}

	return nil
}

func (uc *transactionUsecase) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *transactionUsecase) ListTransactions(ctx context.Context, userID string, filter domain.TransactionFilter, limit, offset int) ([]*domain.Transaction, error) {
	return uc.repo.ListByUserID(ctx, userID, filter, limit, offset)
}

func (uc *transactionUsecase) UpdateTransaction(ctx context.Context, txn *domain.Transaction) error {
	existing, err := uc.repo.GetByID(ctx, txn.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("transaction not found: %s", txn.ID)
	}

	// Update fields
	existing.Symbol = txn.Symbol
	existing.Type = txn.Type
	existing.Quantity = txn.Quantity
	existing.PricePerShare = txn.PricePerShare
	existing.ExecutedAt = txn.ExecutedAt
	existing.Brokerage = txn.Brokerage
	existing.Notes = txn.Notes
	existing.PriceCurrency = txn.PriceCurrency
	existing.BrokerageCurrency = txn.BrokerageCurrency
	existing.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	// Update the input struct to reflect the updated state
	*txn = *existing

	return nil
}

func (uc *transactionUsecase) DeleteTransaction(ctx context.Context, id string) error {
	// TODO: Publish transaction deleted event
	return uc.repo.Delete(ctx, id)
}

func (uc *transactionUsecase) GetOldestTransaction(ctx context.Context, userID string) (*domain.Transaction, error) {
	return uc.repo.GetOldestByUserID(ctx, userID)
}
