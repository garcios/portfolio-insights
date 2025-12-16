// Package gateway defines the gateway interfaces.
package gateway

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

// PortfolioGateway defines the interface for interacting with the portfolio service
type PortfolioGateway interface {
	// GetPortfolio retrieves a portfolio by user ID
	GetPortfolio(ctx context.Context, userID string) (*entity.Portfolio, error)

	// GetPortfolioSummary retrieves portfolio summary metrics
	GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *string) (*entity.PortfolioSummary, error)

	// GetHoldings retrieves all holdings for a user
	GetHoldings(ctx context.Context, userID string) ([]*entity.Holding, error)

	// GetPortfolioPerformance retrieves historical performance data
	GetPortfolioPerformance(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error)
}
