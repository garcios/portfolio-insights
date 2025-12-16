// Package usecase implements the business logic for the portfolio service.
package usecase

import (
	"context"
	"fmt"
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
// GetPortfolioSummary calculates the portfolio summary for a user
// All values are converted to USD (default currency)
// GetPortfolioSummary calculates the portfolio summary for a user
// All values are converted to USD (default currency)
func (uc *portfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error) {
	const defaultCurrency = "AUD"

	// 1. Determine Dates
	summary := &domain.PortfolioSummary{
		UserID:      userID,
		Currency:    defaultCurrency,
		LastUpdated: time.Now(),
	}

	calcStartDate := time.Time{}
	calcEndDate := time.Now()

	if startDate != nil {
		calcStartDate = *startDate
		summary.StartDate = *startDate
	}
	if endDate != nil {
		calcEndDate = *endDate
		summary.EndDate = calcEndDate
	} else {
		summary.EndDate = calcEndDate
	}

	// 2. Perform Replay
	replayResult, err := uc.calculatePeriodGains(ctx, userID, calcStartDate, calcEndDate, defaultCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate period gains: %w", err)
	}

	summary.Dividends = replayResult.Dividends
	// NetInvested is internal metric, mostly for Total Gain calc

	// 3. Calculate Unrealized Gains (on FinalPositions)
	var totalUnrealizedCapital float64
	var totalUnrealizedFX float64
	var endTotalCost float64
	var endHoldingsValue float64

	// Determine if we are "live" or "historical"
	isCurrent := endDate == nil || calcEndDate.After(time.Now().Add(-12*time.Hour))

	// Collect symbols
	var symbols []string
	for sym, pos := range replayResult.FinalPositions {
		if pos.Quantity > 0.000001 { // active position
			symbols = append(symbols, sym)
		}
	}

	// Fetch Prices and track Recency
	currentPrices := make(map[string]float64)
	var latestUpdate time.Time

	if isCurrent && len(symbols) > 0 {
		// Batch Fetch
		prices, err := uc.marketDataGateway.GetCurrentPrices(ctx, symbols)
		if err == nil {
			for s, data := range prices {
				currentPrices[s] = data.Price
				if data.Timestamp.After(latestUpdate) {
					latestUpdate = data.Timestamp
				}
			}
		}
	}

	if isCurrent && !latestUpdate.IsZero() {
		summary.LastUpdated = latestUpdate
	}

	for sym, pos := range replayResult.FinalPositions {
		if pos.Quantity <= 0.000001 {
			continue
		}

		var price float64
		// Get price
		if isCurrent {
			if p, ok := currentPrices[sym]; ok {
				price = p
			} else {
				// Fallback
				price, _ = uc.marketDataGateway.GetCurrentPrice(ctx, sym)
				// Fallback timestamp? Ignore. End of Day logic will use latestUpdate or Now defaults.
			}
		} else {
			price, _ = uc.marketDataGateway.GetPriceOnDate(ctx, sym, calcEndDate)
		}

		// Get FX
		fxRate := 1.0
		if pos.Currency != defaultCurrency {
			if isCurrent {
				r, err := uc.marketDataGateway.GetCurrencyRate(ctx, pos.Currency, defaultCurrency)
				if err == nil {
					fxRate = r
				}
			} else {
				r, err := uc.marketDataGateway.GetCurrencyRateOnDate(ctx, pos.Currency, defaultCurrency, calcEndDate)
				if err == nil {
					fxRate = r
				}
			}
		}

		// Calculate Metrics
		qty := pos.Quantity

		// Value = Price * Qty * FXCurrent
		val := price * qty * fxRate
		endHoldingsValue += val

		// Cost (Base) = AverageCost * Qty
		baseCost := pos.AverageCost * qty
		endTotalCost += baseCost

		// Breakdown
		// Effective Buy FX = AvgCostBase / AvgForeignCost
		avgFXBuy := 1.0
		if pos.AverageForeignCost != 0 {
			avgFXBuy = pos.AverageCost / pos.AverageForeignCost
		}

		// Unrealized Capital Gain = (CurrentPrice - AvgForeignCost) * Qty * AvgFXBuy
		unrealizedCap := (price - pos.AverageForeignCost) * qty * avgFXBuy

		// Unrealized FX Gain = (CurrentPrice * Qty) * (FXCurrent - AvgFXBuy)
		unrealizedFX := (price * qty) * (fxRate - avgFXBuy)

		totalUnrealizedCapital += unrealizedCap
		totalUnrealizedFX += unrealizedFX
	}

	// 4. Assemble Summary
	summary.TotalValue = endHoldingsValue + replayResult.CurrentCash
	summary.TotalCost = endTotalCost // Usually cost of HOLDINGS. Cash has no cost basis? Or Cost=Value.
	// If TotalCost excludes cash, then UnrealizedGain should exclude cash.
	// summary.GainLoss (Total Period Gain) = FinalValue - StartValue - NetInvested.

	// Start Value Calculation
	var startTotalValue float64
	if !calcStartDate.IsZero() {
		// We need historical value at start date.
		// Use existing method?
		snap, err := uc.GetHistoricalPortfolioSummary(ctx, userID, calcStartDate)
		if err == nil {
			startTotalValue = snap.TotalValue
		}
	}

	totalGain := (summary.TotalValue - startTotalValue) - replayResult.NetInvested
	summary.GainLoss = totalGain

	// Capital Gain = Realized Capital + Unrealized Capital
	summary.CapitalGain = replayResult.RealizedCapitalGain + totalUnrealizedCapital

	// Currency Gain = Realized FX + Unrealized FX
	summary.CurrencyGain = replayResult.RealizedFXGain + totalUnrealizedFX

	// Percentages
	if summary.TotalCost > 0 {
		summary.CapitalGainPct = (summary.CapitalGain / summary.TotalCost) * 100
		summary.CurrencyGainPct = (summary.CurrencyGain / summary.TotalCost) * 100
	}
	// For Total Return %, usually (TotalGain / NetInvested) * 100? or (TotalGain / StartValue)?
	// If NetInvested > 0:
	divisor := replayResult.NetInvested
	if divisor == 0 && startTotalValue > 0 {
		divisor = startTotalValue
	}
	if divisor > 0 {
		summary.GainLossPct = (totalGain / divisor) * 100
	}

	// Day Change
	// Simplified: Use simple historical lookup for yesterday
	yesterday := calcEndDate.Add(-24 * time.Hour)
	prevSummary, err := uc.GetHistoricalPortfolioSummary(ctx, userID, yesterday)
	if err == nil {
		summary.DayChange = summary.TotalValue - prevSummary.TotalValue
		if prevSummary.TotalValue > 0 {
			summary.DayChangePct = (summary.DayChange / prevSummary.TotalValue) * 100
		}
	}

	return summary, nil
}

// AssetPosition tracks the cost basis of an asset for realized gain calculation
type AssetPosition struct {
	Quantity           float64
	AverageCost        float64 // In Base Currency (Total Cost / Qty)
	AverageForeignCost float64 // In Foreign Currency (Total Foreign Cost / Qty)
	Currency           string
}

// ReplayResult contains the results of replaying transactions to calculate realized gains and final positions.
type ReplayResult struct {
	RealizedTotalGain   float64
	RealizedCapitalGain float64
	RealizedFXGain      float64
	Dividends           float64
	NetInvested         float64
	CurrentCash         float64
	FinalPositions      map[string]*AssetPosition
}

// Helper to calculate realized gains and net invested by replaying transactions
func (uc *portfolioUsecase) calculatePeriodGains(
	ctx context.Context,
	userID string,
	startDate, endDate time.Time,
	defaultCurrency string,
) (*ReplayResult, error) {
	result := &ReplayResult{
		FinalPositions: make(map[string]*AssetPosition),
	}

	// 1. Fetch ALL transactions
	var allTxns []*transactionpb.Transaction
	pageToken := ""
	for {
		resp, err := uc.transactionClient.ListTransactions(ctx, &transactionpb.ListTransactionsRequest{
			Parent:    fmt.Sprintf("users/%s", userID),
			PageSize:  1000,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list transactions: %w", err)
		}
		allTxns = append(allTxns, resp.Transactions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// 2. Sort transactions by ExecutedAt
	for i := 0; i < len(allTxns); i++ {
		for j := i + 1; j < len(allTxns); j++ {
			ti := allTxns[i].ExecutedAt.AsTime()
			tj := allTxns[j].ExecutedAt.AsTime()
			if ti.After(tj) {
				allTxns[i], allTxns[j] = allTxns[j], allTxns[i]
			}
		}
	}

	// 3. Replay
	for _, txn := range allTxns {
		executedAt := txn.ExecutedAt.AsTime()

		if !endDate.IsZero() && executedAt.After(endDate) {
			continue
		}

		inPeriod := true
		if !startDate.IsZero() && executedAt.Before(startDate) {
			inPeriod = false
		}

		// Helper to get transaction amount in Default Currency
		// Returns (AmountInDefault, ExchangeRateUsed, error)
		getAmountAndRate := func(amt float64, currency string) (float64, float64, error) {
			if currency == defaultCurrency {
				return amt, 1.0, nil
			}
			rate, err := uc.marketDataGateway.GetCurrencyRateOnDate(ctx, currency, defaultCurrency, executedAt)
			if err != nil {
				// Log error? For now assume 1.0 implies error handling upstream or rough justice
				return amt, 1.0, nil
			}
			return amt * rate, rate, nil
		}

		switch txn.Type {
		case "DEP":
			if txn.Amount != nil {
				val, _, _ := getAmountAndRate(*txn.Amount, txn.PriceCurrency)
				result.CurrentCash += val // Always update cash state
				if inPeriod {
					result.NetInvested += val
				}
			}
		case "WIT":
			if txn.Amount != nil {
				val, _, _ := getAmountAndRate(*txn.Amount, txn.PriceCurrency)
				result.CurrentCash -= val
				if inPeriod {
					result.NetInvested -= val
				}
			}
		case "DIV":
			if txn.Amount != nil {
				val, _, _ := getAmountAndRate(*txn.Amount, txn.PriceCurrency)
				result.CurrentCash += val
				if inPeriod {
					result.Dividends += val
				}
			}
		case "BUY":
			if txn.Symbol != nil && txn.Quantity != nil && txn.PricePerShare != nil {
				symbol := *txn.Symbol
				qty := *txn.Quantity
				price := *txn.PricePerShare

				priceInDefault, _, _ := getAmountAndRate(price, txn.PriceCurrency)
				totalCost := qty * priceInDefault

				result.CurrentCash -= totalCost

				pos, exists := result.FinalPositions[symbol]
				if !exists {
					pos = &AssetPosition{
						Quantity:           0,
						AverageCost:        0,
						AverageForeignCost: 0,
						Currency:           txn.PriceCurrency,
					}
					result.FinalPositions[symbol] = pos
				}

				currentTotalCost := (pos.Quantity * pos.AverageCost) + (qty * priceInDefault)
				currentTotalForeignCost := (pos.Quantity * pos.AverageForeignCost) + (qty * price)

				pos.Quantity += qty
				if pos.Quantity > 0 {
					pos.AverageCost = currentTotalCost / pos.Quantity
					pos.AverageForeignCost = currentTotalForeignCost / pos.Quantity
				}
			}
		case "SELL":
			if txn.Symbol != nil && txn.Quantity != nil && txn.PricePerShare != nil {
				symbol := *txn.Symbol
				qty := *txn.Quantity
				price := *txn.PricePerShare // Foreign Price

				priceInDefault, fxSell, _ := getAmountAndRate(price, txn.PriceCurrency)
				totalProceeds := qty * priceInDefault

				result.CurrentCash += totalProceeds

				pos, exists := result.FinalPositions[symbol]
				if exists {
					if inPeriod {
						// Total Realized Gain
						gain := (priceInDefault - pos.AverageCost) * qty
						result.RealizedTotalGain += gain

						// Detailed Breakdown
						avgFXBuy := 1.0
						if pos.AverageForeignCost != 0 {
							avgFXBuy = pos.AverageCost / pos.AverageForeignCost
						}

						capitalGain := (price - pos.AverageForeignCost) * qty * avgFXBuy
						fxGain := (price * qty) * (fxSell - avgFXBuy)

						result.RealizedCapitalGain += capitalGain
						result.RealizedFXGain += fxGain
					}
					pos.Quantity -= qty
					if pos.Quantity < 0 {
						pos.Quantity = 0
					} // Should not happen
				}
			}
		}
	}
	return result, nil
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
