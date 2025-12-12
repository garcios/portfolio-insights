package infrastructure

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"log/slog"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/nats-io/nats.go"
)

// MockHoldingRepository is a mock implementation of domain.HoldingRepository for testing
type MockHoldingRepository struct {
	holdings map[string]*domain.Holding // key: userID-symbol
	getError error
}

func NewMockHoldingRepository() *MockHoldingRepository {
	return &MockHoldingRepository{
		holdings: make(map[string]*domain.Holding),
	}
}

func (m *MockHoldingRepository) Upsert(holding *domain.Holding) error {
	key := holding.UserID + "-" + holding.Symbol
	m.holdings[key] = holding
	return nil
}

func (m *MockHoldingRepository) GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	key := userID + "-" + symbol
	holding, exists := m.holdings[key]
	if !exists {
		return nil, fmt.Errorf("holding not found")
	}
	return holding, nil
}

func (m *MockHoldingRepository) ListByUser(userID string) ([]*domain.Holding, error) {
	var result []*domain.Holding
	for _, h := range m.holdings {
		if h.UserID == userID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (m *MockHoldingRepository) Count() (int, error) {
	return len(m.holdings), nil
}

func (m *MockHoldingRepository) DeleteZeroQuantityHoldings() error {
	for key, h := range m.holdings {
		if h.Quantity == 0 {
			delete(m.holdings, key)
		}
	}
	return nil
}

// MockCashBalanceRepository is a mock implementation of domain.CashBalanceRepository for testing
type MockCashBalanceRepository struct {
	balances map[string]*domain.CashBalance // key: userID-currency
}

func NewMockCashBalanceRepository() *MockCashBalanceRepository {
	return &MockCashBalanceRepository{
		balances: make(map[string]*domain.CashBalance),
	}
}

func (m *MockCashBalanceRepository) Upsert(balance *domain.CashBalance) error {
	key := balance.UserID + "-" + balance.Currency
	m.balances[key] = balance
	return nil
}

func (m *MockCashBalanceRepository) GetByUserAndCurrency(userID, currency string) (*domain.CashBalance, error) {
	key := userID + "-" + currency
	balance, exists := m.balances[key]
	if !exists {
		return nil, fmt.Errorf("cash balance not found")
	}
	return balance, nil
}

func (m *MockCashBalanceRepository) ListByUser(userID string) ([]*domain.CashBalance, error) {
	var result []*domain.CashBalance
	for _, b := range m.balances {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *MockCashBalanceRepository) AddAmount(userID, currency string, amount float64, notes string) error {
	key := userID + "-" + currency
	balance, exists := m.balances[key]
	if !exists {
		balance = &domain.CashBalance{
			UserID:   userID,
			Currency: currency,
			Balance:  0,
			Notes:    notes,
		}
		m.balances[key] = balance
	}
	balance.Balance += amount
	// Update notes if provided
	if notes != "" {
		balance.Notes = notes
	}
	return nil
}

// Helper function to create float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// TestHandleTransactionCreated_INT tests Interest income transactions
func TestHandleTransactionCreated_INT(t *testing.T) {
	tests := []struct {
		name           string
		event          TransactionCreatedEvent
		expectedAmount float64
		shouldFail     bool
	}{
		{
			name: "INT_NewCashBalance",
			event: TransactionCreatedEvent{
				TransactionID: "txn-1",
				UserID:        "user-1",
				Type:          "INT",
				Amount:        float64Ptr(25.50),
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 25.50,
			shouldFail:     false,
		},
		{
			name: "INT_MissingAmount",
			event: TransactionCreatedEvent{
				TransactionID: "txn-2",
				UserID:        "user-1",
				Type:          "INT",
				Amount:        nil,
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 0,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashRepo := NewMockCashBalanceRepository()

			subscriber := &NATSSubscriber{
				repo:            NewMockHoldingRepository(),
				cashBalanceRepo: cashRepo,
				logger:          slog.Default(),
				createdTopic:    "test.created",
			}

			data, _ := json.Marshal(tt.event)
			subscriber.handleTransactionCreated(&nats.Msg{Data: data})

			if !tt.shouldFail {
				balance, err := cashRepo.GetByUserAndCurrency(tt.event.UserID, "USD")
				if err != nil {
					t.Fatalf("expected cash balance to exist, got error: %v", err)
				}
				if balance.Balance != tt.expectedAmount {
					t.Errorf("expected cash amount %.2f, got %.2f", tt.expectedAmount, balance.Balance)
				}
			}
		})
	}
}

// TestHandleTransactionCreated_DIV tests Dividend income transactions
func TestHandleTransactionCreated_DIV(t *testing.T) {
	tests := []struct {
		name           string
		event          TransactionCreatedEvent
		expectedAmount float64
		shouldFail     bool
	}{
		{
			name: "DIV_WithSymbol",
			event: TransactionCreatedEvent{
				TransactionID: "txn-1",
				UserID:        "user-1",
				Type:          "DIV",
				AssetSymbol:   stringPtr("AAPL"),
				Amount:        float64Ptr(15.75),
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 15.75,
			shouldFail:     false,
		},
		{
			name: "DIV_MissingAmount",
			event: TransactionCreatedEvent{
				TransactionID: "txn-2",
				UserID:        "user-1",
				Type:          "DIV",
				AssetSymbol:   stringPtr("AAPL"),
				Amount:        nil,
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 0,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashRepo := NewMockCashBalanceRepository()

			subscriber := &NATSSubscriber{
				repo:            NewMockHoldingRepository(),
				cashBalanceRepo: cashRepo,
				logger:          slog.Default(),
				createdTopic:    "test.created",
			}

			data, _ := json.Marshal(tt.event)
			subscriber.handleTransactionCreated(&nats.Msg{Data: data})

			if !tt.shouldFail {
				balance, err := cashRepo.GetByUserAndCurrency(tt.event.UserID, "USD")
				if err != nil {
					t.Fatalf("expected cash balance to exist, got error: %v", err)
				}
				if balance.Balance != tt.expectedAmount {
					t.Errorf("expected cash amount %.2f, got %.2f", tt.expectedAmount, balance.Balance)
				}
			}
		})
	}
}

// TestHandleTransactionCreated_DEP tests Deposit transactions
func TestHandleTransactionCreated_DEP(t *testing.T) {
	tests := []struct {
		name           string
		event          TransactionCreatedEvent
		expectedAmount float64
		shouldFail     bool
	}{
		{
			name: "DEP_InitialDeposit",
			event: TransactionCreatedEvent{
				TransactionID: "txn-1",
				UserID:        "user-1",
				Type:          "DEP",
				Amount:        float64Ptr(1000.00),
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 1000.00,
			shouldFail:     false,
		},
		{
			name: "DEP_MissingAmount",
			event: TransactionCreatedEvent{
				TransactionID: "txn-2",
				UserID:        "user-1",
				Type:          "DEP",
				Amount:        nil,
				ExecutedAt:    time.Now(),
			},
			expectedAmount: 0,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashRepo := NewMockCashBalanceRepository()

			subscriber := &NATSSubscriber{
				repo:            NewMockHoldingRepository(),
				cashBalanceRepo: cashRepo,
				logger:          slog.Default(),
				createdTopic:    "test.created",
			}

			data, _ := json.Marshal(tt.event)
			subscriber.handleTransactionCreated(&nats.Msg{Data: data})

			if !tt.shouldFail {
				balance, err := cashRepo.GetByUserAndCurrency(tt.event.UserID, "USD")
				if err != nil {
					t.Fatalf("expected cash balance to exist, got error: %v", err)
				}
				if balance.Balance != tt.expectedAmount {
					t.Errorf("expected cash amount %.2f, got %.2f", tt.expectedAmount, balance.Balance)
				}
			}
		})
	}
}

// TestHandleTransactionCreated_WIT tests Withdrawal transactions
func TestHandleTransactionCreated_WIT(t *testing.T) {
	tests := []struct {
		name           string
		event          TransactionCreatedEvent
		initialBalance float64
		expectedAmount float64
		shouldFail     bool
	}{
		{
			name: "WIT_FromExistingBalance",
			event: TransactionCreatedEvent{
				TransactionID: "txn-1",
				UserID:        "user-1",
				Type:          "WIT",
				Amount:        float64Ptr(500.00),
				ExecutedAt:    time.Now(),
			},
			initialBalance: 1000.00,
			expectedAmount: 500.00,
			shouldFail:     false,
		},
		{
			name: "WIT_NegativeBalance",
			event: TransactionCreatedEvent{
				TransactionID: "txn-2",
				UserID:        "user-1",
				Type:          "WIT",
				Amount:        float64Ptr(1500.00),
				ExecutedAt:    time.Now(),
			},
			initialBalance: 1000.00,
			expectedAmount: -500.00,
			shouldFail:     false,
		},
		{
			name: "WIT_MissingAmount",
			event: TransactionCreatedEvent{
				TransactionID: "txn-3",
				UserID:        "user-1",
				Type:          "WIT",
				Amount:        nil,
				ExecutedAt:    time.Now(),
			},
			initialBalance: 0,
			expectedAmount: 0,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashRepo := NewMockCashBalanceRepository()

			// Set initial balance if needed
			if tt.initialBalance > 0 {
				if err := cashRepo.AddAmount("user-1", "USD", tt.initialBalance, ""); err != nil {
					t.Fatalf("failed to set initial balance: %v", err)
				}
			}

			subscriber := &NATSSubscriber{
				repo:            NewMockHoldingRepository(),
				cashBalanceRepo: cashRepo,
				logger:          slog.Default(),
				createdTopic:    "test.created",
			}

			data, _ := json.Marshal(tt.event)
			subscriber.handleTransactionCreated(&nats.Msg{Data: data})

			if !tt.shouldFail {
				balance, err := cashRepo.GetByUserAndCurrency(tt.event.UserID, "USD")
				if err != nil {
					t.Fatalf("expected cash balance to exist, got error: %v", err)
				}
				if balance.Balance != tt.expectedAmount {
					t.Errorf("expected cash amount %.2f, got %.2f", tt.expectedAmount, balance.Balance)
				}
			}
		})
	}
}

// TestUpdateCashBalance tests the updateCashBalance helper function
func TestUpdateCashBalance(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		currency       string
		amount         float64
		expectedAmount float64
	}{
		{
			name:           "CreateNewCashBalance",
			userID:         "user-1",
			currency:       "USD",
			amount:         100.00,
			expectedAmount: 100.00,
		},
		{
			name:           "AddToExistingCash",
			userID:         "user-1",
			currency:       "USD",
			amount:         50.00,
			expectedAmount: 50.00,
		},
		{
			name:           "SubtractFromCash",
			userID:         "user-1",
			currency:       "USD",
			amount:         -30.00,
			expectedAmount: -30.00,
		},
		{
			name:           "MultiCurrency_AUD",
			userID:         "user-2",
			currency:       "AUD",
			amount:         200.00,
			expectedAmount: 200.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashRepo := NewMockCashBalanceRepository()

			subscriber := &NATSSubscriber{
				repo:            NewMockHoldingRepository(),
				cashBalanceRepo: cashRepo,
				logger:          slog.Default(),
			}

			err := subscriber.updateCashBalance(tt.userID, tt.currency, tt.amount, "")
			if err != nil {
				t.Fatalf("updateCashBalance failed: %v", err)
			}

			balance, err := cashRepo.GetByUserAndCurrency(tt.userID, tt.currency)
			if err != nil {
				t.Fatalf("expected cash balance to exist, got error: %v", err)
			}

			if balance.Balance != tt.expectedAmount {
				t.Errorf("expected cash amount %.2f, got %.2f", tt.expectedAmount, balance.Balance)
			}

			if balance.Currency != tt.currency {
				t.Errorf("expected currency %s, got %s", tt.currency, balance.Currency)
			}
		})
	}
}

// TestCashTransactions_Sequential tests multiple cash transactions in sequence
func TestCashTransactions_Sequential(t *testing.T) {
	cashRepo := NewMockCashBalanceRepository()
	subscriber := &NATSSubscriber{
		repo:            NewMockHoldingRepository(),
		cashBalanceRepo: cashRepo,
		logger:          slog.Default(),
		createdTopic:    "test.created",
	}

	userID := "user-1"

	// 1. Initial deposit
	depEvent := TransactionCreatedEvent{
		TransactionID: "txn-1",
		UserID:        userID,
		Type:          "DEP",
		Amount:        float64Ptr(1000.00),
		ExecutedAt:    time.Now(),
	}
	data, _ := json.Marshal(depEvent)
	subscriber.handleTransactionCreated(&nats.Msg{Data: data})

	balance, _ := cashRepo.GetByUserAndCurrency(userID, "USD")
	if balance.Balance != 1000.00 {
		t.Errorf("after deposit: expected 1000.00, got %.2f", balance.Balance)
	}

	// 2. Interest income
	intEvent := TransactionCreatedEvent{
		TransactionID: "txn-2",
		UserID:        userID,
		Type:          "INT",
		Amount:        float64Ptr(25.50),
		ExecutedAt:    time.Now(),
	}
	data, _ = json.Marshal(intEvent)
	subscriber.handleTransactionCreated(&nats.Msg{Data: data})

	balance, _ = cashRepo.GetByUserAndCurrency(userID, "USD")
	if balance.Balance != 1025.50 {
		t.Errorf("after interest: expected 1025.50, got %.2f", balance.Balance)
	}

	// 3. Dividend income
	divEvent := TransactionCreatedEvent{
		TransactionID: "txn-3",
		UserID:        userID,
		Type:          "DIV",
		AssetSymbol:   stringPtr("AAPL"),
		Amount:        float64Ptr(15.75),
		ExecutedAt:    time.Now(),
	}
	data, _ = json.Marshal(divEvent)
	subscriber.handleTransactionCreated(&nats.Msg{Data: data})

	balance, _ = cashRepo.GetByUserAndCurrency(userID, "USD")
	if balance.Balance != 1041.25 {
		t.Errorf("after dividend: expected 1041.25, got %.2f", balance.Balance)
	}

	// 4. Withdrawal
	witEvent := TransactionCreatedEvent{
		TransactionID: "txn-4",
		UserID:        userID,
		Type:          "WIT",
		Amount:        float64Ptr(500.00),
		ExecutedAt:    time.Now(),
	}
	data, _ = json.Marshal(witEvent)
	subscriber.handleTransactionCreated(&nats.Msg{Data: data})

	balance, _ = cashRepo.GetByUserAndCurrency(userID, "USD")
	if balance.Balance != 541.25 {
		t.Errorf("after withdrawal: expected 541.25, got %.2f", balance.Balance)
	}
}

// TestCashTransactions_DoNotAffectEquity tests that cash transactions don't affect equity holdings
func TestCashTransactions_DoNotAffectEquity(t *testing.T) {
	repo := NewMockHoldingRepository()
	cashRepo := NewMockCashBalanceRepository()

	// Create an equity holding
	equityHolding := &domain.Holding{
		UserID:      "user-1",
		Symbol:      "AAPL",
		Quantity:    100.0,
		AverageCost: 150.50,
		Currency:    "USD",
	}
	if err := repo.Upsert(equityHolding); err != nil {
		t.Fatalf("failed to create equity holding: %v", err)
	}

	subscriber := &NATSSubscriber{
		repo:            repo,
		cashBalanceRepo: cashRepo,
		logger:          slog.Default(),
		createdTopic:    "test.created",
	}

	// Process a dividend transaction
	divEvent := TransactionCreatedEvent{
		TransactionID: "txn-1",
		UserID:        "user-1",
		Type:          "DIV",
		AssetSymbol:   stringPtr("AAPL"),
		Amount:        float64Ptr(15.75),
		ExecutedAt:    time.Now(),
	}
	data, _ := json.Marshal(divEvent)
	subscriber.handleTransactionCreated(&nats.Msg{Data: data})

	// Verify equity holding unchanged
	holding, err := repo.GetByUserAndSymbol("user-1", "AAPL")
	if err != nil {
		t.Fatalf("equity holding should still exist")
	}
	if holding.Quantity != 100.0 {
		t.Errorf("equity quantity should be unchanged, got %.2f", holding.Quantity)
	}
	if holding.AverageCost != 150.50 {
		t.Errorf("equity average cost should be unchanged, got %.2f", holding.AverageCost)
	}

	// Verify cash balance created
	balance, err := cashRepo.GetByUserAndCurrency("user-1", "USD")
	if err != nil {
		t.Fatalf("cash balance should be created")
	}
	if balance.Balance != 15.75 {
		t.Errorf("expected cash 15.75, got %.2f", balance.Balance)
	}
}
