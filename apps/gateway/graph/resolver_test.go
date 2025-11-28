package graph

import (
	"context"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock UserServiceClient
type MockUserServiceClient struct {
	userpb.UserServiceClient
	GetUserFunc    func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error)
	CreateUserFunc func(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, in, opts...)
	}
	return nil, nil
}

// Mock PortfolioServiceClient
type MockPortfolioServiceClient struct {
	portfoliopb.PortfolioServiceClient
	GetPortfolioSummaryFunc     func(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioSummaryResponse, error)
	GetHoldingsFunc             func(ctx context.Context, in *portfoliopb.GetHoldingsRequest, opts ...grpc.CallOption) (*portfoliopb.GetHoldingsResponse, error)
	GetPortfolioPerformanceFunc func(ctx context.Context, in *portfoliopb.GetPortfolioPerformanceRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioPerformanceResponse, error)
}

func (m *MockPortfolioServiceClient) GetPortfolioSummary(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioSummaryResponse, error) {
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

func TestQueryResolver_User(t *testing.T) {
	mockUserClient := &MockUserServiceClient{
		GetUserFunc: func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
			return &userpb.GetUserResponse{
				Id:    "user-123",
				Name:  "John Doe",
				Email: "john@example.com",
			}, nil
		},
	}

	resolver := &Resolver{
		UserClient: mockUserClient,
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
		CreateUserFunc: func(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
			return &userpb.CreateUserResponse{
				Id: "new-user-456",
			}, nil
		},
	}

	resolver := &Resolver{
		UserClient: mockUserClient,
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
		GetPortfolioSummaryFunc: func(ctx context.Context, in *portfoliopb.GetPortfolioSummaryRequest, opts ...grpc.CallOption) (*portfoliopb.GetPortfolioSummaryResponse, error) {
			return &portfoliopb.GetPortfolioSummaryResponse{
				Summary: &portfoliopb.PortfolioSummary{
					TotalValue:              10000.50,
					TotalGainLoss:           500.25,
					TotalGainLossPercentage: 5.25,
					Currency:                "USD",
					LastUpdated:             timestamppb.New(now),
				},
			}, nil
		},
	}

	resolver := &Resolver{
		PortfolioClient: mockPortfolioClient,
	}

	portfolioResolver := &portfolioResolver{resolver}

	portfolio := &model.Portfolio{
		ID:     "portfolio-1",
		UserID: "user-123",
	}

	summary, err := portfolioResolver.Summary(context.Background(), portfolio)
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

	resolver := &Resolver{
		PortfolioClient: mockPortfolioClient,
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

	resolver := &Resolver{
		PortfolioClient: mockPortfolioClient,
	}

	queryResolver := &queryResolver{resolver}

	performance, err := queryResolver.PortfolioPerformance(context.Background(), "user-123", "1w")
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
	resolver := &Resolver{}
	queryResolver := &queryResolver{resolver}

	portfolio, err := queryResolver.Portfolio(context.Background(), "user-123")
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
