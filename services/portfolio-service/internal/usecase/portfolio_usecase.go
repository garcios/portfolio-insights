package usecase

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

// PortfolioUsecase defines the business logic for portfolio operations
type PortfolioUsecase interface {
	GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error)
	GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error)
}

type portfolioUsecase struct {
	holdingRepo       domain.HoldingRepository
	marketDataGateway MarketDataGateway
}

// MarketDataGateway defines the interface for fetching current market prices
type MarketDataGateway interface {
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
	GetCurrentPrices(ctx context.Context, symbols []string) (map[string]float64, error)
}

func NewPortfolioUsecase(holdingRepo domain.HoldingRepository, marketDataGateway MarketDataGateway) PortfolioUsecase {
	return &portfolioUsecase{
		holdingRepo:       holdingRepo,
		marketDataGateway: marketDataGateway,
	}
}

// GetHoldings retrieves all holdings for a user with current market prices
func (uc *portfolioUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	// Get holdings from repository
	holdings, err := uc.holdingRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	if len(holdings) == 0 {
		return []*domain.Holding{}, nil
	}

	// Extract symbols for price lookup
	symbols := make([]string, len(holdings))
	for i, holding := range holdings {
		symbols[i] = holding.Symbol
	}

	// Get current prices for all symbols
	prices, err := uc.marketDataGateway.GetCurrentPrices(ctx, symbols)
	if err != nil {
		// If we can't get prices, return holdings without current prices
		// In production, you might want to handle this differently
		return holdings, nil
	}

	// Enrich holdings with current prices
	for _, holding := range holdings {
		if price, ok := prices[holding.Symbol]; ok {
			holding.CurrentPrice = price
		}
	}

	return holdings, nil
}

// GetPortfolioSummary calculates the portfolio summary for a user
func (uc *portfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	// Get holdings with current prices
	holdings, err := uc.GetHoldings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	summary := &domain.PortfolioSummary{
		UserID:      userID,
		TotalValue:  0,
		TotalCost:   0,
		GainLoss:    0,
		GainLossPct: 0,
	}

	// Calculate totals
	for _, holding := range holdings {
		costBasis := holding.Quantity * holding.AverageCost
		currentValue := holding.Quantity * holding.CurrentPrice

		summary.TotalCost += costBasis
		summary.TotalValue += currentValue
	}

	// Calculate gain/loss
	summary.GainLoss = summary.TotalValue - summary.TotalCost

	// Calculate percentage (avoid division by zero)
	if summary.TotalCost > 0 {
		summary.GainLossPct = (summary.GainLoss / summary.TotalCost) * 100
	}

	return summary, nil
}
