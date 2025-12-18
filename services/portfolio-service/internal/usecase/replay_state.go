package usecase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
)

// ReplayState tracks the portfolio state during transaction replay.
type ReplayState struct {
	Holdings      map[string]*domain.AssetPosition
	Cash          map[string]float64
	RealizedGains map[string]float64
	NetInvested   float64
}

// NewReplayState creates a new empty ReplayState.
func NewReplayState() *ReplayState {
	return &ReplayState{
		Holdings:      make(map[string]*domain.AssetPosition),
		Cash:          make(map[string]float64),
		RealizedGains: make(map[string]float64),
		NetInvested:   0,
	}
}

// HydrateFrom loads the optimized JSON snapshot into the working memory.
func (rs *ReplayState) HydrateFrom(snap *domain.PortfolioSnapshot) error {
	for symbol, h := range snap.State.Holdings {
		qty, err := strconv.ParseFloat(h.Quantity, 64)
		if err != nil {
			return fmt.Errorf("failed to parse quantity for %s: %w", symbol, err)
		}
		cost, err := strconv.ParseFloat(h.CostBasis, 64)
		if err != nil {
			return fmt.Errorf("failed to parse cost basis for %s: %w", symbol, err)
		}

		rs.Holdings[symbol] = &domain.AssetPosition{
			Quantity: qty,
			Currency: h.Currency,
		}
		if qty > 0 {
			rs.Holdings[symbol].AverageCost = cost / qty
			// Approximation: We assume ForeignCost ~ Cost if we don't have better data in snapshot
			// For precise accounting, Snapshot should store ForeignCost too.
			// Assuming H.CostBasis is in DefaultCurrency (AUD?)
			rs.Holdings[symbol].AverageForeignCost = cost / qty
		}
	}

	for currency, amountStr := range snap.State.Cash {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("failed to parse cash for %s: %w", currency, err)
		}
		rs.Cash[currency] = amount
	}

	for currency, amountStr := range snap.State.RealizedGains {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("failed to parse realized gains for %s: %w", currency, err)
		}
		rs.RealizedGains[currency] = amount
	}

	if snap.State.NetInvested != "" {
		ni, err := strconv.ParseFloat(snap.State.NetInvested, 64)
		if err == nil {
			rs.NetInvested = ni
		}
	}

	return nil
}

// Apply updates the state based on a single transaction event.
// rate allows converting transaction logic to the default currency tracking context.
// For BUY/SELL, txn.Price is in txn.PriceCurrency. rate = Rate(PriceCurrency -> DefaultCurrency).
func (rs *ReplayState) Apply(txn *transactionpb.Transaction, rate float64, defaultCurrency string) error {

	// Helper for safe pointer dereference
	getFloat := func(f *float64) float64 {
		if f == nil {
			return 0
		}
		return *f
	}

	switch txn.Type {
	case "DEP":
		if txn.Amount != nil {
			// Deposit increases cash
			amount := *txn.Amount
			rs.Cash[txn.PriceCurrency] += amount
			rs.NetInvested += amount * rate
		}
	case "WIT":
		if txn.Amount != nil {
			// Withdrawal decreases cash
			amount := *txn.Amount
			rs.Cash[txn.PriceCurrency] -= amount
			rs.NetInvested -= amount * rate
		}
	case "BUY":
		if txn.Symbol != nil && txn.Quantity != nil && txn.PricePerShare != nil {
			symbol := *txn.Symbol
			qty := *txn.Quantity
			price := getFloat(txn.PricePerShare) // Foreign Price

			// Total Cost in Default Currency
			// rate is FX rate (Foreign -> Default)
			priceInDefault := price * rate
			totalCostDefault := qty * priceInDefault
			totalCostForeign := qty * price

			// Deduct Cash (in Transaction Currency usually, or Settlement Currency?)
			// Typically we pay in the currency of the trade.
			rs.Cash[txn.PriceCurrency] -= totalCostForeign * 1.0 // Assuming Cash in PriceCurrency

			// Update Holding
			pos, exists := rs.Holdings[symbol]
			if !exists {
				pos = &domain.AssetPosition{
					Quantity:           0,
					AverageCost:        0,
					AverageForeignCost: 0,
					Currency:           txn.PriceCurrency,
				}
				rs.Holdings[symbol] = pos
			}

			// Weighted Average Cost Calculation
			currentTotalCost := (pos.Quantity * pos.AverageCost) + totalCostDefault
			currentTotalForeignCost := (pos.Quantity * pos.AverageForeignCost) + totalCostForeign

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
			price := getFloat(txn.PricePerShare)

			// priceInDefault = price * rate (Rate at time of SELL)
			priceInDefault := price * rate

			// Proceeds
			totalProceedsForeign := qty * price
			rs.Cash[txn.PriceCurrency] += totalProceedsForeign

			// Realized Gain logic
			pos, exists := rs.Holdings[symbol]
			if exists {
				// Gain in Default Currency
				// (Sell Price - Avg Cost) * Qty
				gain := (priceInDefault - pos.AverageCost) * qty
				rs.RealizedGains[defaultCurrency] += gain

				// Reduce Quantity
				pos.Quantity -= qty
				if pos.Quantity < 0 {
					pos.Quantity = 0
				}
			}
		}

	case "DIV":
		if txn.Amount != nil {
			amount := *txn.Amount
			rs.Cash[txn.PriceCurrency] += amount
		}
	}

	return nil
}

// ToSnapshot converts the current state into a PortfolioSnapshot.
func (rs *ReplayState) ToSnapshot(userID string, timestamp time.Time) *domain.PortfolioSnapshot {
	snap := &domain.PortfolioSnapshot{
		UserID:    userID,
		Timestamp: timestamp,
		State: domain.SnapshotState{
			Holdings:      make(map[string]domain.HoldingState),
			Cash:          make(map[string]string),
			RealizedGains: make(map[string]string),
			NetInvested:   fmt.Sprintf("%.4f", rs.NetInvested),
		},
	}

	for sym, pos := range rs.Holdings {
		snap.State.Holdings[sym] = domain.HoldingState{
			Quantity:  fmt.Sprintf("%.6f", pos.Quantity),
			CostBasis: fmt.Sprintf("%.4f", pos.AverageCost*pos.Quantity), // Store Total Cost Basis
			Currency:  pos.Currency,
		}
	}

	for curr, amt := range rs.Cash {
		snap.State.Cash[curr] = fmt.Sprintf("%.4f", amt)
	}

	for curr, amt := range rs.RealizedGains {
		snap.State.RealizedGains[curr] = fmt.Sprintf("%.4f", amt)
	}

	return snap
}
