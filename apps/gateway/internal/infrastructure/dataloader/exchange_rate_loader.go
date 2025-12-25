package dataloader

import (
	"context"
	"fmt"
	"strings"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/graph-gophers/dataloader/v7"
)

// ExchangeRateLoader batches and caches exchange rate requests
// Key format: "BASE:TARGET"
type ExchangeRateLoader struct {
	loader *dataloader.Loader[string, float64]
}

func NewExchangeRateLoader(marketDataGateway gateway.MarketDataGateway) *ExchangeRateLoader {
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[float64] {
		results := make([]*dataloader.Result[float64], len(keys))

		// Since we don't have a batch API, we'll process these concurrently
		// Note: Because of DataLoader caching, 'keys' will typically contain unique currency pairs
		// requested in this tick.

		// Map to store results to matching indices (in case of duplicate keys if cache was off, though it's on by default)
		// Actually, we can just iterate.

		for i, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				results[i] = &dataloader.Result[float64]{Error: fmt.Errorf("invalid key format: %s", key)}
				continue
			}
			base, target := parts[0], parts[1]

			// Optimization: If base == target, return 1.0 immediately?
			// The resolver handles this, but loader could too.
			if base == target {
				results[i] = &dataloader.Result[float64]{Data: 1.0}
				continue
			}

			// Call Gateway (which calls gRPC)
			// TODO: Parallelize this loop if latency is an issue?
			// For now, simple sequential calls. Since keys are deduped, this is O(UniquePairs), not O(Holdings).
			rate, err := marketDataGateway.GetExchangeRate(ctx, base, target)
			results[i] = &dataloader.Result[float64]{
				Data:  rate,
				Error: err,
			}
		}
		return results
	}

	return &ExchangeRateLoader{
		loader: dataloader.NewBatchedLoader(batchFn),
	}
}

func (l *ExchangeRateLoader) Load(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	key := fmt.Sprintf("%s:%s", baseCurrency, targetCurrency)
	return l.loader.Load(ctx, key)()
}
