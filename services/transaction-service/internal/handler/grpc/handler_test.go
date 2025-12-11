package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
)

type MockTransactionUsecase struct {
	transactions map[string]*domain.Transaction
}

func (m *MockTransactionUsecase) CreateTransaction(ctx context.Context, txn *domain.Transaction) error {
	if txn.UserID == "invalid-user" {
		return errors.New("user not found")
	}
	txn.ID = "test-id"
	txn.CreatedAt = time.Now()
	txn.UpdatedAt = time.Now()
	m.transactions[txn.ID] = txn
	return nil
}

func (m *MockTransactionUsecase) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	if tx, ok := m.transactions[id]; ok {
		return tx, nil
	}
	return nil, errors.New("not found")
}

func (m *MockTransactionUsecase) ListTransactions(ctx context.Context, userID string, filter domain.TransactionFilter, limit, offset int) ([]*domain.Transaction, error) {
	return nil, nil
}

func (m *MockTransactionUsecase) UpdateTransaction(ctx context.Context, txn *domain.Transaction) error {
	if _, ok := m.transactions[txn.ID]; !ok {
		return errors.New("not found")
	}
	txn.UpdatedAt = time.Now()
	m.transactions[txn.ID] = txn
	return nil
}

func (m *MockTransactionUsecase) DeleteTransaction(ctx context.Context, id string) error {
	return nil
}

func (m *MockTransactionUsecase) GetOldestTransaction(ctx context.Context, userID string) (*domain.Transaction, error) {
	if userID == "user-1" {
		return &domain.Transaction{
			UserID:     userID,
			Symbol:     domain.StringPtr("TEST"),
			ExecutedAt: time.Now(),
		}, nil
	}
	return nil, nil
}

func TestCreateTransactionHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	t.Run("Success", func(t *testing.T) {
		symbol := "AAPL"
		quantity := 10.0
		pricePerShare := 150.0
		req := &pb.CreateTransactionRequest{
			UserId:        "user-1",
			Symbol:        &symbol,
			Type:          "BUY",
			Quantity:      &quantity,
			PricePerShare: &pricePerShare,
		}
		resp, err := handler.CreateTransaction(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Transaction.Id != "test-id" {
			t.Errorf("expected test-id, got %s", resp.Transaction.Id)
		}
	})

	t.Run("InvalidArgs", func(t *testing.T) {
		req := &pb.CreateTransactionRequest{
			UserId: "", // Invalid
		}
		_, err := handler.CreateTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestListTransactionsHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	t.Run("Success", func(t *testing.T) {
		req := &pb.ListTransactionsRequest{
			UserId: "user-1",
			Filter: &pb.TransactionFilter{
				Symbol: "AAPL",
			},
		}
		_, err := handler.ListTransactions(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
