// Package usecase implements the business logic for the portfolio service.
package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
)

// BackfillResult represents the result of a backfill operation.
type BackfillResult struct {
	Created       int
	Skipped       int
	Errors        int
	ErrorMessages []string
	Status        string
}

// PortfolioUsecase defines the business logic for portfolio operations
type PortfolioUsecase interface {
	GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error)
	GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error)
	GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error)
	BackfillPortfolioHistory(ctx context.Context, userIDs []string, startDate, endDate time.Time, dryRun bool) BackfillResult
}

// PriceData represents a price point for an asset.
type PriceData struct {
	Price     float64
	Timestamp time.Time
}

type portfolioUsecase struct {
	holdingRepo       domain.HoldingRepository
	historyRepo       domain.PortfolioHistoryRepository
	cashBalanceRepo   domain.CashBalanceRepository
	marketDataGateway MarketDataGateway
	transactionClient transactionpb.TransactionServiceClient
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
func NewPortfolioUsecase(
	holdingRepo domain.HoldingRepository,
	historyRepo domain.PortfolioHistoryRepository,
	cashBalanceRepo domain.CashBalanceRepository,
	marketDataGateway MarketDataGateway,
	transactionClient transactionpb.TransactionServiceClient,
) PortfolioUsecase {
	return &portfolioUsecase{
		holdingRepo:       holdingRepo,
		historyRepo:       historyRepo,
		cashBalanceRepo:   cashBalanceRepo,
		marketDataGateway: marketDataGateway,
		transactionClient: transactionClient,
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
// GetPortfolioSummary calculates the portfolio summary for a user
// All values are converted to USD (default currency)
func (uc *portfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error) {
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

	// Calculate totals with currency conversion (Holdings)
	for _, holding := range holdings {
		// Get exchange rate if currency is not AUD
		exchangeRate := 1.0
		if holding.Currency != defaultCurrency {
			rate, err := uc.marketDataGateway.GetCurrencyRate(ctx, holding.Currency, defaultCurrency)
			if err != nil {
				// Log error but continue with rate of 1.0
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

	// Snapshot Holdings Value before adding cash
	holdingsValue := summary.TotalValue

	// Add Cash Balance to Total Value (Current)
	cashBalances, err := uc.cashBalanceRepo.ListByUser(userID)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch cash balances: %v\n", err)
	} else {
		for _, cb := range cashBalances {
			value := cb.Balance
			if cb.Currency != defaultCurrency {
				rate, err := uc.marketDataGateway.GetCurrencyRate(ctx, cb.Currency, defaultCurrency)
				if err == nil {
					value *= rate
				}
			}
			summary.TotalValue += value
		}
	}

	// Calculate Current Capital Gain
	currentCapitalGain := holdingsValue - summary.TotalCost

	// Fetch Transactions for Net Invested and Dividends
	netInvested := 0.0
	dividends := 0.0

	// Default start/end dates for the entire history
	summary.StartDate = time.Time{}
	summary.EndDate = time.Now()
	if endDate != nil {
		summary.EndDate = *endDate
	}

	// If startDate is set, we need the historical snapshot
	var startTotalValue, startCapitalGain float64
	if startDate != nil {
		summary.StartDate = *startDate
		startSnapshot, err := uc.GetHistoricalPortfolioSummary(ctx, userID, *startDate)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch historical snapshot for %s: %v. Assuming 0.\n", startDate.Format("2006-01-02"), err)
		} else {
			startTotalValue = startSnapshot.TotalValue
			startCapitalGain = startSnapshot.GainLoss // In historical summary, GainLoss is CapitalGain (TotalValue - TotalCost)
		}
	} else {
		// If no startDate, find the first transaction date
		var firstTxnDate time.Time
		pageToken := ""
		for {
			resp, err := uc.transactionClient.ListTransactions(ctx, &transactionpb.ListTransactionsRequest{
				Parent:    fmt.Sprintf("users/%s", userID),
				PageSize:  1000,
				PageToken: pageToken,
			})
			if err != nil {
				break
			}
			for _, txn := range resp.Transactions {
				executedAt := txn.ExecutedAt.AsTime()
				if firstTxnDate.IsZero() || executedAt.Before(firstTxnDate) {
					firstTxnDate = executedAt
				}
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		summary.StartDate = firstTxnDate
	}

	// Fetch transactions again (efficient enough for now, or could optimize to single pass if needed)
	// We need to filter transactions within [StartDate, EndDate]
	var pageToken string
	for {
		resp, err := uc.transactionClient.ListTransactions(ctx, &transactionpb.ListTransactionsRequest{
			Parent:    fmt.Sprintf("users/%s", userID),
			PageSize:  1000,
			PageToken: pageToken,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to fetch transactions: %v\n", err)
			break
		}

		for _, txn := range resp.Transactions {
			executedAt := txn.ExecutedAt.AsTime()

			// Filter by date range
			if !summary.StartDate.IsZero() && executedAt.Before(summary.StartDate) {
				continue
			}
			if endDate != nil && executedAt.After(*endDate) {
				continue
			}

			// Handle cash flows
			amount := 0.0
			if txn.Amount != nil {
				amount = *txn.Amount
			}

			// Convert amount to default currency if needed
			if txn.PriceCurrency != defaultCurrency {
				rate, err := uc.marketDataGateway.GetCurrencyRateOnDate(ctx, txn.PriceCurrency, defaultCurrency, executedAt)
				if err != nil {
					fmt.Printf("Warning: Failed to get exchange rate for transaction %s (%s to %s) on %s: %v. Using rate 1.0\n",
						txn.Name, txn.PriceCurrency, defaultCurrency, executedAt.Format("2006-01-02"), err)
				} else {
					amount *= rate
				}
			}

			switch txn.Type {
			case "DEP":
				netInvested += amount
			case "WIT":
				netInvested -= amount
			case "DIV":
				dividends += amount
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Calculate Deltas
	summary.CapitalGain = currentCapitalGain - startCapitalGain
	summary.Dividends = dividends

	// Total Gain for Period = (EndValue - StartValue) - NetInvestedPeriod
	// EndValue is Current TotalValue (assuming endDate is effectively now or we use current value as requested)
	totalGain := (summary.TotalValue - startTotalValue) - netInvested

	// Currency Gain = Total Gain - Capital Gain - Dividends
	summary.CurrencyGain = totalGain - summary.CapitalGain - summary.Dividends

	// Percentages
	denominator := netInvested
	// Fallback for denominator if NetInvested is 0 (e.g. no flows in period), use Start Value
	if denominator == 0 {
		denominator = startTotalValue
	}
	// Fallback if Period Start was 0, use Total Cost as proxy for "invested"
	if denominator == 0 {
		denominator = summary.TotalCost
	}

	if summary.TotalCost > 0 {
		summary.CapitalGainPct = (summary.CapitalGain / summary.TotalCost) * 100
	}
	if denominator > 0 {
		summary.CurrencyGainPct = (summary.CurrencyGain / denominator) * 100
		summary.DividendsPct = (summary.Dividends / denominator) * 100
	}

	// Annualized Returns (CAGR) if period > 1 year
	years := 0.0
	if !summary.StartDate.IsZero() {
		years = time.Since(summary.StartDate).Hours() / (24 * 365.25)
	}

	if years > 1 {
		// Applying CAGR logic to the components relative to the denominator
		if summary.TotalCost > 0 && summary.TotalCost+summary.CapitalGain > 0 {
			ratio := (summary.TotalCost + summary.CapitalGain) / summary.TotalCost
			summary.CapitalGainPct = (math.Pow(ratio, 1.0/years) - 1) * 100
		}
		if denominator > 0 && denominator+summary.CurrencyGain > 0 {
			ratio := (denominator + summary.CurrencyGain) / denominator
			summary.CurrencyGainPct = (math.Pow(ratio, 1.0/years) - 1) * 100
		}
		if denominator > 0 && denominator+summary.Dividends > 0 {
			ratio := (denominator + summary.Dividends) / denominator
			summary.DividendsPct = (math.Pow(ratio, 1.0/years) - 1) * 100
		}
	}

	// Overall Gain/Loss fields
	summary.GainLoss = totalGain
	if denominator > 0 {
		summary.GainLossPct = (summary.GainLoss / denominator) * 100
	}

	// Day Change (Keep existing logic)
	var latestUpdate time.Time
	for _, h := range holdings {
		if h.PriceLastUpdated.After(latestUpdate) {
			latestUpdate = h.PriceLastUpdated
		}
	}
	if latestUpdate.IsZero() {
		latestUpdate = time.Now()
	}
	yesterday := latestUpdate.Add(-24 * time.Hour)
	prevSummary, err := uc.GetHistoricalPortfolioSummary(ctx, userID, yesterday)
	if err != nil {
		fmt.Printf("Warning: Failed to get historical summary for day change calculation: %v\n", err)
	} else {
		summary.DayChange = summary.TotalValue - prevSummary.TotalValue
		if prevSummary.TotalValue > 0 {
			summary.DayChangePct = (summary.DayChange / prevSummary.TotalValue) * 100
		}
	}

	summary.Currency = defaultCurrency
	return summary, nil
}

// GetHistoricalPortfolioSummary calculates the portfolio summary for a user at a specific date
func (uc *portfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	const defaultCurrency = "AUD"

	// Try to get from history repo first
	// We check for a snapshot at the exact start of the day (or the provided time)
	// Since we care about "end of previous day" usually, or "start of this day".
	// The backfill creates snapshots at 00:00:00 UTC usually.
	// Let's assume the date passed is normalized.

	// Check if snapshot exists
	exists, err := uc.historyRepo.SnapshotExists(ctx, userID, date)
	if err == nil && exists {
		snapshots, err := uc.historyRepo.GetHistory(ctx, userID, date, date)
		if err == nil && len(snapshots) > 0 {
			// Found snapshot
			s := snapshots[0]
			return &domain.PortfolioSummary{
				UserID:      s.UserID,
				TotalValue:  s.TotalValue,
				TotalCost:   s.TotalCostBasis,
				GainLoss:    s.TotalValue - s.TotalCostBasis,
				GainLossPct: 0, // Recalculate if needed: (val-cost)/cost * 100
				LastUpdated: s.Timestamp,
			}, nil
		}
	}

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

// BackfillPortfolioHistory runs the backfill process for the given users and date range
func (uc *portfolioUsecase) BackfillPortfolioHistory(
	ctx context.Context,
	userIDs []string,
	startDate, endDate time.Time,
	dryRun bool,
) BackfillResult {
	result := BackfillResult{
		ErrorMessages: []string{},
	}

	for _, userID := range userIDs {
		// Backfill for each day in range
		for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
			// Check if snapshot already exists
			exists, _ := uc.historyRepo.SnapshotExists(ctx, userID, date)
			if exists {
				result.Skipped++
				continue
			}

			if dryRun {
				// In a real logger we would log this
				result.Created++
				continue
			}

			// Create snapshot for this date
			created, err := uc.createHistoricalSnapshot(ctx, userID, date)
			if err != nil {
				result.Errors++
				result.ErrorMessages = append(result.ErrorMessages,
					fmt.Sprintf("user=%s date=%s: %v", userID, date.Format("2006-01-02"), err),
				)
				continue
			}

			if created {
				result.Created++
			}
		}
	}

	// Determine overall status
	if result.Errors == 0 {
		result.Status = "success"
	} else if result.Created > 0 {
		result.Status = "partial"
	} else {
		result.Status = "failed"
	}

	return result
}

func (uc *portfolioUsecase) createHistoricalSnapshot(
	ctx context.Context,
	userID string,
	date time.Time,
) (bool, error) {
	// Get historical portfolio summary for this user
	summary, err := uc.GetHistoricalPortfolioSummary(ctx, userID, date)
	if err != nil {
		return false, fmt.Errorf("failed to get historical summary: %w", err)
	}

	// Skip if total value is 0 (likely missing price data)
	if summary.TotalValue == 0 {
		return false, nil
	}

	// Create snapshot with the specified date
	snapshot := &domain.PortfolioSnapshot{
		UserID:         userID,
		TotalValue:     summary.TotalValue,
		TotalCostBasis: summary.TotalCost,
		Timestamp:      date,
	}

	return true, uc.historyRepo.CreateSnapshot(ctx, snapshot)
}
