package infrastructure

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockTransactionClient struct {
	oldestTx *transactionpb.Transaction
	err      error
}

func (m *mockTransactionClient) CreateTransaction(ctx context.Context, in *transactionpb.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionClient) GetTransaction(ctx context.Context, in *transactionpb.GetTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionClient) ListTransactions(ctx context.Context, in *transactionpb.ListTransactionsRequest, opts ...grpc.CallOption) (*transactionpb.ListTransactionsResponse, error) {
	return nil, nil
}

func (m *mockTransactionClient) UpdateTransaction(ctx context.Context, in *transactionpb.UpdateTransactionRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionClient) DeleteTransaction(ctx context.Context, in *transactionpb.DeleteTransactionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockTransactionClient) GetOldestTransactionForUser(ctx context.Context, in *transactionpb.GetOldestTransactionForUserRequest, opts ...grpc.CallOption) (*transactionpb.Transaction, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.oldestTx, nil
}

type mockPortfolioUsecase struct {
	summary *domain.PortfolioSummary
	err     error
}

func (m *mockPortfolioUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	return nil, nil
}

func (m *mockPortfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return a summary that matches the user ID if possible, or just the default one
	s := *m.summary
	s.UserID = userID
	return &s, nil
}

func (m *mockPortfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	return nil, nil
}

func (m *mockPortfolioUsecase) BackfillPortfolioHistory(
	ctx context.Context,
	userIDs []string,
	startDate, endDate time.Time,
	dryRun bool,
) usecase.BackfillResult {
	return usecase.BackfillResult{}
}

type mockPortfolioHistoryRepository struct {
	mu        sync.Mutex
	snapshots []*domain.PortfolioSnapshot
	userIDs   []string
	err       error
}

func (m *mockPortfolioHistoryRepository) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.snapshots = append(m.snapshots, snapshot)
	return nil
}

func (m *mockPortfolioHistoryRepository) GetSnapshots() []*domain.PortfolioSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to avoid race conditions on the slice if the test iterates while worker appends
	result := make([]*domain.PortfolioSnapshot, len(m.snapshots))
	copy(result, m.snapshots)
	return result
}

func (m *mockPortfolioHistoryRepository) GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error) {
	return nil, nil
}

func (m *mockPortfolioHistoryRepository) GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*domain.PortfolioSnapshot, error) {
	return nil, nil
}

func (m *mockPortfolioHistoryRepository) SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error) {
	return false, nil
}

func (m *mockPortfolioHistoryRepository) GetAllUserIDs(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.userIDs, nil
}

func TestSnapshotWorker_CreateSnapshots(t *testing.T) {
	// Setup
	uc := &mockPortfolioUsecase{
		summary: &domain.PortfolioSummary{
			UserID:     "default",
			TotalValue: 1000.0,
			TotalCost:  800.0,
		},
	}
	repo := &mockPortfolioHistoryRepository{
		userIDs: []string{"user-1", "user-2"},
	}
	mockTxClient := &mockTransactionClient{}
	logger := slog.Default()

	worker := NewSnapshotWorker(uc, repo, mockTxClient, logger)

	// Execute via TriggerNow which runs in a goroutine
	worker.TriggerNow()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Get snapshots safely
	snapshots := repo.GetSnapshots()

	// Assert
	if len(snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(snapshots))
	}

	// Verify snapshot content
	var s1 *domain.PortfolioSnapshot
	for _, s := range snapshots {
		if s.UserID == "user-1" {
			s1 = s
			break
		}
	}

	if s1 == nil {
		t.Fatal("Snapshot for user-1 not found")
	}

	if s1.TotalValue != 1000.0 {
		t.Errorf("Expected total value 1000.0, got %f", s1.TotalValue)
	}
	if s1.TotalCostBasis != 800.0 {
		t.Errorf("Expected total cost 800.0, got %f", s1.TotalCostBasis)
	}
}
