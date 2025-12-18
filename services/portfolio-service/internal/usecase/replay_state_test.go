package usecase

import (
	"math"
	"testing"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
)

func TestReplayState_Apply_Deposit_Withdraw(t *testing.T) {
	rs := NewReplayState()
	amount := 1000.0

	// Test Deposit
	err := rs.Apply(&transactionpb.Transaction{
		Type:          "DEP",
		Amount:        &amount,
		PriceCurrency: "AUD",
	}, 1.0, "AUD")

	if err != nil {
		t.Fatalf("Apply(DEP) failed: %v", err)
	}

	if rs.Cash["AUD"] != 1000.0 {
		t.Errorf("Expected Cash 1000, got %f", rs.Cash["AUD"])
	}

	// Test Withdrawal
	wAmt := 500.0
	err = rs.Apply(&transactionpb.Transaction{
		Type:          "WIT",
		Amount:        &wAmt,
		PriceCurrency: "AUD",
	}, 1.0, "AUD")

	if err != nil {
		t.Fatalf("Apply(WIT) failed: %v", err)
	}

	if rs.Cash["AUD"] != 500.0 {
		t.Errorf("Expected Cash 500, got %f", rs.Cash["AUD"])
	}
}

func TestReplayState_Apply_Buy_Sell(t *testing.T) {
	rs := NewReplayState()
	rs.Cash["AUD"] = 20000.0

	symbol := "AAPL"
	qty := 10.0
	price := 150.0 // USD
	rate := 1.5    // AUD per USD

	// Test BUY
	// Cost in USD: 1500
	// Cost in AUD: 2250 (1500 * 1.5)
	err := rs.Apply(&transactionpb.Transaction{
		Type:          "BUY",
		Symbol:        &symbol,
		Quantity:      &qty,
		PricePerShare: &price,
		PriceCurrency: "USD",
	}, rate, "AUD")

	if err != nil {
		t.Fatalf("Apply(BUY) failed: %v", err)
	}

	// Check Holding
	pos, exists := rs.Holdings["AAPL"]
	if !exists {
		t.Fatal("Holding AAPL not created")
	}
	if pos.Quantity != 10.0 {
		t.Errorf("Expected Qty 10, got %f", pos.Quantity)
	}
	// Avg Cost (AUD): 2250 / 10 = 225
	if pos.AverageCost != 225.0 {
		t.Errorf("Expected AvgCost 225.0, got %f", pos.AverageCost)
	}
	// Avg Foreign Cost (USD): 150.0
	if pos.AverageForeignCost != 150.0 {
		t.Errorf("Expected AvgForeignCost 150.0, got %f", pos.AverageForeignCost)
	}
	// Cash Check: Should reduce by USD amount? Apply logic said:
	// "rs.Cash[txn.PriceCurrency] -= totalCostForeign"
	// So 20000 AUD cash remains (assuming account handles FX elsewhere or we seeded USD).
	// If we only have AUD, this test setup is slightly weird if we assume auto-conversion.
	// But `ReplayState` manages multi-currency cash buckets.
	// So USD cash should be -1500.
	if rs.Cash["USD"] != -1500.0 {
		t.Errorf("Expected USD Cash -1500, got %f", rs.Cash["USD"])
	}

	// Test SELL (Partial)
	// Sell 5 units @ 200 USD. Rate 1.6
	sellQty := 5.0
	sellPrice := 200.0
	sellRate := 1.6

	err = rs.Apply(&transactionpb.Transaction{
		Type:          "SELL",
		Symbol:        &symbol,
		Quantity:      &sellQty,
		PricePerShare: &sellPrice,
		PriceCurrency: "USD",
	}, sellRate, "AUD")

	if err != nil {
		t.Fatalf("Apply(SELL) failed: %v", err)
	}

	// Check Holding
	if pos.Quantity != 5.0 {
		t.Errorf("Expected Qty 5, got %f", pos.Quantity)
	}
	// Avg Cost should remain same
	if pos.AverageCost != 225.0 {
		t.Errorf("Expected AvgCost 225.0, got %f", pos.AverageCost)
	}

	// Check Realized Gain (AUD)
	// Sell Proceeds AUD: 5 * 200 * 1.6 = 1600
	// Cost Basis AUD: 5 * 225 = 1125
	// Gain: 1600 - 1125 = 475
	if math.Abs(rs.RealizedGains["AUD"]-475.0) > 0.000001 {
		t.Errorf("Expected RealizedGain 475.0, got %f", rs.RealizedGains["AUD"])
	}

	// Check Cash USD
	// +1000 USD (5*200)
	// Prev: -1500. New: -500.
	if rs.Cash["USD"] != -500.0 {
		t.Errorf("Expected USD Cash -500, got %f", rs.Cash["USD"])
	}
}

func TestReplayState_HydrateFrom(t *testing.T) {
	rs := NewReplayState()

	snap := &domain.PortfolioSnapshot{
		State: domain.SnapshotState{
			Holdings: map[string]domain.HoldingState{
				"AAPL": {
					Quantity:  "10.0",
					CostBasis: "2250.0", // Total Cost in Default Currency? Or Per Unit?
					// In Apply we store AverageCost.
					// Hydrate logic: "AverageCost = cost / qty"
					// So Snapshot CostBasis field stores TOTAL cost.
					Currency: "USD",
				},
			},
			Cash: map[string]string{
				"AUD": "1000.0",
			},
			RealizedGains: map[string]string{
				"AUD": "500.0",
			},
		},
	}

	err := rs.HydrateFrom(snap)
	if err != nil {
		t.Fatalf("HydrateFrom failed: %v", err)
	}

	// Verify
	pos, ok := rs.Holdings["AAPL"]
	if !ok {
		t.Fatal("Holdings[AAPL] missing")
	}
	if pos.Quantity != 10.0 {
		t.Errorf("Qty mismatch: %f", pos.Quantity)
	}
	// AvgCost = 2250 / 10 = 225
	if pos.AverageCost != 225.0 {
		t.Errorf("AvgCost mismatch: %f", pos.AverageCost)
	}

	if rs.Cash["AUD"] != 1000.0 {
		t.Errorf("Cash mismatch: %f", rs.Cash["AUD"])
	}

	if rs.RealizedGains["AUD"] != 500.0 {
		t.Errorf("Gains mismatch: %f", rs.RealizedGains["AUD"])
	}
}
