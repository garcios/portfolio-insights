package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

// Mock RedisClient
type mockRedisClient struct {
	data map[string]string
	err  error
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
	}
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = value.(string)
	return nil
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("redis: nil")
	}
	return val, nil
}

func (m *mockRedisClient) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockRedisClient) Close() error                   { return nil }
func (m *mockRedisClient) Ping(ctx context.Context) error { return nil }

// Mock Delegate PortfolioUsecase
type mockDelegateUsecase struct {
	summary *domain.PortfolioSummary
	err     error
	called  bool
}

func (m *mockDelegateUsecase) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.summary, nil
}

func (m *mockDelegateUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.summary, nil
}

// Methods not used in caching tests can be minimal implementation
func (m *mockDelegateUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	return nil, nil
}
func (m *mockDelegateUsecase) BackfillPortfolioHistory(ctx context.Context, userIDs []string, startDate, endDate time.Time, dryRun bool) BackfillResult {
	return BackfillResult{}
}

func TestGetPortfolioSummary_CacheHit(t *testing.T) {
	// Setup
	mockRedis := newMockRedisClient()
	mockDelegate := &mockDelegateUsecase{
		summary: &domain.PortfolioSummary{UserID: "user-1", TotalValue: 200},
	}

	// Pre-populate cache
	cachedSummary := domain.PortfolioSummary{UserID: "user-1", TotalValue: 100}
	bytes, _ := json.Marshal(cachedSummary)
	mockRedis.data["portfolio:summary:user-1"] = string(bytes)

	uc := NewCachingPortfolioUsecase(mockDelegate, mockRedis, config.CachingConfig{Enabled: true, SummaryTTLSeconds: 300})

	// Execute
	summary, err := uc.GetPortfolioSummary(context.Background(), "user-1")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if summary.TotalValue != 100 { // Should get cached value (100) not delegate value (200)
		t.Errorf("Expected cached value 100, got %f", summary.TotalValue)
	}
	if mockDelegate.called {
		t.Error("Expected delegate NOT to be called on cache hit")
	}
}

func TestGetPortfolioSummary_CacheMiss(t *testing.T) {
	// Setup
	mockRedis := newMockRedisClient()
	mockDelegate := &mockDelegateUsecase{
		summary: &domain.PortfolioSummary{UserID: "user-1", TotalValue: 200},
	}

	uc := NewCachingPortfolioUsecase(mockDelegate, mockRedis, config.CachingConfig{Enabled: true, SummaryTTLSeconds: 300})

	// Execute
	summary, err := uc.GetPortfolioSummary(context.Background(), "user-1")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if summary.TotalValue != 200 { // Should get delegate value
		t.Errorf("Expected delegate value 200, got %f", summary.TotalValue)
	}
	if !mockDelegate.called {
		t.Error("Expected delegate to be called on cache miss")
	}

	// Verify it wrote to cache
	cached, ok := mockRedis.data["portfolio:summary:user-1"]
	if !ok {
		t.Error("Expected result to be cached")
	}
	var cachedObj domain.PortfolioSummary
	if err := json.Unmarshal([]byte(cached), &cachedObj); err != nil {
		t.Fatalf("Failed to unmarshal cached object: %v", err)
	}
	if cachedObj.TotalValue != 200 {
		t.Errorf("Expected cached value 200, got %f", cachedObj.TotalValue)
	}
}

func TestGetHistoricalPortfolioSummary_CacheHit(t *testing.T) {
	// Setup
	mockRedis := newMockRedisClient()
	mockDelegate := &mockDelegateUsecase{
		summary: &domain.PortfolioSummary{UserID: "user-1", TotalValue: 500},
	}

	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dateStr := "2025-01-01"

	// Pre-populate cache
	cachedSummary := domain.PortfolioSummary{UserID: "user-1", TotalValue: 300}
	bytes, _ := json.Marshal(cachedSummary)
	mockRedis.data["portfolio:history:user-1:"+dateStr] = string(bytes) // Note key format implementation

	uc := NewCachingPortfolioUsecase(mockDelegate, mockRedis, config.CachingConfig{Enabled: true, HistoryTTLSeconds: 86400})

	// Execute
	summary, err := uc.GetHistoricalPortfolioSummary(context.Background(), "user-1", date)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if summary.TotalValue != 300 {
		t.Errorf("Expected cached value 300, got %f", summary.TotalValue)
	}
}
