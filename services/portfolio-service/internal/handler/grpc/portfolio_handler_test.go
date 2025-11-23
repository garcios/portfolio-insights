package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock PortfolioUsecase
type mockPortfolioUsecase struct {
	holdings []*domain.Holding
	summary  *domain.PortfolioSummary
	err      error
}

func (m *mockPortfolioUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.holdings, nil
}

func (m *mockPortfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.summary, nil
}

func TestGetHoldings_Success(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{
		holdings: []*domain.Holding{
			{
				UserID:       "user-123",
				Symbol:       "AAPL",
				Quantity:     10,
				AverageCost:  150.00,
				CurrentPrice: 175.50,
			},
			{
				UserID:       "user-123",
				Symbol:       "GOOGL",
				Quantity:     5,
				AverageCost:  2800.00,
				CurrentPrice: 2950.00,
			},
		},
	}

	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetHoldingsRequest{
		UserId: "user-123",
	}
	resp, err := handler.GetHoldings(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Holdings) != 2 {
		t.Fatalf("Expected 2 holdings, got %d", len(resp.Holdings))
	}

	// Verify AAPL holding
	aaplHolding := resp.Holdings[0]
	if aaplHolding.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", aaplHolding.Symbol)
	}

	if aaplHolding.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %f", aaplHolding.Quantity)
	}

	if aaplHolding.AveragePrice != 150.00 {
		t.Errorf("Expected average price 150.00, got %f", aaplHolding.AveragePrice)
	}

	if aaplHolding.CurrentPrice != 175.50 {
		t.Errorf("Expected current price 175.50, got %f", aaplHolding.CurrentPrice)
	}

	// Verify calculated fields
	expectedCurrentValue := 10 * 175.50 // 1755.00
	if aaplHolding.CurrentValue != expectedCurrentValue {
		t.Errorf("Expected current value %f, got %f", expectedCurrentValue, aaplHolding.CurrentValue)
	}

	expectedGainLoss := (10 * 175.50) - (10 * 150.00) // 255.00
	if aaplHolding.GainLoss != expectedGainLoss {
		t.Errorf("Expected gain/loss %f, got %f", expectedGainLoss, aaplHolding.GainLoss)
	}

	expectedGainLossPct := (255.00 / 1500.00) * 100 // 17.0
	if aaplHolding.GainLossPercentage != expectedGainLossPct {
		t.Errorf("Expected gain/loss pct %f, got %f", expectedGainLossPct, aaplHolding.GainLossPercentage)
	}
}

func TestGetHoldings_EmptyUserId(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetHoldingsRequest{
		UserId: "",
	}
	_, err := handler.GetHoldings(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty user_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}

	if st.Message() != "user_id is required" {
		t.Errorf("Expected message 'user_id is required', got '%s'", st.Message())
	}
}

func TestGetHoldings_UsecaseError(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{
		err: errors.New("database connection failed"),
	}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetHoldingsRequest{
		UserId: "user-123",
	}
	_, err := handler.GetHoldings(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("Expected code Internal, got %v", st.Code())
	}
}

func TestGetHoldings_EmptyHoldings(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{
		holdings: []*domain.Holding{},
	}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetHoldingsRequest{
		UserId: "user-456",
	}
	resp, err := handler.GetHoldings(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Holdings) != 0 {
		t.Errorf("Expected 0 holdings, got %d", len(resp.Holdings))
	}
}

func TestGetPortfolioSummary_Success(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{
		summary: &domain.PortfolioSummary{
			UserID:      "user-123",
			TotalValue:  16505.00,
			TotalCost:   15500.00,
			GainLoss:    1005.00,
			GainLossPct: 6.48,
		},
	}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetPortfolioSummaryRequest{
		UserId: "user-123",
	}
	resp, err := handler.GetPortfolioSummary(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Summary == nil {
		t.Fatal("Expected summary, got nil")
	}

	if resp.Summary.UserId != "user-123" {
		t.Errorf("Expected user_id 'user-123', got '%s'", resp.Summary.UserId)
	}

	if resp.Summary.TotalValue != 16505.00 {
		t.Errorf("Expected total value 16505.00, got %f", resp.Summary.TotalValue)
	}

	if resp.Summary.TotalGainLoss != 1005.00 {
		t.Errorf("Expected gain/loss 1005.00, got %f", resp.Summary.TotalGainLoss)
	}

	if resp.Summary.TotalGainLossPercentage != 6.48 {
		t.Errorf("Expected gain/loss pct 6.48, got %f", resp.Summary.TotalGainLossPercentage)
	}

	if resp.Summary.LastUpdated == nil {
		t.Error("Expected last_updated timestamp, got nil")
	}
}

func TestGetPortfolioSummary_EmptyUserId(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetPortfolioSummaryRequest{
		UserId: "",
	}
	_, err := handler.GetPortfolioSummary(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty user_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestGetPortfolioSummary_UsecaseError(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{
		err: errors.New("failed to calculate summary"),
	}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetPortfolioSummaryRequest{
		UserId: "user-123",
	}
	_, err := handler.GetPortfolioSummary(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("Expected code Internal, got %v", st.Code())
	}
}

func TestGetPortfolioPerformance_EmptyUserId(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetPortfolioPerformanceRequest{
		UserId: "",
		Period: "1m",
	}
	_, err := handler.GetPortfolioPerformance(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty user_id, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestGetPortfolioPerformance_Stub(t *testing.T) {
	// Setup
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetPortfolioPerformanceRequest{
		UserId: "user-123",
		Period: "1m",
	}
	resp, err := handler.GetPortfolioPerformance(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Currently returns empty data points (stub implementation)
	if len(resp.DataPoints) != 0 {
		t.Errorf("Expected 0 data points (stub), got %d", len(resp.DataPoints))
	}
}

func TestGetHoldings_CalculationsWithZeroCost(t *testing.T) {
	// Edge case: holding with zero cost basis
	mockUC := &mockPortfolioUsecase{
		holdings: []*domain.Holding{
			{
				UserID:       "user-123",
				Symbol:       "FREE",
				Quantity:     10,
				AverageCost:  0,
				CurrentPrice: 100.00,
			},
		},
	}
	handler := NewPortfolioHandler(mockUC)

	// Execute
	req := &pb.GetHoldingsRequest{
		UserId: "user-123",
	}
	resp, err := handler.GetHoldings(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	holding := resp.Holdings[0]

	// Should not panic with division by zero
	if holding.GainLossPercentage != 0 {
		t.Errorf("Expected gain/loss pct 0 (avoid division by zero), got %f", holding.GainLossPercentage)
	}
}
