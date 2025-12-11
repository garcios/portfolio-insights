package grpc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock PortfolioUsecase
type mockPortfolioUsecase struct {
	holdings       []*domain.Holding
	summary        *domain.PortfolioSummary
	err            error
	backfillResult usecase.BackfillResult
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

func (m *mockPortfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.summary, nil
}

func (m *mockPortfolioUsecase) BackfillPortfolioHistory(
	ctx context.Context,
	userIDs []string,
	startDate, endDate time.Time,
	dryRun bool,
) usecase.BackfillResult {
	return m.backfillResult
}

// Mock PortfolioHistoryRepository
type mockHistoryRepo struct {
	snapshots []*domain.PortfolioSnapshot
	err       error
}

func (m *mockHistoryRepo) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	return m.err
}

func (m *mockHistoryRepo) GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error) {
	return m.snapshots, m.err
}

func (m *mockHistoryRepo) GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*domain.PortfolioSnapshot, error) {
	return m.snapshots, m.err
}

func (m *mockHistoryRepo) SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error) {
	return false, m.err
}

func (m *mockHistoryRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	return []string{"user-123"}, m.err
}

func TestGetHoldings_Success(t *testing.T) {
	mockUC := &mockPortfolioUsecase{
		holdings: []*domain.Holding{
			{
				UserID:       "user-123",
				Symbol:       "AAPL",
				Quantity:     10,
				AverageCost:  150.00,
				CurrentPrice: 175.50,
			},
		},
	}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetHoldingsRequest{
		Parent: resourcenames.UserName("user-123"),
	}
	resp, err := handler.GetHoldings(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Holdings) != 1 {
		t.Fatalf("Expected 1 holding, got %d", len(resp.Holdings))
	}

	holding := resp.Holdings[0]
	if holding.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", holding.Symbol)
	}

	expectedName := resourcenames.HoldingName("user-123", "AAPL")
	if holding.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, holding.Name)
	}
}

func TestGetHoldings_InvalidParent(t *testing.T) {
	mockUC := &mockPortfolioUsecase{}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetHoldingsRequest{
		Parent: "invalid-parent",
	}
	_, err := handler.GetHoldings(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid parent, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestGetPortfolioSummary_Success(t *testing.T) {
	mockUC := &mockPortfolioUsecase{
		summary: &domain.PortfolioSummary{
			UserID:      "user-123",
			TotalValue:  16505.00,
			TotalCost:   15500.00,
			GainLoss:    1005.00,
			GainLossPct: 6.48,
		},
	}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetPortfolioSummaryRequest{
		Name: resourcenames.PortfolioName("user-123"),
	}
	resp, err := handler.GetPortfolioSummary(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.UserId != "user-123" {
		t.Errorf("Expected user_id 'user-123', got '%s'", resp.UserId)
	}

	expectedName := resourcenames.PortfolioName("user-123")
	if resp.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, resp.Name)
	}

	if resp.TotalValue != 16505.00 {
		t.Errorf("Expected total value 16505.00, got %f", resp.TotalValue)
	}
}

func TestGetPortfolioSummary_InvalidResourceName(t *testing.T) {
	mockUC := &mockPortfolioUsecase{}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetPortfolioSummaryRequest{
		Name: "invalid-name",
	}
	_, err := handler.GetPortfolioSummary(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid resource name, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestGetPortfolioPerformance_Success(t *testing.T) {
	now := time.Now()
	mockRepo := &mockHistoryRepo{
		snapshots: []*domain.PortfolioSnapshot{
			{
				UserID:         "user-123",
				TotalValue:     10000.0,
				TotalCostBasis: 9000.0,
				Timestamp:      now.Add(-24 * time.Hour),
			},
			{
				UserID:         "user-123",
				TotalValue:     10500.0,
				TotalCostBasis: 9000.0,
				Timestamp:      now,
			},
		},
	}
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetPortfolioPerformanceRequest{
		Name:   resourcenames.PortfolioName("user-123"),
		Period: "1m",
	}
	resp, err := handler.GetPortfolioPerformance(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.DataPoints) != 2 {
		t.Errorf("Expected 2 data points, got %d", len(resp.DataPoints))
	}

	if resp.DataPoints[0].Value != 10000.0 {
		t.Errorf("Expected first point value 10000.0, got %f", resp.DataPoints[0].Value)
	}
}

func TestGetPortfolioPerformance_InvalidResourceName(t *testing.T) {
	mockRepo := &mockHistoryRepo{}
	mockUC := &mockPortfolioUsecase{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.GetPortfolioPerformanceRequest{
		Name:   "invalid-name",
		Period: "1m",
	}
	_, err := handler.GetPortfolioPerformance(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid resource name, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestBackfillHistory_Success(t *testing.T) {
	if err := os.Setenv("ADMIN_TOKEN", "secret-token"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ADMIN_TOKEN"); err != nil {
			t.Errorf("failed to unset env var: %v", err)
		}
	}()

	mockUC := &mockPortfolioUsecase{
		backfillResult: usecase.BackfillResult{
			Created: 2,
			Status:  "success",
		},
	}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.BackfillHistoryRequest{
		AdminToken: "secret-token",
		StartDate:  "2023-01-01",
		EndDate:    "2023-01-02",
		Name:       resourcenames.PortfolioName("user-123"),
	}
	resp, err := handler.BackfillHistory(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected status success, got %s", resp.Status)
	}

	if resp.SnapshotsCreated != 2 {
		t.Errorf("Expected 2 snapshots created, got %d", resp.SnapshotsCreated)
	}
}

func TestBackfillHistory_InvalidResourceName(t *testing.T) {
	if err := os.Setenv("ADMIN_TOKEN", "secret-token"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ADMIN_TOKEN"); err != nil {
			t.Errorf("failed to unset env var: %v", err)
		}
	}()

	mockUC := &mockPortfolioUsecase{}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.BackfillHistoryRequest{
		AdminToken: "secret-token",
		StartDate:  "2023-01-01",
		Name:       "invalid-name",
	}
	_, err := handler.BackfillHistory(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid resource name, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected code InvalidArgument, got %v", st.Code())
	}
}

func TestBackfillHistory_InvalidAdminToken(t *testing.T) {
	if err := os.Setenv("ADMIN_TOKEN", "secret-token"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ADMIN_TOKEN"); err != nil {
			t.Errorf("failed to unset env var: %v", err)
		}
	}()

	mockUC := &mockPortfolioUsecase{}
	mockRepo := &mockHistoryRepo{}
	handler := NewPortfolioHandler(mockUC, mockRepo)

	req := &pb.BackfillHistoryRequest{
		AdminToken: "wrong-token",
		StartDate:  "2023-01-01",
	}
	_, err := handler.BackfillHistory(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid admin token, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected code Unauthenticated, got %v", st.Code())
	}
}
