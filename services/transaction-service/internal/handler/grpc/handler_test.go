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

func (m *MockTransactionUsecase) CreateTransaction(ctx context.Context, userID, symbol, txType string, quantity, price float64, executedAt time.Time) (*domain.Transaction, error) {
	if userID == "invalid-user" {
		return nil, errors.New("user not found")
	}
	return &domain.Transaction{
		ID:            "test-id",
		UserID:        userID,
		Symbol:        symbol,
		Type:          txType,
		Quantity:      quantity,
		PricePerShare: price,
		ExecutedAt:    executedAt,
	}, nil
}

func (m *MockTransactionUsecase) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	if tx, ok := m.transactions[id]; ok {
		return tx, nil
	}
	return nil, errors.New("not found")
}

func (m *MockTransactionUsecase) ListTransactions(ctx context.Context, userID string, pageSize int, pageToken string) ([]*domain.Transaction, string, error) {
	return nil, "", nil
}

func (m *MockTransactionUsecase) UpdateTransaction(ctx context.Context, id, symbol, txType string, quantity, price float64, executedAt time.Time) (*domain.Transaction, error) {
	return nil, nil
}

func (m *MockTransactionUsecase) DeleteTransaction(ctx context.Context, id string) error {
	return nil
}

func TestCreateTransactionHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	t.Run("Success", func(t *testing.T) {
		req := &pb.CreateTransactionRequest{
			UserId:        "user-1",
			Symbol:        "AAPL",
			Type:          "BUY",
			Quantity:      10,
			PricePerShare: 150.0,
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
