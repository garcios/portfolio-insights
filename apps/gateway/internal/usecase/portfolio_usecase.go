package usecase

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
)

// PortfolioUseCase handles portfolio-related business logic
type PortfolioUseCase struct {
	portfolioGateway gateway.PortfolioGateway
}

// NewPortfolioUseCase creates a new PortfolioUseCase
func NewPortfolioUseCase(portfolioGateway gateway.PortfolioGateway) *PortfolioUseCase {
	return &PortfolioUseCase{
		portfolioGateway: portfolioGateway,
	}
}

// GetPortfolio retrieves a complete portfolio with summary and holdings
func (uc *PortfolioUseCase) GetPortfolio(ctx context.Context, userID string) (*entity.Portfolio, error) {
	// Create basic portfolio structure
	portfolio := entity.NewPortfolio(userID, userID, "My Portfolio")

	// Note: Summary and Holdings are loaded lazily via field resolvers in GraphQL
	// This allows for efficient data fetching based on what the client requests
	return portfolio, nil
}

// GetPortfolioSummary retrieves portfolio summary metrics
func (uc *PortfolioUseCase) GetPortfolioSummary(ctx context.Context, userID string) (*entity.PortfolioSummary, error) {
	return uc.portfolioGateway.GetPortfolioSummary(ctx, userID)
}

// GetHoldings retrieves all holdings for a user
func (uc *PortfolioUseCase) GetHoldings(ctx context.Context, userID string) ([]*entity.Holding, error) {
	return uc.portfolioGateway.GetHoldings(ctx, userID)
}

// GetPortfolioPerformance retrieves historical performance data
func (uc *PortfolioUseCase) GetPortfolioPerformance(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error) {
	// Here you could add validation for the period parameter
	// e.g., ensure it's one of: "1D", "1W", "1M", "3M", "1Y", "ALL"
	return uc.portfolioGateway.GetPortfolioPerformance(ctx, userID, period)
}
