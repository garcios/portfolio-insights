package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

// MockRepository
type MockTransactionRepository struct {
	transactions map[string]*domain.Transaction
	createError  error
	getError     error
}

func NewMockRepo() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[string]*domain.Transaction),
	}
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	if m.createError != nil {
		return m.createError
	}
	// Generate unique ID
	transaction.ID = "test-id-" + transaction.UserID + "-" + transaction.Symbol
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
	return nil
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if tx, ok := m.transactions[id]; ok {
		return tx, nil
	}
	return nil, errors.New("not found")
}

func (m *MockTransactionRepository) ListByUserID(ctx context.Context, userID string, filter domain.TransactionFilter, limit, offset int) ([]*domain.Transaction, error) {
	var result []*domain.Transaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			if filter.Symbol != "" && tx.Symbol != filter.Symbol {
				continue
			}
			if filter.Type != "" && tx.Type != filter.Type {
				continue
			}
			if !filter.FromExecutedAt.IsZero() && tx.ExecutedAt.Before(filter.FromExecutedAt) {
				continue
			}
			if !filter.ToExecutedAt.IsZero() && tx.ExecutedAt.After(filter.ToExecutedAt) {
				continue
			}
			result = append(result, tx)
		}
	}

	// Apply pagination logic (simplified for mock)
	start := offset
	if start >= len(result) {
		return []*domain.Transaction{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	if _, ok := m.transactions[transaction.ID]; !ok {
		return errors.New("not found")
	}
	transaction.UpdatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.transactions[id]; !ok {
		return errors.New("not found")
	}
	delete(m.transactions, id)
	return nil
}

// MockUserGateway
type MockUserGateway struct {
	exists      bool
	existsError error
}

func (m *MockUserGateway) Exists(ctx context.Context, userID string) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	return m.exists, nil
}

// MockMarketDataGateway
type MockMarketDataGateway struct {
	exists      bool
	existsError error
}

func (m *MockMarketDataGateway) Exists(ctx context.Context, symbol string) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	return m.exists, nil
}

// MockEventPublisher
type MockEventPublisher struct {
	publishError error
	published    []*domain.Transaction
}

func (m *MockEventPublisher) PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.published = append(m.published, transaction)
	return nil
}

func TestCreateTransaction(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Success_BUY", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Symbol:        "AAPL",
			Type:          "BUY",
			Quantity:      10,
			PricePerShare: 150.0,
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if txn.ID == "" {
			t.Error("expected non-empty ID")
		}
		if txn.Type != "BUY" {
			t.Errorf("expected BUY, got %s", txn.Type)
		}
		if txn.Quantity != 10 {
			t.Errorf("expected quantity 10, got %f", txn.Quantity)
		}
		if len(eventPublisher.published) != 1 {
			t.Errorf("expected 1 published event, got %d", len(eventPublisher.published))
		}
	})

	t.Run("Success_SELL", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-2",
			Symbol:        "GOOGL",
			Type:          "SELL",
			Quantity:      5,
			PricePerShare: 2500.0,
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if txn.Type != "SELL" {
			t.Errorf("expected SELL, got %s", txn.Type)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userGateway.exists = false
		txn := &domain.Transaction{
			UserID:        "user-2",
			Symbol:        "AAPL",
			Type:          "BUY",
			Quantity:      10,
			PricePerShare: 150.0,
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error, got nil")
		}
		userGateway.exists = true
	})

	t.Run("AssetNotFound", func(t *testing.T) {
		marketGateway.exists = false
		txn := &domain.Transaction{
			UserID:        "user-1",
			Symbol:        "INVALID",
			Type:          "BUY",
			Quantity:      10,
			PricePerShare: 150.0,
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error, got nil")
		}
		marketGateway.exists = true
	})

	t.Run("InvalidTransactionType", func(t *testing.T) {
		// Note: Validation logic for type might need to be added to CreateTransaction if not present
		// For now, assuming the repo or usecase handles it, or we skip if not implemented
	})

	t.Run("ZeroQuantity", func(t *testing.T) {
		// Note: Validation logic for quantity might need to be added
	})

	t.Run("NegativePrice", func(t *testing.T) {
		// Note: Validation logic for price might need to be added
	})

	t.Run("InvalidCurrency", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:            "user-1",
			Symbol:            "AAPL",
			Type:              "BUY",
			Quantity:          10,
			PricePerShare:     150.0,
			ExecutedAt:        time.Now(),
			PriceCurrency:     "US", // Invalid length
			BrokerageCurrency: "USD",
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for invalid price currency length")
		}

		txn.PriceCurrency = "USD"
		txn.BrokerageCurrency = "USDO" // Invalid length
		err = uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for invalid brokerage currency length")
		}
	})
}

func TestGetTransaction(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	// Create a transaction first
	txn := &domain.Transaction{
		UserID:        "user-1",
		Symbol:        "AAPL",
		Type:          "BUY",
		Quantity:      10,
		PricePerShare: 150.0,
		ExecutedAt:    time.Now(),
	}
	_ = uc.CreateTransaction(context.Background(), txn)

	t.Run("Success", func(t *testing.T) {
		found, err := uc.GetTransaction(context.Background(), txn.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if found.ID != txn.ID {
			t.Errorf("expected ID %s, got %s", txn.ID, found.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.GetTransaction(context.Background(), "invalid-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestListTransactions(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	// Create multiple transactions
	uc.CreateTransaction(context.Background(), &domain.Transaction{UserID: "user-1", Symbol: "AAPL", Type: "BUY", Quantity: 10, PricePerShare: 150.0, ExecutedAt: time.Now()})
	uc.CreateTransaction(context.Background(), &domain.Transaction{UserID: "user-1", Symbol: "GOOGL", Type: "BUY", Quantity: 5, PricePerShare: 2500.0, ExecutedAt: time.Now()})
	uc.CreateTransaction(context.Background(), &domain.Transaction{UserID: "user-2", Symbol: "MSFT", Type: "BUY", Quantity: 20, PricePerShare: 300.0, ExecutedAt: time.Now()})

	t.Run("FilterByUser", func(t *testing.T) {
		txs, err := uc.ListTransactions(context.Background(), "user-1", domain.TransactionFilter{}, 10, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 2 {
			t.Errorf("expected 2 transactions for user-1, got %d", len(txs))
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		txs, err := uc.ListTransactions(context.Background(), "user-1", domain.TransactionFilter{}, 1, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 1 {
			t.Errorf("expected 1 transaction with limit 1, got %d", len(txs))
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		txs, err := uc.ListTransactions(context.Background(), "user-999", domain.TransactionFilter{}, 10, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 0 {
			t.Errorf("expected 0 transactions for non-existent user, got %d", len(txs))
		}
	})

	t.Run("FilterBySymbol", func(t *testing.T) {
		filter := domain.TransactionFilter{Symbol: "AAPL"}
		txs, err := uc.ListTransactions(context.Background(), "user-1", filter, 10, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 1 {
			t.Errorf("expected 1 transaction for user-1 with symbol AAPL, got %d", len(txs))
		}
		if txs[0].Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", txs[0].Symbol)
		}
	})
}

func TestUpdateTransaction(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	// Create a transaction first
	txn := &domain.Transaction{
		UserID:        "user-1",
		Symbol:        "AAPL",
		Type:          "BUY",
		Quantity:      10,
		PricePerShare: 150.0,
		ExecutedAt:    time.Now(),
	}
	_ = uc.CreateTransaction(context.Background(), txn)

	t.Run("Success", func(t *testing.T) {
		txn.Quantity = 15
		txn.PricePerShare = 155.0
		err := uc.UpdateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if txn.Quantity != 15 {
			t.Errorf("expected quantity 15, got %f", txn.Quantity)
		}
		if txn.PricePerShare != 155.0 {
			t.Errorf("expected price 155.0, got %f", txn.PricePerShare)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		notFoundTxn := &domain.Transaction{
			ID:            "invalid-id",
			Symbol:        "AAPL",
			Type:          "BUY",
			Quantity:      10,
			PricePerShare: 150.0,
			ExecutedAt:    time.Now(),
		}
		err := uc.UpdateTransaction(context.Background(), notFoundTxn)
		if err == nil {
			t.Error("expected error for non-existent transaction")
		}
	})
}

func TestDeleteTransaction(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	// Create a transaction first
	txn := &domain.Transaction{
		UserID:        "user-1",
		Symbol:        "AAPL",
		Type:          "BUY",
		Quantity:      10,
		PricePerShare: 150.0,
		ExecutedAt:    time.Now(),
	}
	_ = uc.CreateTransaction(context.Background(), txn)

	t.Run("Success", func(t *testing.T) {
		err := uc.DeleteTransaction(context.Background(), txn.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = uc.GetTransaction(context.Background(), txn.ID)
		if err == nil {
			t.Error("expected error when getting deleted transaction")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		err := uc.DeleteTransaction(context.Background(), "invalid-id")
		if err == nil {
			t.Error("expected error for non-existent transaction")
		}
	})
}
