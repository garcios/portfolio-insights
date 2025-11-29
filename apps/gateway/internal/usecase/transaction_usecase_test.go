package usecase

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
)

// MockTransactionGateway is a manual mock for TransactionGateway
type MockTransactionGateway struct {
	CreateTransactionFunc func(ctx context.Context, input gateway.CreateTransactionInput) (*entity.Transaction, error)
}

func (m *MockTransactionGateway) CreateTransaction(ctx context.Context, input gateway.CreateTransactionInput) (*entity.Transaction, error) {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, input)
	}
	return nil, nil
}

// MockTransactionFileGateway is a manual mock for TransactionFileGateway
type MockTransactionFileGateway struct {
	UploadCSVFunc func(ctx context.Context, userID string, file io.Reader, filename string) error
}

func (m *MockTransactionFileGateway) UploadCSV(ctx context.Context, userID string, file io.Reader, filename string) error {
	if m.UploadCSVFunc != nil {
		return m.UploadCSVFunc(ctx, userID, file, filename)
	}
	return nil
}

func TestTransactionUseCase_CreateTransaction(t *testing.T) {
	mockGateway := &MockTransactionGateway{
		CreateTransactionFunc: func(ctx context.Context, input gateway.CreateTransactionInput) (*entity.Transaction, error) {
			return entity.NewTransaction(
				"tx-1",
				input.UserID,
				input.Symbol,
				input.Type,
				input.Quantity,
				input.PricePerShare,
				input.PriceCurrency,
				input.ExecutedAt,
			), nil
		},
	}

	mockFileGateway := &MockTransactionFileGateway{}

	uc := NewTransactionUseCase(mockGateway, mockFileGateway)

	validInput := CreateTransactionInput{
		Symbol:        "AAPL",
		Type:          entity.TransactionTypeBuy,
		Quantity:      10,
		PricePerShare: 150.0,
		PriceCurrency: "USD",
		ExecutedAt:    time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	t.Run("success", func(t *testing.T) {
		tx, err := uc.CreateTransaction(context.Background(), "user-1", validInput)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tx.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", tx.Symbol)
		}
	})

	t.Run("validation error - empty symbol", func(t *testing.T) {
		input := validInput
		input.Symbol = ""
		_, err := uc.CreateTransaction(context.Background(), "user-1", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "symbol is required") {
			t.Errorf("expected 'symbol is required' error, got %v", err)
		}
	})

	t.Run("validation error - negative quantity", func(t *testing.T) {
		input := validInput
		input.Quantity = -10
		_, err := uc.CreateTransaction(context.Background(), "user-1", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("validation error - future date", func(t *testing.T) {
		input := validInput
		input.ExecutedAt = time.Now().Add(24 * time.Hour)
		_, err := uc.CreateTransaction(context.Background(), "user-1", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("validation error - invalid type", func(t *testing.T) {
		input := validInput
		input.Type = "INVALID"
		_, err := uc.CreateTransaction(context.Background(), "user-1", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
