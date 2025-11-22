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
}

func NewMockRepo() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[string]*domain.Transaction),
	}
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	transaction.ID = "test-id"
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
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
	return result, nil
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id string) error {
	delete(m.transactions, id)
	return nil
}

// MockUserGateway
type MockUserGateway struct {
	exists bool
}

func (m *MockUserGateway) Exists(ctx context.Context, userID string) (bool, error) {
	return m.exists, nil
}

// MockMarketDataGateway
type MockMarketDataGateway struct {
	exists bool
}

func (m *MockMarketDataGateway) Exists(ctx context.Context, symbol string) (bool, error) {
	return m.exists, nil
}

// MockEventPublisher
type MockEventPublisher struct{}

func (m *MockEventPublisher) PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error {
	return nil
}

func TestCreateTransaction(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Success", func(t *testing.T) {
		tx, err := uc.CreateTransaction(context.Background(), "user-1", "AAPL", "BUY", 10, 150.0, time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tx.ID != "test-id" {
			t.Errorf("expected test-id, got %s", tx.ID)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userGateway.exists = false
		_, err := uc.CreateTransaction(context.Background(), "user-2", "AAPL", "BUY", 10, 150.0, time.Now())
		if err == nil {
			t.Error("expected error, got nil")
		}
		userGateway.exists = true // Reset
	})

	t.Run("AssetNotFound", func(t *testing.T) {
		marketGateway.exists = false
		_, err := uc.CreateTransaction(context.Background(), "user-1", "INVALID", "BUY", 10, 150.0, time.Now())
		if err == nil {
			t.Error("expected error, got nil")
		}
		marketGateway.exists = true // Reset
	})
}
