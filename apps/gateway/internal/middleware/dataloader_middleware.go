package middleware

import (
	"context"
	"net/http"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/dataloader"
)

type contextKey string

const loadersKey contextKey = "dataloaders"

// Loaders holds references to all data loaders
type Loaders struct {
	UserLoader         *dataloader.UserLoader
	ExchangeRateLoader *dataloader.ExchangeRateLoader
}

// DataloaderMiddleware injects data loaders into the request context
func DataloaderMiddleware(userGateway gateway.UserGateway, marketDataGateway gateway.MarketDataGateway) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := &Loaders{
				UserLoader:         dataloader.NewUserLoader(userGateway),
				ExchangeRateLoader: dataloader.NewExchangeRateLoader(marketDataGateway),
			}
			ctx := context.WithValue(r.Context(), loadersKey, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLoaders retrieves the Loaders from the context
func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// WithLoaders injects the Loaders into the context (useful for testing)
func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, loaders)
}
