// Package usecase implements the business logic for the portfolio service.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

// PortfolioUsecase defines the business logic for portfolio operations
type PortfolioUsecase interface {
	GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error)
	GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error)
	GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error)
}

// PriceData represents a price point for an asset.
type PriceData struct {
	Price     float64
	Timestamp time.Time
}

type portfolioUsecase struct {
	holdingRepo       domain.HoldingRepository
	marketDataGateway MarketDataGateway
}

// MarketDataGateway defines the interface for fetching current market prices
type MarketDataGateway interface {
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
	GetCurrentPrices(ctx context.Context, symbols []string) (map[string]PriceData, error)
	GetCurrencyRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error)
	GetAssetName(ctx context.Context, symbol string) (string, error)
	GetPriceOnDate(ctx context.Context, symbol string, date time.Time) (float64, error)
	GetCurrencyRateOnDate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (float64, error)
}

// NewPortfolioUsecase creates a new portfolio usecase.
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

	// Enrich holdings with current prices and asset names
	for _, holding := range holdings {
		if data, ok := prices[holding.Symbol]; ok {
			holding.CurrentPrice = data.Price
			holding.PriceLastUpdated = data.Timestamp
		}

		// Fetch asset name (best effort)
		name, err := uc.marketDataGateway.GetAssetName(ctx, holding.Symbol)
		if err == nil {
			holding.AssetName = name
		}
	}

	return holdings, nil
}

// GetPortfolioSummary calculates the portfolio summary for a user
// All values are converted to USD (default currency)
func (uc *portfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	const defaultCurrency = "AUD"

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

	// Calculate totals with currency conversion
	for _, holding := range holdings {
		// Get exchange rate if currency is not AUD
		exchangeRate := 1.0
		if holding.Currency != defaultCurrency {
			rate, err := uc.marketDataGateway.GetCurrencyRate(ctx, holding.Currency, defaultCurrency)
			if err != nil {
				// Log error but continue with rate of 1.0
				// In production, you might want to handle this differently
				fmt.Printf("Warning: Failed to get exchange rate for %s to %s: %v. Using rate 1.0\n",
					holding.Currency, defaultCurrency, err)
			} else {
				exchangeRate = rate
			}
		}

		// Convert to default currency
		costBasis := holding.Quantity * holding.AverageCost * exchangeRate
		currentValue := holding.Quantity * holding.CurrentPrice * exchangeRate

		summary.TotalCost += costBasis
		summary.TotalValue += currentValue
	}

	// Calculate gain/loss
	summary.GainLoss = summary.TotalValue - summary.TotalCost

	// Calculate percentage (avoid division by zero)
	if summary.TotalCost > 0 {
		summary.GainLossPct = (summary.GainLoss / summary.TotalCost) * 100
	}

	// Calculate Day Change
	// Determine the reference date for "yesterday"
	// If we have recent price data, use that as the anchor.
	// Otherwise, default to time.Now()
	// anchorDate := time.Now()
	/*
		for _, h := range holdings {
			if !h.PriceLastUpdated.IsZero() && h.PriceLastUpdated.After(anchorDate.Add(-24*time.Hour)) {
				// If we have a price update from "today" (or recent), use it.
				// But wait, if the price is from "yesterday" (UTC), we want to compare with "day before yesterday".
				// Let's use the latest price timestamp as the "current" time.
				if h.PriceLastUpdated.After(anchorDate) {
					// This shouldn't happen if anchorDate is Now(), unless clock skew.
				}
				// We want the max timestamp?
			}
		}
	*/
	// Actually, simpler: Find the latest PriceLastUpdated.
	var latestUpdate time.Time
	for _, h := range holdings {
		if h.PriceLastUpdated.After(latestUpdate) {
			latestUpdate = h.PriceLastUpdated
		}
	}

	// If we have a valid latest update, use it to calculate "yesterday".
	// If latestUpdate is zero (no prices), use Now().
	if latestUpdate.IsZero() {
		latestUpdate = time.Now()
	}

	// Calculate yesterday relative to the latest data we have.
	// e.g. if data is from Dec 1, we want comparison with Nov 30.
	// if data is from Dec 2, we want comparison with Dec 1.
	yesterday := latestUpdate.Add(-24 * time.Hour)
	prevSummary, err := uc.GetHistoricalPortfolioSummary(ctx, userID, yesterday)
	if err != nil {
		// Log warning but don't fail the request
		fmt.Printf("Warning: Failed to get historical summary for day change calculation: %v\n", err)
	} else {
		summary.DayChange = summary.TotalValue - prevSummary.TotalValue
		if prevSummary.TotalValue > 0 {
			summary.DayChangePct = (summary.DayChange / prevSummary.TotalValue) * 100
		}
	}

	// Set the currency of the summary (default currency)
	summary.Currency = defaultCurrency
	return summary, nil
}

// GetHistoricalPortfolioSummary calculates the portfolio summary for a user at a specific date
func (uc *portfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	const defaultCurrency = "AUD"

	// Get holdings (Note: using current holdings as proxy for historical holdings)
	holdings, err := uc.holdingRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	summary := &domain.PortfolioSummary{
		UserID:      userID,
		TotalValue:  0,
		TotalCost:   0,
		GainLoss:    0,
		GainLossPct: 0,
		LastUpdated: date,
	}

	// Calculate totals with currency conversion
	for _, holding := range holdings {
		// Get historical price
		price, err := uc.marketDataGateway.GetPriceOnDate(ctx, holding.Symbol, date)
		if err != nil {
			// Log warning and skip or use 0?
			// For backfilling, missing price is critical.
			// But maybe we can skip this asset?
			// Let's log and continue, effectively treating value as 0 for this asset on that day.
			fmt.Printf("Warning: Failed to get historical price for %s on %s: %v\n", holding.Symbol, date.Format("2006-01-02"), err)
			continue
		}

		// Get exchange rate if currency is not AUD
		// Now using HISTORICAL exchange rate for the specified date
		exchangeRate := 1.0
		if holding.Currency != defaultCurrency {
			rate, err := uc.marketDataGateway.GetCurrencyRateOnDate(ctx, holding.Currency, defaultCurrency, date)
			if err != nil {
				fmt.Printf("Warning: Failed to get historical exchange rate for %s to %s on %s: %v. Using rate 1.0\n",
					holding.Currency, defaultCurrency, date.Format("2006-01-02"), err)
			} else {
				exchangeRate = rate
			}
		}

		// Convert to default currency
		// Cost basis is historical (average cost), assuming it hasn't changed much or using current.
		// CurrentValue is Quantity * HistoricalPrice * ExchangeRate
		costBasis := holding.Quantity * holding.AverageCost * exchangeRate
		currentValue := holding.Quantity * price * exchangeRate

		summary.TotalCost += costBasis
		summary.TotalValue += currentValue
	}

	// Calculate gain/loss
	summary.GainLoss = summary.TotalValue - summary.TotalCost

	// Calculate percentage (avoid division by zero)
	if summary.TotalCost > 0 {
		summary.GainLossPct = (summary.GainLoss / summary.TotalCost) * 100
	}

	// Set the currency of the summary
	summary.Currency = defaultCurrency
	return summary, nil
}
