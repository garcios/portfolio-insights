package graph

import (
	"context"
	"testing"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/container"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/dataloader"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/middleware"
	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPortfolioResolver_Allocations(t *testing.T) {
	// Mock User Service
	mockUserClient := &MockUserServiceClient{
		GetUserFunc: func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.User, error) {
			prefs, _ := structpb.NewStruct(map[string]interface{}{
				"default_currency": "EUR",
			})
			return &userpb.User{
				UserId:      "user-123",
				Preferences: prefs,
			}, nil
		},
	}

	// Mock MarketData Service
	mockMarketDataClient := &MockMarketDataServiceClient{}
	mockMarketDataClient.GetLatestCurrencyRateFunc = func(ctx context.Context, in *marketdatapb.GetLatestCurrencyRateRequest, opts ...grpc.CallOption) (*marketdatapb.CurrencyRate, error) {
		// USD -> EUR = 0.85
		if in.BaseCurrency == "USD" && in.TargetCurrency == "EUR" {
			return &marketdatapb.CurrencyRate{Rate: 0.85}, nil
		}
		// GBP -> EUR = 1.15
		if in.BaseCurrency == "GBP" && in.TargetCurrency == "EUR" {
			return &marketdatapb.CurrencyRate{Rate: 1.15}, nil
		}
		return &marketdatapb.CurrencyRate{Rate: 1.0}, nil
	}

	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{Container: c}
	portfolioResolver := &portfolioResolver{resolver}

	// Setup Context with Loaders
	loaders := &middleware.Loaders{
		UserLoader:         dataloader.NewUserLoader(c.UserGateway),
		ExchangeRateLoader: dataloader.NewExchangeRateLoader(c.MarketDataGateway),
	}
	ctx := middleware.WithLoaders(context.Background(), loaders)

	t.Run("Normal Case", func(t *testing.T) {
		portfolio := &model.Portfolio{
			ID:     "p1",
			UserID: "user-123",
			Holdings: []*entity.Holding{
				{Symbol: "AAPL", Currency: "USD", CurrentValue: 1000, AssetName: "Apple", UserID: "user-123"}, // 850 EUR
				{Symbol: "BP", Currency: "GBP", CurrentValue: 1000, AssetName: "BP", UserID: "user-123"},      // 1150 EUR
			},
		}

		allocations, err := portfolioResolver.Allocations(ctx, portfolio)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allocations) != 2 {
			t.Fatalf("expected 2 allocations, got %d", len(allocations))
		}

		totalVal := 850.0 + 1150.0 // 2000.0

		// Check AAPL
		aapl := allocations[0]
		if aapl.Symbol != "AAPL" {
			t.Errorf("expected AAPL first")
		}
		expectedPct1 := (850.0 / totalVal) * 100
		if aapl.Percentage != expectedPct1 {
			t.Errorf("expected AAPL pct %f, got %f", expectedPct1, aapl.Percentage)
		}
		if *aapl.MarketValueInTargetCurrency != 850.0 {
			t.Errorf("expected AAPL val 850.0, got %f", *aapl.MarketValueInTargetCurrency)
		}

		// Check BP
		bp := allocations[1]
		expectedPct2 := (1150.0 / totalVal) * 100
		if bp.Percentage != expectedPct2 {
			t.Errorf("expected BP pct %f, got %f", expectedPct2, bp.Percentage)
		}
	})

	t.Run("Zero Total Value", func(t *testing.T) {
		portfolio := &model.Portfolio{
			ID:     "p2",
			UserID: "user-123",
			Holdings: []*entity.Holding{
				{Symbol: "JUNK", Currency: "USD", CurrentValue: 0, AssetName: "Junk", UserID: "user-123"},
			},
		}

		allocations, err := portfolioResolver.Allocations(ctx, portfolio)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if allocations[0].Percentage != 0 {
			t.Errorf("expected 0 percentage for zero value portfolio")
		}
	})

	t.Run("Fetches Holdings", func(t *testing.T) {
		// Setup mock to return holdings
		mockPortfolioClient.GetHoldingsFunc = func(ctx context.Context, in *portfoliopb.GetHoldingsRequest, opts ...grpc.CallOption) (*portfoliopb.GetHoldingsResponse, error) {
			return &portfoliopb.GetHoldingsResponse{
				Holdings: []*portfoliopb.Holding{
					{Symbol: "AAPL", Currency: "USD", CurrentValue: 1000.0, AssetName: "Apple"},
				},
			}, nil
		}

		portfolio := &model.Portfolio{
			ID:     "p3",
			UserID: "user-123",
			// Holdings is empty/nil
		}

		allocations, err := portfolioResolver.Allocations(ctx, portfolio)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allocations) != 1 {
			t.Fatalf("expected 1 allocation, got %d", len(allocations))
		}
		if allocations[0].Symbol != "AAPL" {
			t.Errorf("expected AAPL, got %s", allocations[0].Symbol)
		}
	})
}
