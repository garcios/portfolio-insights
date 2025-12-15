package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock HoldingRepository
type mockHoldingRepository struct {
	holdings map[string]*domain.Holding
	err      error
}

func newMockHoldingRepository() *mockHoldingRepository {
	return &mockHoldingRepository{
		holdings: make(map[string]*domain.Holding),
	}
}

func (m *mockHoldingRepository) Upsert(holding *domain.Holding) error {
	if m.err != nil {
		return m.err
	}
	key := holding.UserID + ":" + holding.Symbol
	m.holdings[key] = holding
	return nil
}

func (m *mockHoldingRepository) GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := userID + ":" + symbol
	holding, exists := m.holdings[key]
	if !exists {
		return nil, errors.New("holding not found")
	}
	return holding, nil
}

func (m *mockHoldingRepository) ListByUser(userID string) ([]*domain.Holding, error) {
	if m.err != nil {
		return nil, m.err
	}
	var holdings []*domain.Holding
	for _, holding := range m.holdings {
		if holding.UserID == userID {
			holdings = append(holdings, holding)
		}
	}
	return holdings, nil
}

func (m *mockHoldingRepository) Count() (int64, error) {
	return int64(len(m.holdings)), nil
}

// Mock PortfolioHistoryRepository
type mockPortfolioHistoryRepository struct {
	snapshots map[string]*domain.PortfolioSnapshot
	err       error
}

func newMockPortfolioHistoryRepository() *mockPortfolioHistoryRepository {
	return &mockPortfolioHistoryRepository{
		snapshots: make(map[string]*domain.PortfolioSnapshot),
	}
}

func (m *mockPortfolioHistoryRepository) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	if m.err != nil {
		return m.err
	}
	key := fmt.Sprintf("%s:%s", snapshot.UserID, snapshot.Timestamp.Format("2006-01-02"))
	m.snapshots[key] = snapshot
	return nil
}

func (m *mockPortfolioHistoryRepository) GetHistoryByPeriod(ctx context.Context, userID, period string) ([]*domain.PortfolioSnapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	// For testing, just return all snapshots for user
	var snapshots []*domain.PortfolioSnapshot
	for _, s := range m.snapshots {
		if s.UserID == userID {
			snapshots = append(snapshots, s)
		}
	}
	return snapshots, nil
}

func (m *mockPortfolioHistoryRepository) GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	var snapshots []*domain.PortfolioSnapshot
	for _, s := range m.snapshots {
		if s.UserID == userID {
			if (s.Timestamp.Equal(from) || s.Timestamp.After(from)) && (s.Timestamp.Equal(to) || s.Timestamp.Before(to)) {
				snapshots = append(snapshots, s)
			}
		}
	}
	return snapshots, nil
}

func (m *mockPortfolioHistoryRepository) GetAllUserIDs(ctx context.Context) ([]string, error) {
	// Not implemented for this mock as it's not used in existing tests yet
	return []string{}, nil
}

func (m *mockPortfolioHistoryRepository) SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	key := fmt.Sprintf("%s:%s", userID, date.Format("2006-01-02"))
	_, exists := m.snapshots[key]
	return exists, nil
}

// Mock MarketDataGateway
type mockMarketDataGateway struct {
	prices             map[string]float64
	timestamps         map[string]time.Time
	err                error
	getPriceOnDateFunc func(symbol string, date time.Time) (float64, error)
}

func newMockMarketDataGateway() *mockMarketDataGateway {
	return &mockMarketDataGateway{
		prices:     make(map[string]float64),
		timestamps: make(map[string]time.Time),
	}
}

func (m *mockMarketDataGateway) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	price, exists := m.prices[symbol]
	if !exists {
		return 0, errors.New("price not found")
	}
	return price, nil
}

func (m *mockMarketDataGateway) GetCurrentPrices(ctx context.Context, symbols []string) (map[string]PriceData, error) {
	if m.err != nil {
		return nil, m.err
	}
	prices := make(map[string]PriceData)
	for _, symbol := range symbols {
		if price, exists := m.prices[symbol]; exists {
			ts := time.Now()
			if t, ok := m.timestamps[symbol]; ok {
				ts = t
			}
			prices[symbol] = PriceData{
				Price:     price,
				Timestamp: ts,
			}
		}
	}
	return prices, nil
}

func (m *mockMarketDataGateway) GetCurrencyRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	// For testing, return 1.0 (no conversion)
	return 1.0, nil
}

func (m *mockMarketDataGateway) GetAssetName(ctx context.Context, symbol string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return symbol + " Name", nil
}

func (m *mockMarketDataGateway) GetPriceOnDate(ctx context.Context, symbol string, date time.Time) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.getPriceOnDateFunc != nil {
		return m.getPriceOnDateFunc(symbol, date)
	}
	// For testing, return same price as current price
	price, exists := m.prices[symbol]
	if !exists {
		return 0, errors.New("price not found")
	}
	return price, nil
}

func (m *mockMarketDataGateway) GetCurrencyRateOnDate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	// For testing, return 1.0 (no conversion)
	return 1.0, nil
}

func (m *mockMarketDataGateway) Close() error {
	return nil
}

// Mock CashBalanceRepository
type mockCashBalanceRepository struct {
	balances map[string]*domain.CashBalance
	err      error
}

func newMockCashBalanceRepository() *mockCashBalanceRepository {
	return &mockCashBalanceRepository{
		balances: make(map[string]*domain.CashBalance),
	}
}

func (m *mockCashBalanceRepository) Upsert(balance *domain.CashBalance) error {
	if m.err != nil {
		return m.err
	}
	key := balance.UserID + ":" + balance.Currency
	m.balances[key] = balance
	return nil
}

func (m *mockCashBalanceRepository) GetByUserAndCurrency(userID, currency string) (*domain.CashBalance, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := userID + ":" + currency
	balance, exists := m.balances[key]
	if !exists {
		return nil, errors.New("balance not found")
	}
	return balance, nil
}

func (m *mockCashBalanceRepository) ListByUser(userID string) ([]*domain.CashBalance, error) {
	if m.err != nil {
		return nil, m.err
	}
	var balances []*domain.CashBalance
	for _, b := range m.balances {
		if b.UserID == userID {
			balances = append(balances, b)
		}
	}
	return balances, nil
}

func (m *mockCashBalanceRepository) AddAmount(userID, currency string, amount float64, notes string) error {
	// Not implemented for fetching logic tests
	return nil
}

// Mock TransactionServiceClient
type mockTransactionServiceClient struct {
	transactions []*transactionpb.Transaction
	err          error
}

func newMockTransactionServiceClient() *mockTransactionServiceClient {
	return &mockTransactionServiceClient{
		transactions: []*transactionpb.Transaction{},
	}
}

func (m *mockTransactionServiceClient) CreateTransaction(ctx context.Context, in *transactionpb.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionServiceClient) GetTransaction(ctx context.Context, in *transactionpb.GetTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionServiceClient) UpdateTransaction(ctx context.Context, in *transactionpb.UpdateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionServiceClient) DeleteTransaction(ctx context.Context, in *transactionpb.DeleteTransactionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockTransactionServiceClient) ListTransactions(ctx context.Context, in *transactionpb.ListTransactionsRequest, opts ...grpc.CallOption) (*transactionpb.ListTransactionsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Simple return all transactions
	return &transactionpb.ListTransactionsResponse{
		Transactions: m.transactions,
	}, nil
}

func (m *mockTransactionServiceClient) GetOldestTransactionForUser(ctx context.Context, in *transactionpb.GetOldestTransactionForUserRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func TestGetHoldings_Success(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holdings
	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 150.00,
		LastUpdated: time.Now(),
	}
	repo.holdings["user-123:GOOGL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "GOOGL",
		Quantity:    5,
		AverageCost: 2800.00,
		LastUpdated: time.Now(),
	}

	// Add market prices
	marketData.prices["AAPL"] = 175.50
	marketData.prices["GOOGL"] = 2950.00

	// Execute
	ctx := context.Background()
	holdings, err := uc.GetHoldings(ctx, "user-123")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(holdings) != 2 {
		t.Fatalf("Expected 2 holdings, got %d", len(holdings))
	}

	// Verify AAPL holding
	var aaplHolding *domain.Holding
	for _, h := range holdings {
		if h.Symbol == "AAPL" {
			aaplHolding = h
			break
		}
	}

	if aaplHolding == nil {
		t.Fatal("AAPL holding not found")
	}

	if aaplHolding.CurrentPrice != 175.50 {
		t.Errorf("Expected current price 175.50, got %f", aaplHolding.CurrentPrice)
	}

	if aaplHolding.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %f", aaplHolding.Quantity)
	}
}

func TestGetHoldings_EmptyHoldings(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Execute
	ctx := context.Background()
	holdings, err := uc.GetHoldings(ctx, "user-456")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(holdings) != 0 {
		t.Fatalf("Expected 0 holdings, got %d", len(holdings))
	}
}

func TestGetHoldings_RepositoryError(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	repo.err = errors.New("database error")
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Execute
	ctx := context.Background()
	_, err := uc.GetHoldings(ctx, "user-123")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, repo.err) && err.Error() != "failed to get holdings: database error" {
		t.Errorf("Expected wrapped database error, got %v", err)
	}
}

func TestGetHoldings_MarketDataError(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	marketData.err = errors.New("market data service unavailable")
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holding
	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 150.00,
		LastUpdated: time.Now(),
	}

	// Execute
	ctx := context.Background()
	holdings, err := uc.GetHoldings(ctx, "user-123")

	// Assert - should still return holdings even if prices fail
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(holdings) != 1 {
		t.Fatalf("Expected 1 holding, got %d", len(holdings))
	}

	// Current price should be 0 when market data fails
	if holdings[0].CurrentPrice != 0 {
		t.Errorf("Expected current price 0, got %f", holdings[0].CurrentPrice)
	}
}

func TestGetPortfolioSummary_Success(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()

	// Add Deposit transaction to equal expected Total Cost (15500.0)
	amount := 15500.0
	transactionClient.transactions = append(transactionClient.transactions, &transactionpb.Transaction{
		Type:       "DEP",
		Amount:     &amount,
		ExecutedAt: timestamppb.Now(),
	})

	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holdings
	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 150.00,
		LastUpdated: time.Now(),
	}
	repo.holdings["user-123:GOOGL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "GOOGL",
		Quantity:    5,
		AverageCost: 2800.00,
		LastUpdated: time.Now(),
	}

	// Add market prices
	marketData.prices["AAPL"] = 175.50   // Gain: (175.50 - 150) * 10 = 255
	marketData.prices["GOOGL"] = 2950.00 // Gain: (2950 - 2800) * 5 = 750

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-123", nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if summary.UserID != "user-123" {
		t.Errorf("Expected user_id 'user-123', got '%s'", summary.UserID)
	}

	// Total cost: (150 * 10) + (2800 * 5) = 1500 + 14000 = 15500
	expectedTotalCost := 15500.0
	if summary.TotalCost != expectedTotalCost {
		t.Errorf("Expected total cost %f, got %f", expectedTotalCost, summary.TotalCost)
	}

	// Total value: (175.50 * 10) + (2950 * 5) = 1755 + 14750 = 16505
	expectedTotalValue := 16505.0
	if summary.TotalValue != expectedTotalValue {
		t.Errorf("Expected total value %f, got %f", expectedTotalValue, summary.TotalValue)
	}

	// Gain/Loss: 16505 - 15500 = 1005
	expectedGainLoss := 1005.0
	if summary.GainLoss != expectedGainLoss {
		t.Errorf("Expected gain/loss %f, got %f", expectedGainLoss, summary.GainLoss)
	}

	// Gain/Loss %: (1005 / 15500) * 100 = 6.48387...
	expectedGainLossPct := 6.483870967741935
	if summary.GainLossPct != expectedGainLossPct {
		t.Errorf("Expected gain/loss pct %f, got %f", expectedGainLossPct, summary.GainLossPct)
	}
}

func TestGetPortfolioSummary_EmptyPortfolio(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-456", nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if summary.TotalValue != 0 {
		t.Errorf("Expected total value 0, got %f", summary.TotalValue)
	}

	if summary.TotalCost != 0 {
		t.Errorf("Expected total cost 0, got %f", summary.TotalCost)
	}

	if summary.GainLoss != 0 {
		t.Errorf("Expected gain/loss 0, got %f", summary.GainLoss)
	}

	if summary.GainLossPct != 0 {
		t.Errorf("Expected gain/loss pct 0, got %f", summary.GainLossPct)
	}
}

func TestGetPortfolioSummary_ZeroCostBasis(t *testing.T) {
	// Setup - edge case where cost basis is 0 (shouldn't happen in practice)
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holding with zero cost
	repo.holdings["user-123:FREE"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "FREE",
		Quantity:    10,
		AverageCost: 0,
		LastUpdated: time.Now(),
	}

	marketData.prices["FREE"] = 100.00

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-123", nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should not divide by zero
	if summary.GainLossPct != 0 {
		t.Errorf("Expected gain/loss pct 0 (avoid division by zero), got %f", summary.GainLossPct)
	}
}

func TestGetPortfolioSummary_RepositoryError(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	repo.err = errors.New("database error")
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Execute
	ctx := context.Background()
	_, err := uc.GetPortfolioSummary(ctx, "user-123", nil, nil)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetPortfolioSummary_DayChange(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository() // Added
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient() // Added

	// Define historical prices
	// Current: AAPL 150, GOOGL 2800
	// Yesterday: AAPL 140, GOOGL 2700
	marketData.getPriceOnDateFunc = func(symbol string, date time.Time) (float64, error) {
		// Check if date is roughly yesterday (within a minute tolerance or just check if it's in the past)
		// Since the usecase subtracts exactly 24 hours, we can check if it's before now.
		// For this test, we assume any call to GetPriceOnDate is for the historical calculation.
		if symbol == "AAPL" {
			return 140.0, nil
		}
		if symbol == "GOOGL" {
			return 2700.0, nil
		}
		return 0, errors.New("price not found")
	}

	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holdings
	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00, // Cost basis doesn't affect day change, but needed for total cost
		LastUpdated: time.Now(),
	}
	repo.holdings["user-123:GOOGL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "GOOGL",
		Quantity:    5,
		AverageCost: 2000.00,
		LastUpdated: time.Now(),
	}

	// Add current market prices
	marketData.prices["AAPL"] = 150.00
	marketData.prices["GOOGL"] = 2800.00

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-123", nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Calculate expected values
	// Current Value: (10 * 150) + (5 * 2800) = 1500 + 14000 = 15500
	// Yesterday Value: (10 * 140) + (5 * 2700) = 1400 + 13500 = 14900
	// Day Change: 15500 - 14900 = 600
	// Day Change Pct: (600 / 14900) * 100 = 4.0268...

	expectedTotalValue := 15500.0
	if summary.TotalValue != expectedTotalValue {
		t.Errorf("Expected total value %f, got %f", expectedTotalValue, summary.TotalValue)
	}

	expectedDayChange := 600.0
	if summary.DayChange != expectedDayChange {
		t.Errorf("Expected day change %f, got %f", expectedDayChange, summary.DayChange)
	}

	expectedDayChangePct := (600.0 / 14900.0) * 100
	if math.Abs(summary.DayChangePct-expectedDayChangePct) > 0.000001 {
		t.Errorf("Expected day change pct %f, got %f", expectedDayChangePct, summary.DayChangePct)
	}
}

func TestGetPortfolioSummary_DayChange_Timezone(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()

	// Scenario:
	// Current Time: Dec 2 12:00 UTC
	// Latest Price Update: Dec 1 23:00 UTC (e.g. ingested late previous day)
	// Without fix: "Yesterday" from Now() is Dec 1 12:00. GetPriceOnDate(Dec 1) finds the Dec 1 23:00 price. Result: 0 change.
	// With fix: "Yesterday" from LatestUpdate is Nov 30 23:00. GetPriceOnDate(Nov 30) finds Nov 30 price. Result: Correct change.

	now := time.Date(2025, 12, 2, 12, 0, 0, 0, time.UTC)
	latestPriceTime := time.Date(2025, 12, 1, 23, 0, 0, 0, time.UTC)

	marketData.getPriceOnDateFunc = func(symbol string, date time.Time) (float64, error) {
		// Check if date matches Nov 30
		if date.Year() == 2025 && date.Month() == 11 && date.Day() == 30 {
			if symbol == "AAPL" {
				return 140.0, nil
			}
		}
		// If it asks for Dec 1, it might find the current price (simulating DB behavior)
		if date.Year() == 2025 && date.Month() == 12 && date.Day() == 1 {
			if symbol == "AAPL" {
				return 150.0, nil
			}
		}
		return 0, errors.New("price not found")
	}

	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holdings
	repo.holdings["user-tz:AAPL"] = &domain.Holding{
		UserID:      "user-tz",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00,
		LastUpdated: now,
	}

	// Add current market prices with timestamp
	marketData.prices["AAPL"] = 150.00
	marketData.timestamps["AAPL"] = latestPriceTime

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-tz", nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Expected:
	// Current Value: 10 * 150 = 1500
	// Historical Value (Nov 30): 10 * 140 = 1400
	// Day Change: 100

	expectedDayChange := 100.0
	if summary.DayChange != expectedDayChange {
		t.Errorf("Expected day change %f, got %f. This implies it compared with Dec 1 price instead of Nov 30.", expectedDayChange, summary.DayChange)
	}
}

func TestBackfillPortfolioHistory_Success(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// Add test holdings
	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00,
		LastUpdated: time.Now(),
	}

	// Mock historical prices for Backfill
	marketData.prices["AAPL"] = 150.00

	// Execute
	ctx := context.Background()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	userIDs := []string{"user-123"}
	result := uc.BackfillPortfolioHistory(ctx, userIDs, startDate, endDate, false)

	// Assert
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
	if result.Created != 2 {
		t.Errorf("Expected 2 snapshots created, got %d", result.Created)
	}
	if result.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", result.Errors)
	}
}

func TestBackfillPortfolioHistory_DryRun(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00,
		LastUpdated: time.Now(),
	}
	marketData.prices["AAPL"] = 150.00

	// Execute
	ctx := context.Background()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	userIDs := []string{"user-123"}
	result := uc.BackfillPortfolioHistory(ctx, userIDs, startDate, endDate, true)

	// Assert
	if result.Created != 1 {
		t.Errorf("Expected 1 snapshot created (in dry run count), got %d", result.Created)
	}
	// Verify no snapshot was actually stored
	key := fmt.Sprintf("user-123:%s", startDate.Format("2006-01-02"))
	if _, exists := historyRepo.snapshots[key]; exists {
		t.Error("Expected no snapshot to be stored in dry run")
	}
}

func TestBackfillPortfolioHistory_AlreadyExists(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	repo.holdings["user-123:AAPL"] = &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00,
		LastUpdated: time.Now(),
	}
	marketData.prices["AAPL"] = 150.00

	// Pre-create a snapshot
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = historyRepo.CreateSnapshot(context.Background(), &domain.PortfolioSnapshot{
		UserID:    "user-123",
		Timestamp: date,
	})

	// Execute
	ctx := context.Background()
	userIDs := []string{"user-123"}
	result := uc.BackfillPortfolioHistory(ctx, userIDs, date, date, false)

	// Assert
	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", result.Skipped)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created, got %d", result.Created)
	}
}

func TestGetPortfolioSummary_PeriodDelta(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	cashBalanceRepo := newMockCashBalanceRepository()
	marketData := newMockMarketDataGateway()
	transactionClient := newMockTransactionServiceClient()
	uc := NewPortfolioUsecase(repo, historyRepo, cashBalanceRepo, marketData, transactionClient)

	// User ID
	userID := "user-delta"

	// 1. Setup Current State (End Date)
	// Holding: AAPL: 10 units. Current Price: 150. Current Value: 1500.
	// Cash: 0 (simplified)
	repo.holdings[userID+":AAPL"] = &domain.Holding{
		UserID:      userID,
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 100.00,
		LastUpdated: time.Now(),
	}
	marketData.prices["AAPL"] = 150.00

	// 2. Setup Historical State (Start Date) - via History Repo Mock
	// Snapshot at Start Date:
	// Total Value: 1200. (Maybe price was 120).
	// Total Cost: 1000.
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	historyRepo.snapshots[userID+":2025-01-01"] = &domain.PortfolioSnapshot{
		UserID:         userID,
		Timestamp:      startDate,
		TotalValue:     1200.00,
		TotalCostBasis: 1000.00,
	}

	// 3. Setup Transactions (Net Invested & Dividends)
	// Let's say we bought 0 units in period. Net Invested = 0.
	// But we received Dividends.
	// Transaction Client Mock needs to return transactions.
	// We'll trust the mock default returns empty list for now, so NetInvested=0, Dividends=0.

	// Execute
	ctx := context.Background()
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	summary, err := uc.GetPortfolioSummary(ctx, userID, &startDate, &endDate)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Current Value should be 1500
	if summary.TotalValue != 1500.00 {
		t.Errorf("Expected TotalValue 1500.00, got %f", summary.TotalValue)
	}

	// Capital Gain Delta:
	// Current Cap Gain = 1500 - 1000 = 500.
	// Start Cap Gain = 1200 - 1000 = 200.
	// Delta = 500 - 200 = 300.
	if summary.CapitalGain != 300.00 {
		t.Errorf("Expected CapitalGain 300.00 (500-200), got %f", summary.CapitalGain)
	}

	// Currency Gain:
	// Formula: (EndValue - StartValue) - NetInvested - CapGain - Dividends
	// (1500 - 1200) - 0 - 300 - 0 = 300 - 300 = 0.
	if math.Abs(summary.CurrencyGain) > 0.01 {
		t.Errorf("Expected CurrencyGain 0.00, got %f", summary.CurrencyGain)
	}
}
