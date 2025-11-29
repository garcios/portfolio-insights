package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

// MockPortfolioGateway is a manual mock for PortfolioGateway
type MockPortfolioGateway struct {
	GetPortfolioFunc            func(ctx context.Context, userID string) (*entity.Portfolio, error)
	GetPortfolioSummaryFunc     func(ctx context.Context, userID string) (*entity.PortfolioSummary, error)
	GetHoldingsFunc             func(ctx context.Context, userID string) ([]*entity.Holding, error)
	GetPortfolioPerformanceFunc func(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error)
}

func (m *MockPortfolioGateway) GetPortfolio(ctx context.Context, userID string) (*entity.Portfolio, error) {
	if m.GetPortfolioFunc != nil {
		return m.GetPortfolioFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPortfolioGateway) GetPortfolioSummary(ctx context.Context, userID string) (*entity.PortfolioSummary, error) {
	if m.GetPortfolioSummaryFunc != nil {
		return m.GetPortfolioSummaryFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPortfolioGateway) GetHoldings(ctx context.Context, userID string) ([]*entity.Holding, error) {
	if m.GetHoldingsFunc != nil {
		return m.GetHoldingsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPortfolioGateway) GetPortfolioPerformance(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error) {
	if m.GetPortfolioPerformanceFunc != nil {
		return m.GetPortfolioPerformanceFunc(ctx, userID, period)
	}
	return nil, nil
}

func TestPortfolioUseCase_GetPortfolio(t *testing.T) {
	mockGateway := &MockPortfolioGateway{}
	uc := NewPortfolioUseCase(mockGateway)

	portfolio, err := uc.GetPortfolio(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if portfolio.UserID != "user-1" {
		t.Errorf("expected UserID user-1, got %s", portfolio.UserID)
	}
}

func TestPortfolioUseCase_GetPortfolioSummary(t *testing.T) {
	mockGateway := &MockPortfolioGateway{
		GetPortfolioSummaryFunc: func(ctx context.Context, userID string) (*entity.PortfolioSummary, error) {
			return &entity.PortfolioSummary{
				TotalValue: 1000.0,
				Currency:   "USD",
			}, nil
		},
	}
	uc := NewPortfolioUseCase(mockGateway)

	summary, err := uc.GetPortfolioSummary(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.TotalValue != 1000.0 {
		t.Errorf("expected TotalValue 1000.0, got %f", summary.TotalValue)
	}
}

func TestPortfolioUseCase_GetHoldings(t *testing.T) {
	mockGateway := &MockPortfolioGateway{
		GetHoldingsFunc: func(ctx context.Context, userID string) ([]*entity.Holding, error) {
			return []*entity.Holding{
				{Symbol: "AAPL", Quantity: 10},
			}, nil
		},
	}
	uc := NewPortfolioUseCase(mockGateway)

	holdings, err := uc.GetHoldings(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(holdings) != 1 {
		t.Errorf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", holdings[0].Symbol)
	}
}

func TestPortfolioUseCase_GetPortfolioPerformance(t *testing.T) {
	mockGateway := &MockPortfolioGateway{
		GetPortfolioPerformanceFunc: func(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error) {
			return []*entity.PortfolioPerformancePoint{
				{Value: 100.0, Timestamp: time.Now()},
			}, nil
		},
	}
	uc := NewPortfolioUseCase(mockGateway)

	points, err := uc.GetPortfolioPerformance(context.Background(), "user-1", "1D")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(points) != 1 {
		t.Errorf("expected 1 point, got %d", len(points))
	}
}
