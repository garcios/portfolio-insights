package usecase

import (
	"fmt"
	"log"
)

// ValuationHolding represents the minimal data required to value a holding.
type ValuationHolding struct {
	Ticker        string
	Quantity      float64
	AssetCurrency string
}

// ValuationResult contains the total value and any warnings encountered.
type ValuationResult struct {
	TotalValue float64
	Warnings   []string
}

// CalculatePortfolioValue calculates the Total Current Value of a portfolio in the target currency.
//
// Precision Note:
// This implementation uses float64 for calculations. While suitable for general estimations and
// many financial applications, binary floating-point arithmetic can introduce small precision errors
// (e.g., 0.1 + 0.2 != 0.3).
// For strict accounting or ledger systems where exact decimal precision is critical, considering
// a library like 'github.com/shopspring/decimal' or using integer based calculations (cents) is recommended.
// Here, we maintain float64 for simplicity and compatibility with existing domain models, but explicitly
// round or truncating for display should be handled at the presentation layer.
func CalculatePortfolioValue(
	holdings []ValuationHolding,
	marketQuotes map[string]float64,
	fxRates map[string]float64,
	targetCurrency string,
) (*ValuationResult, error) {

	totalValue := 0.0
	var warnings []string

	for _, h := range holdings {
		log.Printf("Processing holding: %s, Quantity: %f, AssetCurrency: %s", h.Ticker, h.Quantity, h.AssetCurrency)
		// 1. Fetch Market Price
		price, ok := marketQuotes[h.Ticker]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("Missing price for ticker: %s", h.Ticker))
			// Do not silently assume 0. Skip adding value, but report warning.
			continue
		}

		// 2. Calculate Market Value in Native Currency
		nativeValue := price * h.Quantity

		// 3. Determine FX Rate
		rate := 1.0
		if h.AssetCurrency != targetCurrency {
			// Construct key e.g. "USD/AUD"
			fxKey := fmt.Sprintf("%s/%s", h.AssetCurrency, targetCurrency)
			fxRate, found := fxRates[fxKey]
			if !found {
				// Try reverse? The requirement implies specific Key format.
				// We stick to requirement.
				log.Printf("Missing FX rate for: %s", fxKey)
				warnings = append(warnings, fmt.Sprintf("Missing FX rate for: %s", fxKey))
				continue
			}
			rate = fxRate
		}

		// 4. Convert to Target Currency
		convertedValue := nativeValue * rate

		// 5. Sum
		totalValue += convertedValue
	}

	// We return a result even with warnings.
	// We only return error if the input data is structurally invalid if needed, but here warnings cover missing data.
	return &ValuationResult{
		TotalValue: totalValue, // In Target Currency
		Warnings:   warnings,
	}, nil
}
