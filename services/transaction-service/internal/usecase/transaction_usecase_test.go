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

func (m *MockTransactionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Transaction, error) {
	var result []*domain.Transaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
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
		tx, err := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tx.ID == "" {
			t.Error("expected non-empty ID")
		}
		if tx.Type != "BUY" {
			t.Errorf("expected BUY, got %s", tx.Type)
		}
		if tx.Quantity != 10 {
			t.Errorf("expected quantity 10, got %f", tx.Quantity)
		}
		if len(eventPublisher.published) != 1 {
			t.Errorf("expected 1 published event, got %d", len(eventPublisher.published))
		}
	})

	t.Run("Success_SELL", func(t *testing.T) {
		tx, err := uc.CreateTransaction(context.Background(), "user-2", "GOOGL", "SELL", 5, 2500.0, time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tx.Type != "SELL" {
			t.Errorf("expected SELL, got %s", tx.Type)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userGateway.exists = false
		_, err := uc.CreateTransaction(context.Background(), "user-2", "AAPL", "BUY", 10, 150.0, time.Now())
		if err == nil {
			t.Error("expected error, got nil")
		}
		userGateway.exists = true
	})

	t.Run("AssetNotFound", func(t *testing.T) {
		marketGateway.exists = false
		_, err := uc.CreateTransaction(context.Background(), "user-1", "INVALID", "BUY", 10, 150.0, time.Now())
		if err == nil {
			t.Error("expected error, got nil")
		}
		marketGateway.exists = true
	})

	t.Run("InvalidTransactionType", func(t *testing.T) {
		_, err := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "INVALID", 10, 150.0, time.Now())
		if err == nil {
			t.Error("expected error for invalid transaction type")
		}
	})

	t.Run("ZeroQuantity", func(t *testing.T) {
		_, err := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 0, 150.0, time.Now())
		if err == nil {
			t.Error("expected error for zero quantity")
		}
	})

	t.Run("NegativePrice", func(t *testing.T) {
		_, err := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, -150.0, time.Now())
		if err == nil {
			t.Error("expected error for negative price")
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
	tx, _ := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())

	t.Run("Success", func(t *testing.T) {
		found, err := uc.GetTransaction(context.Background(), tx.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if found.ID != tx.ID {
			t.Errorf("expected ID %s, got %s", tx.ID, found.ID)
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
	uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())
	uc.CreateTransaction(context.Background(), "user-1", "GOOGL", "BUY", 5, 2500.0, time.Now())
	uc.CreateTransaction(context.Background(), "user-2", "MSFT", "BUY", 20, 300.0, time.Now())

	t.Run("FilterByUser", func(t *testing.T) {
		txs, _, err := uc.ListTransactions(context.Background(), "user-1", 10, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 2 {
			t.Errorf("expected 2 transactions for user-1, got %d", len(txs))
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		txs, _, err := uc.ListTransactions(context.Background(), "user-1", 1, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 1 {
			t.Errorf("expected 1 transaction with limit 1, got %d", len(txs))
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		txs, _, err := uc.ListTransactions(context.Background(), "user-999", 10, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(txs) != 0 {
			t.Errorf("expected 0 transactions for non-existent user, got %d", len(txs))
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
	tx, _ := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())

	t.Run("Success", func(t *testing.T) {
		updated, err := uc.UpdateTransaction(context.Background(), tx.ID, "AAPL", "BUY", 15, 155.0, time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updated.Quantity != 15 {
			t.Errorf("expected quantity 15, got %f", updated.Quantity)
		}
		if updated.PricePerShare != 155.0 {
			t.Errorf("expected price 155.0, got %f", updated.PricePerShare)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.UpdateTransaction(context.Background(), "invalid-id", "AAPL", "BUY", 10, 150.0, time.Now())
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
	tx, _ := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())

	t.Run("Success", func(t *testing.T) {
		err := uc.DeleteTransaction(context.Background(), tx.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = uc.GetTransaction(context.Background(), tx.ID)
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
