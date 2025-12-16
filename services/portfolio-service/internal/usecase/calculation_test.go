package usecase

import (
	"math"
	"testing"
)

func TestCalculatePortfolioValue(t *testing.T) {

	// Mock Data
	marketQuotes := map[string]float64{
		"AAPL": 150.0,
		"CBA":  100.0,
		"TSLA": 900.0,
	}

	fxRates := map[string]float64{
		"USD/AUD": 1.50,
		"AUD/AUD": 1.0,
		"JPY/AUD": 0.01,
	}

	tests := []struct {
		name           string
		holdings       []ValuationHolding
		expectedValue  float64
		expectedWarns  int
		targetCurrency string
	}{
		{
			name: "Success - Mixed Currencies (User Test Case)",
			holdings: []ValuationHolding{
				{Ticker: "AAPL", Quantity: 10, AssetCurrency: "USD"},
				{Ticker: "CBA", Quantity: 5, AssetCurrency: "AUD"},
			},
			// AAPL: 10 * 150 * 1.5 = 2250
			// CBA:  5 * 100 * 1.0 = 500
			// Total: 2750
			expectedValue:  2750.0,
			expectedWarns:  0,
			targetCurrency: "AUD",
		},
		{
			name: "Missing Price",
			holdings: []ValuationHolding{
				{Ticker: "UNKNOWN", Quantity: 10, AssetCurrency: "USD"},
				{Ticker: "CBA", Quantity: 5, AssetCurrency: "AUD"},
			},
			// UNKNOWN: Skipped (Warning)
			// CBA: 500
			expectedValue:  500.0,
			expectedWarns:  1,
			targetCurrency: "AUD",
		},
		{
			name: "Missing FX Rate",
			holdings: []ValuationHolding{
				{Ticker: "TSLA", Quantity: 1, AssetCurrency: "MARS_COIN"},
			},
			// TSLA Price 900 exists.
			// FX "MARS_COIN/AUD" missing. Skipped.
			expectedValue:  0.0,
			expectedWarns:  1,
			targetCurrency: "AUD",
		},
		{
			name: "Same Currency Optimization",
			holdings: []ValuationHolding{
				{Ticker: "CBA", Quantity: 10, AssetCurrency: "AUD"},
			},
			// FX Rate lookup skipped effectively if logic handles it?
			// Implementation checks: if h.AssetCurrency != targetCurrency.
			// Here AUD == AUD. Rate = 1.0. logic applies `rate = 1.0`.
			// Value: 10 * 100 * 1.0 = 1000.
			expectedValue:  1000.0,
			expectedWarns:  0,
			targetCurrency: "AUD",
		},
		{
			name: "Precision Check (Small floats)",
			holdings: []ValuationHolding{
				// 0.1 quantity * 0.2 price = 0.02
				// FX 1.0
				{Ticker: "CBA", Quantity: 0.1, AssetCurrency: "AUD"},
			},
			// CBA Price 100.
			// Value: 0.1 * 100 = 10.
			expectedValue:  10.0,
			expectedWarns:  0,
			targetCurrency: "AUD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculatePortfolioValue(tt.holdings, marketQuotes, fxRates, tt.targetCurrency)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result.Warnings) != tt.expectedWarns {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectedWarns, len(result.Warnings), result.Warnings)
			}

			if math.Abs(result.TotalValue-tt.expectedValue) > 0.0001 {
				t.Errorf("Expected Value %f, got %f", tt.expectedValue, result.TotalValue)
			}
		})
	}
}
