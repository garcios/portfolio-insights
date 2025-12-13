package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockTransactionUsecase struct {
	transactions map[string]*domain.Transaction
}

func (m *MockTransactionUsecase) CreateTransaction(ctx context.Context, txn *domain.Transaction) error {
	if txn.UserID == "invalid-user" {
		return errors.New("user not found")
	}
	if txn.ID == "" {
		txn.ID = "test-txn-id"
	}
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
	var result []*domain.Transaction
	for _, txn := range m.transactions {
		if txn.UserID == userID {
			result = append(result, txn)
		}
	}
	return result, nil
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
	if _, ok := m.transactions[id]; !ok {
		return errors.New("not found")
	}
	delete(m.transactions, id)
	return nil
}

func (m *MockTransactionUsecase) GetOldestTransaction(ctx context.Context, userID string) (*domain.Transaction, error) {
	if userID == "user-1" {
		return &domain.Transaction{
			ID:         "oldest-txn",
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
			Parent: resourcenames.UserName("user-1"),
			Transaction: &pb.Transaction{
				Symbol:            &symbol,
				Type:              "BUY",
				Quantity:          &quantity,
				PricePerShare:     &pricePerShare,
				ExecutedAt:        timestamppb.Now(),
				PriceCurrency:     "USD",
				BrokerageCurrency: "USD",
			},
		}
		resp, err := handler.CreateTransaction(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.TransactionId != "test-txn-id" {
			t.Errorf("expected test-txn-id, got %s", resp.TransactionId)
		}
		if resp.Name != resourcenames.TransactionName("user-1", "test-txn-id") {
			t.Errorf("expected resource name users/user-1/transactions/test-txn-id, got %s", resp.Name)
		}
	})

	t.Run("InvalidParent", func(t *testing.T) {
		req := &pb.CreateTransactionRequest{
			Parent: "invalid-parent",
			Transaction: &pb.Transaction{
				Type:       "BUY",
				ExecutedAt: timestamppb.Now(),
			},
		}
		_, err := handler.CreateTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid parent, got nil")
		}
	})

	t.Run("NilTransaction", func(t *testing.T) {
		req := &pb.CreateTransactionRequest{
			Parent:      resourcenames.UserName("user-1"),
			Transaction: nil,
		}
		_, err := handler.CreateTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for nil transaction, got nil")
		}
	})
}

func TestGetTransactionHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	// Create a test transaction
	testTxn := &domain.Transaction{
		ID:         "txn-123",
		UserID:     "user-1",
		Type:       "BUY",
		ExecutedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mockUC.transactions["txn-123"] = testTxn

	t.Run("Success", func(t *testing.T) {
		req := &pb.GetTransactionRequest{
			Name: resourcenames.TransactionName("user-1", "txn-123"),
		}
		resp, err := handler.GetTransaction(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.TransactionId != "txn-123" {
			t.Errorf("expected txn-123, got %s", resp.TransactionId)
		}
		if resp.UserId != "user-1" {
			t.Errorf("expected user-1, got %s", resp.UserId)
		}
	})

	t.Run("InvalidResourceName", func(t *testing.T) {
		req := &pb.GetTransactionRequest{
			Name: "invalid-name",
		}
		_, err := handler.GetTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid resource name, got nil")
		}
	})

	t.Run("WrongUser", func(t *testing.T) {
		req := &pb.GetTransactionRequest{
			Name: resourcenames.TransactionName("user-2", "txn-123"),
		}
		_, err := handler.GetTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for wrong user, got nil")
		}
	})
}

func TestListTransactionsHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	// Add test transactions
	mockUC.transactions["txn-1"] = &domain.Transaction{
		ID:         "txn-1",
		UserID:     "user-1",
		Type:       "BUY",
		ExecutedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		req := &pb.ListTransactionsRequest{
			Parent: resourcenames.UserName("user-1"),
			Filter: &pb.TransactionFilter{
				Symbol: "AAPL",
			},
		}
		resp, err := handler.ListTransactions(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(resp.Transactions) != 1 {
			t.Errorf("expected 1 transaction, got %d", len(resp.Transactions))
		}
	})

	t.Run("InvalidParent", func(t *testing.T) {
		req := &pb.ListTransactionsRequest{
			Parent: "invalid-parent",
		}
		_, err := handler.ListTransactions(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid parent, got nil")
		}
	})
}

func TestUpdateTransactionHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	// Create a test transaction
	testTxn := &domain.Transaction{
		ID:         "txn-123",
		UserID:     "user-1",
		Type:       "BUY",
		ExecutedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mockUC.transactions["txn-123"] = testTxn

	t.Run("Success_WithFieldMask", func(t *testing.T) {
		req := &pb.UpdateTransactionRequest{
			Transaction: &pb.Transaction{
				Name:  resourcenames.TransactionName("user-1", "txn-123"),
				Type:  "SELL",
				Notes: "Updated notes",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"type", "notes"},
			},
		}
		resp, err := handler.UpdateTransaction(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Type != "SELL" {
			t.Errorf("expected SELL, got %s", resp.Type)
		}
		if resp.Notes != "Updated notes" {
			t.Errorf("expected 'Updated notes', got %s", resp.Notes)
		}
	})

	t.Run("InvalidResourceName", func(t *testing.T) {
		req := &pb.UpdateTransactionRequest{
			Transaction: &pb.Transaction{
				Name: "invalid-name",
			},
		}
		_, err := handler.UpdateTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid resource name, got nil")
		}
	})
}

func TestDeleteTransactionHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	// Create a test transaction
	testTxn := &domain.Transaction{
		ID:         "txn-123",
		UserID:     "user-1",
		Type:       "BUY",
		ExecutedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	mockUC.transactions["txn-123"] = testTxn

	t.Run("Success", func(t *testing.T) {
		req := &pb.DeleteTransactionRequest{
			Name: resourcenames.TransactionName("user-1", "txn-123"),
		}
		_, err := handler.DeleteTransaction(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Verify transaction was deleted
		if _, exists := mockUC.transactions["txn-123"]; exists {
			t.Error("expected transaction to be deleted")
		}
	})

	t.Run("InvalidResourceName", func(t *testing.T) {
		req := &pb.DeleteTransactionRequest{
			Name: "invalid-name",
		}
		_, err := handler.DeleteTransaction(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid resource name, got nil")
		}
	})
}

func TestGetOldestTransactionForUserHandler(t *testing.T) {
	mockUC := &MockTransactionUsecase{
		transactions: make(map[string]*domain.Transaction),
	}
	handler := NewTransactionHandler(mockUC)

	t.Run("Success", func(t *testing.T) {
		req := &pb.GetOldestTransactionForUserRequest{
			Parent: resourcenames.UserName("user-1"),
		}
		resp, err := handler.GetOldestTransactionForUser(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.TransactionId != "oldest-txn" {
			t.Errorf("expected oldest-txn, got %s", resp.TransactionId)
		}
	})

	t.Run("InvalidParent", func(t *testing.T) {
		req := &pb.GetOldestTransactionForUserRequest{
			Parent: "invalid-parent",
		}
		_, err := handler.GetOldestTransactionForUser(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid parent, got nil")
		}
	})
}
