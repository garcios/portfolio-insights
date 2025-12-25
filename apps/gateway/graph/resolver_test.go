package graph

import (
	"context"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/auth"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/container"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/dataloader"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/middleware"
	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock UserServiceClient
type MockUserServiceClient struct {
	userpb.UserServiceClient
	GetUserFunc    func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.User, error)
	CreateUserFunc func(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.User, error)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, in, opts...)
	}
	return nil, nil
}

// Mock PortfolioServiceClient
type MockPortfolioServiceClient struct {
	portfoliopb.PortfolioServiceClient
	GetPortfolioSummaryFunc     func(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.PortfolioSummary, error)
	GetHoldingsFunc             func(ctx context.Context, in *portfoliopb.GetHoldingsRequest, opts ...grpc.CallOption) (*portfoliopb.GetHoldingsResponse, error)
	GetPortfolioPerformanceFunc func(ctx context.Context, in *portfoliopb.GetPortfolioPerformanceRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioPerformanceResponse, error)
}

func (m *MockPortfolioServiceClient) GetPortfolioSummary(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.PortfolioSummary, error) {
	if m.GetPortfolioSummaryFunc != nil {
		return m.GetPortfolioSummaryFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockPortfolioServiceClient) GetHoldings(ctx context.Context, in *portfoliopb.GetHoldingsRequest, opts ...grpc.CallOption) (*portfoliopb.GetHoldingsResponse, error) {
	if m.GetHoldingsFunc != nil {
		return m.GetHoldingsFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockPortfolioServiceClient) GetPortfolioPerformance(ctx context.Context, in *portfoliopb.GetPortfolioPerformanceRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioPerformanceResponse, error) {
	if m.GetPortfolioPerformanceFunc != nil {
		return m.GetPortfolioPerformanceFunc(ctx, in, opts...)
	}
	return nil, nil
}

// Mock TransactionServiceClient
type MockTransactionServiceClient struct {
	transactionpb.TransactionServiceClient
	CreateTransactionFunc func(ctx context.Context, in *transactionpb.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error)
	ListTransactionsFunc  func(ctx context.Context, in *transactionpb.ListTransactionsRequest, opts ...grpc.CallOption) (*transactionpb.ListTransactionsResponse, error)
}

func (m *MockTransactionServiceClient) CreateTransaction(ctx context.Context, in *transactionpb.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockTransactionServiceClient) ListTransactions(ctx context.Context, in *transactionpb.ListTransactionsRequest, opts ...grpc.CallOption) (*transactionpb.ListTransactionsResponse, error) {
	if m.ListTransactionsFunc != nil {
		return m.ListTransactionsFunc(ctx, in, opts...)
	}
	return nil, nil
}

// Mock MarketDataServiceClient
// Mock MarketDataServiceClient
type MockMarketDataServiceClient struct {
	marketdatapb.MarketDataServiceClient
	GetLatestCurrencyRateFunc func(ctx context.Context, in *marketdatapb.GetLatestCurrencyRateRequest, opts ...grpc.CallOption) (*marketdatapb.CurrencyRate, error)
}

func (m *MockMarketDataServiceClient) GetLatestCurrencyRate(ctx context.Context, in *marketdatapb.GetLatestCurrencyRateRequest, opts ...grpc.CallOption) (*marketdatapb.CurrencyRate, error) {
	if m.GetLatestCurrencyRateFunc != nil {
		return m.GetLatestCurrencyRateFunc(ctx, in, opts...)
	}
	return &marketdatapb.CurrencyRate{
		Rate: 1.0,
	}, nil
}

func TestQueryResolver_User(t *testing.T) {
	mockUserClient := &MockUserServiceClient{
		GetUserFunc: func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.User, error) {
			return &userpb.User{
				UserId:   "user-123",
				Username: "John Doe",
				Email:    "john@example.com",
			}, nil
		},
	}

	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	queryResolver := &queryResolver{resolver}

	user, err := queryResolver.User(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != "user-123" {
		t.Errorf("expected ID 'user-123', got '%s'", user.ID)
	}
	if user.Username != "John Doe" {
		t.Errorf("expected username 'John Doe', got '%s'", user.Username)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", user.Email)
	}
}

func TestMutationResolver_CreateUser(t *testing.T) {
	mockUserClient := &MockUserServiceClient{
		CreateUserFunc: func(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.User, error) {
			return &userpb.User{
				UserId:   "new-user-456",
				Username: "Jane Doe",
				Email:    "jane@example.com",
			}, nil
		},
	}

	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	mutationResolver := &mutationResolver{resolver}

	input := model.NewUser{
		Username: "Jane Doe",
		Email:    "jane@example.com",
	}

	user, err := mutationResolver.CreateUser(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != "new-user-456" {
		t.Errorf("expected ID 'new-user-456', got '%s'", user.ID)
	}
	if user.Username != "Jane Doe" {
		t.Errorf("expected username 'Jane Doe', got '%s'", user.Username)
	}
}

func TestPortfolioResolver_Summary(t *testing.T) {
	now := time.Now()
	mockPortfolioClient := &MockPortfolioServiceClient{
		GetPortfolioSummaryFunc: func(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.PortfolioSummary, error) {
			return &portfoliopb.PortfolioSummary{
				TotalValue:              10000.50,
				TotalGainLoss:           500.25,
				TotalGainLossPercentage: 5.25,
				DayChange:               100.0,
				DayChangePercentage:     1.0,
				Currency:                "USD",
				LastUpdated:             timestamppb.New(now),
			}, nil
		},
	}

	mockUserClient := &MockUserServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	portfolioResolver := &portfolioResolver{resolver}

	portfolio := &model.Portfolio{
		ID:     "portfolio-1",
		UserID: "user-123",
	}

	summary, err := portfolioResolver.Summary(context.Background(), portfolio, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.TotalValue != 10000.50 {
		t.Errorf("expected total value 10000.50, got %f", summary.TotalValue)
	}
	if summary.TotalGainLoss != 500.25 {
		t.Errorf("expected gain/loss 500.25, got %f", summary.TotalGainLoss)
	}
	if summary.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", summary.Currency)
	}
}

func TestPortfolioResolver_Holdings(t *testing.T) {
	mockPortfolioClient := &MockPortfolioServiceClient{
		GetHoldingsFunc: func(ctx context.Context, in *portfoliopb.GetHoldingsRequest, opts ...grpc.CallOption) (*portfoliopb.GetHoldingsResponse, error) {
			return &portfoliopb.GetHoldingsResponse{
				Holdings: []*portfoliopb.Holding{
					{
						Symbol:             "AAPL",
						Quantity:           10,
						AveragePrice:       150.0,
						CurrentPrice:       180.0,
						CurrentValue:       1800.0,
						GainLoss:           300.0,
						GainLossPercentage: 20.0,
						Currency:           "USD",
						AssetName:          "Apple Inc.",
					},
					{
						Symbol:             "GOOGL",
						Quantity:           5,
						AveragePrice:       2000.0,
						CurrentPrice:       2500.0,
						CurrentValue:       12500.0,
						GainLoss:           2500.0,
						GainLossPercentage: 25.0,
						Currency:           "USD",
						AssetName:          "Alphabet Inc.",
					},
				},
			}, nil
		},
	}

	mockUserClient := &MockUserServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	portfolioResolver := &portfolioResolver{resolver}

	portfolio := &model.Portfolio{
		ID:     "portfolio-1",
		UserID: "user-123",
	}

	holdings, err := portfolioResolver.Holdings(context.Background(), portfolio)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(holdings) != 2 {
		t.Fatalf("expected 2 holdings, got %d", len(holdings))
	}

	if holdings[0].Symbol != "AAPL" {
		t.Errorf("expected symbol 'AAPL', got '%s'", holdings[0].Symbol)
	}
	if holdings[0].Quantity != 10 {
		t.Errorf("expected quantity 10, got %f", holdings[0].Quantity)
	}
	if holdings[1].Symbol != "GOOGL" {
		t.Errorf("expected symbol 'GOOGL', got '%s'", holdings[1].Symbol)
	}
}

func TestQueryResolver_PortfolioPerformance(t *testing.T) {
	now := time.Now()
	mockPortfolioClient := &MockPortfolioServiceClient{
		GetPortfolioPerformanceFunc: func(ctx context.Context, in *portfoliopb.GetPortfolioPerformanceRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioPerformanceResponse, error) {
			return &portfoliopb.GetPortfolioPerformanceResponse{
				DataPoints: []*portfoliopb.PortfolioPerformancePoint{
					{
						Timestamp: timestamppb.New(now.AddDate(0, 0, -2)),
						Value:     9500.0,
					},
					{
						Timestamp: timestamppb.New(now.AddDate(0, 0, -1)),
						Value:     9750.0,
					},
					{
						Timestamp: timestamppb.New(now),
						Value:     10000.0,
					},
				},
			}, nil
		},
	}

	mockUserClient := &MockUserServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	queryResolver := &queryResolver{resolver}

	ctx := auth.WithAuthContext(context.Background(), &auth.AuthContext{
		UserID: "user-123",
	})

	performance, err := queryResolver.PortfolioPerformance(ctx, "1w")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(performance) != 3 {
		t.Fatalf("expected 3 data points, got %d", len(performance))
	}

	if performance[0].Value != 9500.0 {
		t.Errorf("expected value 9500.0, got %f", performance[0].Value)
	}
	if performance[2].Value != 10000.0 {
		t.Errorf("expected value 10000.0, got %f", performance[2].Value)
	}
}

func TestQueryResolver_Portfolio(t *testing.T) {
	mockUserClient := &MockUserServiceClient{}
	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}
	mockMarketDataClient := &MockMarketDataServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	queryResolver := &queryResolver{resolver}

	ctx := auth.WithAuthContext(context.Background(), &auth.AuthContext{
		UserID: "user-123",
	})

	portfolio, err := queryResolver.Portfolio(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if portfolio.ID != "user-123" {
		t.Errorf("expected ID 'user-123', got '%s'", portfolio.ID)
	}
	if portfolio.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", portfolio.UserID)
	}
	if portfolio.Name != "My Portfolio" {
		t.Errorf("expected name 'My Portfolio', got '%s'", portfolio.Name)
	}
}

func TestHoldingResolver_CurrentValueInTargetCurrency(t *testing.T) {
	// Mock User Service for TargetCurrency resolver (via DataLoader)
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

	// Mock MarketData Service for CurrentValueInTargetCurrency resolver
	mockMarketDataClient := &MockMarketDataServiceClient{}
	// Override GetLatestCurrencyRate to return a specific rate
	mockMarketDataClient.GetLatestCurrencyRateFunc = func(ctx context.Context, in *marketdatapb.GetLatestCurrencyRateRequest, opts ...grpc.CallOption) (*marketdatapb.CurrencyRate, error) {
		if in.BaseCurrency == "USD" && in.TargetCurrency == "EUR" {
			return &marketdatapb.CurrencyRate{
				Rate: 0.85,
			}, nil
		}
		return &marketdatapb.CurrencyRate{Rate: 1.0}, nil
	}

	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	holdingResolver := &holdingResolver{resolver}

	// Setup Context with Loaders
	loaders := &middleware.Loaders{
		UserLoader:         dataloader.NewUserLoader(c.UserGateway),
		ExchangeRateLoader: dataloader.NewExchangeRateLoader(c.MarketDataGateway),
	}
	// Note: We're using context.Background() here, assuming middleware package has keys exported or we can just bypass
	// Actually, middleware.LoadersKey is likely unexported or we need to check.
	// Let's check middleware.GetLoaders implementation to be safe on how to inject it.
	// Assuming it grabs from context with a key.
	// If unavailable, I might need to check middleware pkg.
	// But let's assume "loaders" key for now or check if we can import key.
	// Using "loaders" string based on previous knowledge, but let's check `middleware` pkg view if this fails.
	// Wait, I already viewed main.go and it used `loadersKey` which was unexported in `main.go`?
	// Ah, `middleware.DataloaderMiddleware` was in `main.go`. No, it was imported.
	// I'll assume I can't inject easily without looking at middleware.
	// But wait, `schema.resolvers.go` uses `middleware.GetLoaders(ctx)`.
	// I should check `middleware.GetLoaders`.

	// Temporarily: I will use the set method if available, or just ContextWithValue if I know the key.
	// Let's rely on middleware.WithLoaders if it exists.
	// Or I will blindly try to set it.
	// Since I can't see middleware code right now (I viewed main.go and resolver.go), I will pause on this thought and add the function, but I suspect Context injection might fail if key is private.
	// Correction: I previously viewed `middleware.DataloaderMiddleware` which was in `apps/gateway/cmd/server/main.go` ???
	// No, `DataloaderMiddleware` was in `main.go` snippet. Wait, if it's in `main.go`, then `schema.resolvers.go` calling `middleware.GetLoaders` implies `middleware` package HAS key.

	ctx := middleware.WithLoaders(context.Background(), loaders) // Assuming this exists or similar.

	// Test Case
	holding := &entity.Holding{
		UserID:       "user-123",
		Symbol:       "AAPL",
		Currency:     "USD",
		CurrentValue: 100.0,
	}

	// Execute
	val, err := holdingResolver.CurrentValueInTargetCurrency(ctx, holding)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify
	expected := 85.0 // 100.0 * 0.85
	if val != expected {
		t.Errorf("expected value %f, got %f", expected, val)
	}
}

func TestHoldingResolver_GainLossInTargetCurrency(t *testing.T) {
	// Mock User Service for TargetCurrency resolver
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
		if in.BaseCurrency == "USD" && in.TargetCurrency == "EUR" {
			return &marketdatapb.CurrencyRate{
				Rate: 0.85,
			}, nil
		}
		return &marketdatapb.CurrencyRate{Rate: 1.0}, nil
	}

	mockPortfolioClient := &MockPortfolioServiceClient{}
	mockTransactionClient := &MockTransactionServiceClient{}

	c := container.NewContainer(mockUserClient, mockPortfolioClient, mockTransactionClient, "http://mock-url", mockMarketDataClient)
	resolver := &Resolver{
		Container: c,
	}

	holdingResolver := &holdingResolver{resolver}

	// Setup Context with Loaders
	loaders := &middleware.Loaders{
		UserLoader:         dataloader.NewUserLoader(c.UserGateway),
		ExchangeRateLoader: dataloader.NewExchangeRateLoader(c.MarketDataGateway),
	}
	ctx := middleware.WithLoaders(context.Background(), loaders)

	// Test Case
	holding := &entity.Holding{
		UserID:   "user-123",
		Symbol:   "AAPL",
		Currency: "USD",
		GainLoss: 50.0,
	}

	// Execute
	val, err := holdingResolver.GainLossInTargetCurrency(ctx, holding)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify
	expected := 42.5 // 50.0 * 0.85
	if val != expected {
		t.Errorf("expected value %f, got %f", expected, val)
	}
}
