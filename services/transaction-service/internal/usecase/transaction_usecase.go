package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/metrics"
)

type transactionUsecase struct {
	repo              domain.TransactionRepository
	userGateway       domain.UserGateway
	marketDataGateway domain.MarketDataGateway
	eventPublisher    domain.EventPublisher
}

func NewTransactionUsecase(repo domain.TransactionRepository, userGateway domain.UserGateway, marketDataGateway domain.MarketDataGateway, eventPublisher domain.EventPublisher) domain.TransactionUsecase {
	return &transactionUsecase{
		repo:              repo,
		userGateway:       userGateway,
		marketDataGateway: marketDataGateway,
		eventPublisher:    eventPublisher,
	}
}

func (uc *transactionUsecase) CreateTransaction(ctx context.Context, userID, symbol, txType string, quantity, price float64, executedAt time.Time) (*domain.Transaction, error) {
	start := time.Now()
	defer func() {
		metrics.TransactionProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	// Validate input
	if txType != "BUY" && txType != "SELL" {
		return nil, fmt.Errorf("invalid transaction type: %s", txType)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	// Validate User
	userValidationStart := time.Now()
	exists, err := uc.userGateway.Exists(ctx, userID)
	metrics.UserValidationDuration.Observe(time.Since(userValidationStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user %s does not exist", userID)
	}

	// Validate Asset
	assetValidationStart := time.Now()
	exists, err = uc.marketDataGateway.Exists(ctx, symbol)
	metrics.AssetValidationDuration.Observe(time.Since(assetValidationStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to validate asset: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("asset %s does not exist", symbol)
	}

	tx := &domain.Transaction{
		UserID:        userID,
		Symbol:        symbol,
		Type:          txType,
		Quantity:      quantity,
		PricePerShare: price,
		ExecutedAt:    executedAt,
	}
	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// Record business metrics
	metrics.TransactionsCreatedTotal.WithLabelValues(txType).Inc()
	metrics.TransactionValueTotal.WithLabelValues(txType).Add(quantity * price)

	// Publish transaction created event
	if err := uc.eventPublisher.PublishTransactionCreated(ctx, tx); err != nil {
		// Log the error but don't fail the transaction creation
		// In production, you might want to use a more sophisticated error handling strategy
		fmt.Printf("failed to publish transaction created event: %v\n", err)
	}

	return tx, nil
}

func (uc *transactionUsecase) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *transactionUsecase) ListTransactions(ctx context.Context, userID string, pageSize int, pageToken string) ([]*domain.Transaction, string, error) {
	limit := pageSize
	offset := 0
	if pageToken != "" {
		var err error
		offset, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", err
		}
	}

	transactions, err := uc.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, "", err
	}

	nextPageToken := ""
	if len(transactions) == limit {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return transactions, nextPageToken, nil
}

func (uc *transactionUsecase) UpdateTransaction(ctx context.Context, id, symbol, txType string, quantity, price float64, executedAt time.Time) (*domain.Transaction, error) {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx.Symbol = symbol
	tx.Type = txType
	tx.Quantity = quantity
	tx.PricePerShare = price
	tx.ExecutedAt = executedAt

	if err := uc.repo.Update(ctx, tx); err != nil {
		return nil, err
	}

	// TODO: Publish transaction updated event

	return tx, nil
}

func (uc *transactionUsecase) DeleteTransaction(ctx context.Context, id string) error {

	// TODO: Publish transaction deleted event

	return uc.repo.Delete(ctx, id)
}
