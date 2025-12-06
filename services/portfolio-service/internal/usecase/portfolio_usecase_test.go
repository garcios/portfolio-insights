package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
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

func TestGetHoldings_Success(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	marketData.err = errors.New("market data service unavailable")
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	summary, err := uc.GetPortfolioSummary(ctx, "user-123")

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

	// Execute
	ctx := context.Background()
	summary, err := uc.GetPortfolioSummary(ctx, "user-456")

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	summary, err := uc.GetPortfolioSummary(ctx, "user-123")

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

	// Execute
	ctx := context.Background()
	_, err := uc.GetPortfolioSummary(ctx, "user-123")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetPortfolioSummary_DayChange(t *testing.T) {
	// Setup
	repo := newMockHoldingRepository()
	historyRepo := newMockPortfolioHistoryRepository()
	marketData := newMockMarketDataGateway()

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

	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	summary, err := uc.GetPortfolioSummary(ctx, "user-123")

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
	marketData := newMockMarketDataGateway()

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

	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	summary, err := uc.GetPortfolioSummary(ctx, "user-tz")

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
	marketData := newMockMarketDataGateway()
	uc := NewPortfolioUsecase(repo, historyRepo, marketData)

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
