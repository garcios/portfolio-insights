package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

// TestTransactionValidation_BUY tests validation for BUY transactions
func TestTransactionValidation_BUY(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_BUY", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			Quantity:      domain.Float64Ptr(10),
			PricePerShare: domain.Float64Ptr(150.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid BUY, got %v", err)
		}
	})

	t.Run("BUY_MissingSymbol", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Quantity:      domain.Float64Ptr(10),
			PricePerShare: domain.Float64Ptr(150.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY without symbol")
		}
	})

	t.Run("BUY_MissingQuantity", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			PricePerShare: domain.Float64Ptr(150.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY without quantity")
		}
	})

	t.Run("BUY_MissingPrice", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeBuy,
			Symbol:     domain.StringPtr("AAPL"),
			Quantity:   domain.Float64Ptr(10),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY without price_per_share")
		}
	})

	t.Run("BUY_ZeroQuantity", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			Quantity:      domain.Float64Ptr(0),
			PricePerShare: domain.Float64Ptr(150.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY with zero quantity")
		}
	})

	t.Run("BUY_NegativePrice", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			Quantity:      domain.Float64Ptr(10),
			PricePerShare: domain.Float64Ptr(-150.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY with negative price")
		}
	})

	t.Run("BUY_WithAmount_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			Quantity:      domain.Float64Ptr(10),
			PricePerShare: domain.Float64Ptr(150.0),
			Amount:        domain.Float64Ptr(1500.0), // Should not be set for BUY
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for BUY with amount field set")
		}
	})
}

// TestTransactionValidation_SELL tests validation for SELL transactions
func TestTransactionValidation_SELL(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_SELL", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeSell,
			Symbol:        domain.StringPtr("GOOGL"),
			Quantity:      domain.Float64Ptr(5),
			PricePerShare: domain.Float64Ptr(2500.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid SELL, got %v", err)
		}
	})

	t.Run("SELL_MissingSymbol", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeSell,
			Quantity:      domain.Float64Ptr(5),
			PricePerShare: domain.Float64Ptr(2500.0),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for SELL without symbol")
		}
	})
}

// TestTransactionValidation_INT tests validation for Interest transactions
func TestTransactionValidation_INT(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_INT", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeInterest,
			Amount:     domain.Float64Ptr(25.50),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid INT, got %v", err)
		}
	})

	t.Run("INT_MissingAmount", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeInterest,
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for INT without amount")
		}
	})

	t.Run("INT_ZeroAmount", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeInterest,
			Amount:     domain.Float64Ptr(0),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for INT with zero amount")
		}
	})

	t.Run("INT_WithSymbol_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeInterest,
			Symbol:     domain.StringPtr("AAPL"), // Should not be set for INT
			Amount:     domain.Float64Ptr(25.50),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for INT with symbol field set")
		}
	})

	t.Run("INT_WithQuantity_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeInterest,
			Quantity:   domain.Float64Ptr(10), // Should not be set for INT
			Amount:     domain.Float64Ptr(25.50),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for INT with quantity field set")
		}
	})
}

// TestTransactionValidation_DIV tests validation for Dividend transactions
func TestTransactionValidation_DIV(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_DIV", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDividend,
			Symbol:     domain.StringPtr("AAPL"),
			Amount:     domain.Float64Ptr(15.75),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid DIV, got %v", err)
		}
	})

	t.Run("DIV_MissingSymbol", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDividend,
			Amount:     domain.Float64Ptr(15.75),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DIV without symbol")
		}
	})

	t.Run("DIV_MissingAmount", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDividend,
			Symbol:     domain.StringPtr("AAPL"),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DIV without amount")
		}
	})

	t.Run("DIV_WithQuantity_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDividend,
			Symbol:     domain.StringPtr("AAPL"),
			Quantity:   domain.Float64Ptr(10), // Should not be set for DIV
			Amount:     domain.Float64Ptr(15.75),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DIV with quantity field set")
		}
	})

	t.Run("DIV_AssetNotFound", func(t *testing.T) {
		marketGateway.exists = false
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDividend,
			Symbol:     domain.StringPtr("INVALID"),
			Amount:     domain.Float64Ptr(15.75),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DIV with invalid asset")
		}
		marketGateway.exists = true
	})
}

// TestTransactionValidation_DEP tests validation for Deposit transactions
func TestTransactionValidation_DEP(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_DEP", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDeposit,
			Amount:     domain.Float64Ptr(1000.00),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid DEP, got %v", err)
		}
	})

	t.Run("DEP_MissingAmount", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDeposit,
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DEP without amount")
		}
	})

	t.Run("DEP_WithSymbol_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeDeposit,
			Symbol:     domain.StringPtr("AAPL"), // Should not be set for DEP
			Amount:     domain.Float64Ptr(1000.00),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for DEP with symbol field set")
		}
	})
}

// TestTransactionValidation_WIT tests validation for Withdraw transactions
func TestTransactionValidation_WIT(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("Valid_WIT", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeWithdraw,
			Amount:     domain.Float64Ptr(500.00),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err != nil {
			t.Fatalf("expected no error for valid WIT, got %v", err)
		}
	})

	t.Run("WIT_MissingAmount", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       domain.TransactionTypeWithdraw,
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for WIT without amount")
		}
	})

	t.Run("WIT_WithPricePerShare_ShouldFail", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeWithdraw,
			PricePerShare: domain.Float64Ptr(150.0), // Should not be set for WIT
			Amount:        domain.Float64Ptr(500.00),
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for WIT with price_per_share field set")
		}
	})
}

// TestTransactionValidation_Common tests common validation logic
func TestTransactionValidation_Common(t *testing.T) {
	repo := NewMockRepo()
	userGateway := &MockUserGateway{exists: true}
	marketGateway := &MockMarketDataGateway{exists: true}
	eventPublisher := &MockEventPublisher{}
	uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)

	t.Run("InvalidTransactionType", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:     "user-1",
			Type:       "INVALID",
			Amount:     domain.Float64Ptr(100.00),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for invalid transaction type")
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userGateway.exists = false
		txn := &domain.Transaction{
			UserID:     "invalid-user",
			Type:       domain.TransactionTypeDeposit,
			Amount:     domain.Float64Ptr(100.00),
			ExecutedAt: time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
		userGateway.exists = true
	})

	t.Run("InvalidPriceCurrency", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:        "user-1",
			Type:          domain.TransactionTypeBuy,
			Symbol:        domain.StringPtr("AAPL"),
			Quantity:      domain.Float64Ptr(10),
			PricePerShare: domain.Float64Ptr(150.0),
			PriceCurrency: "US", // Invalid: must be 3 characters
			ExecutedAt:    time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for invalid price currency length")
		}
	})

	t.Run("InvalidBrokerageCurrency", func(t *testing.T) {
		txn := &domain.Transaction{
			UserID:            "user-1",
			Type:              domain.TransactionTypeBuy,
			Symbol:            domain.StringPtr("AAPL"),
			Quantity:          domain.Float64Ptr(10),
			PricePerShare:     domain.Float64Ptr(150.0),
			BrokerageCurrency: "USDO", // Invalid: must be 3 characters
			ExecutedAt:        time.Now(),
		}
		err := uc.CreateTransaction(context.Background(), txn)
		if err == nil {
			t.Error("expected error for invalid brokerage currency length")
		}
	})
}
