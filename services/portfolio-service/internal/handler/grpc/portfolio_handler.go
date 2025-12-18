// Package grpc implements gRPC handlers for the portfolio service.
package grpc

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PortfolioHandler implements the gRPC portfolio service.
type PortfolioHandler struct {
	pb.UnimplementedPortfolioServiceServer
	portfolioUsecase usecase.PortfolioUsecase
	historyRepo      domain.PortfolioHistoryRepository
}

// NewPortfolioHandler creates a new portfolio handler.
func NewPortfolioHandler(
	portfolioUsecase usecase.PortfolioUsecase,
	historyRepo domain.PortfolioHistoryRepository,
) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioUsecase: portfolioUsecase,
		historyRepo:      historyRepo,
	}
}

// GetHoldings retrieves all holdings for a user.
// AIP-132 compliant: uses parent field instead of user_id.
func (h *PortfolioHandler) GetHoldings(ctx context.Context, req *pb.GetHoldingsRequest) (*pb.GetHoldingsResponse, error) {
	// Parse parent resource name to get user ID
	userID, err := resourcenames.ParseHoldingParent(req.Parent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid parent: %v", err)
	}

	holdings, err := h.portfolioUsecase.GetHoldings(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get holdings: %v", err)
	}

	// Convert domain holdings to proto holdings
	pbHoldings := make([]*pb.Holding, len(holdings))
	for i, holding := range holdings {
		currentValue := holding.Quantity * holding.CurrentPrice
		costBasis := holding.Quantity * holding.AverageCost
		gainLoss := currentValue - costBasis
		gainLossPct := 0.0
		if costBasis > 0 {
			gainLossPct = (gainLoss / costBasis) * 100
		}

		pbHoldings[i] = &pb.Holding{
			Name:               resourcenames.HoldingName(userID, holding.Symbol),
			Symbol:             holding.Symbol,
			Quantity:           holding.Quantity,
			AveragePrice:       holding.AverageCost,
			CurrentPrice:       holding.CurrentPrice,
			CurrentValue:       currentValue,
			GainLoss:           gainLoss,
			GainLossPercentage: gainLossPct,
			Currency:           holding.Currency,
			AssetName:          holding.AssetName,
			HoldingId:          holding.Symbol, // Use symbol as holding ID
		}
	}

	return &pb.GetHoldingsResponse{
		Holdings: pbHoldings,
	}, nil
}

// GetPortfolioSummary retrieves the portfolio summary for a user.
// AIP-131 compliant: uses singleton resource name.
func (h *PortfolioHandler) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.PortfolioSummary, error) {
	userID, err := resourcenames.ParsePortfolioName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: %v", err)
		}
		startDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format: %v", err)
		}
		// Set end date to end of day? Or just beginning?
		// Usually inclusive end date means until 23:59:59.
		// Let's set it to valid time.
		// If transaction time is exactly at 00:00:00, it's fine.
		// But usually we want end of day.
		// Let's rely on the usecase handling, but typical convention is inclusive.
		// Let's add 23h59m59s to end date if it's 00:00:00?
		// Or just leave as is. User typically provides date.
		endDate = &t
	}

	summary, err := h.portfolioUsecase.GetPortfolioSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get portfolio summary: %v", err)
	}

	return &pb.PortfolioSummary{
		Name:                        resourcenames.PortfolioName(userID),
		UserId:                      summary.UserID,
		TotalValue:                  summary.TotalValue,
		TotalGainLoss:               summary.GainLoss,
		TotalGainLossPercentage:     summary.GainLossPct,
		DayChange:                   summary.DayChange,
		DayChangePercentage:         summary.DayChangePct,
		Currency:                    summary.Currency,
		LastUpdated:                 timestamppb.New(time.Now()),
		CapitalGainLoss:             summary.CapitalGain,
		CapitalGainLossPercentage:   summary.CapitalGainPct,
		CurrencyGainLoss:            summary.CurrencyGain,
		CurrencyGainLossPercentage:  summary.CurrencyGainPct,
		DividendsReceived:           summary.Dividends,
		DividendsReceivedPercentage: summary.DividendsPct,
		StartDate:                   timestamppb.New(summary.StartDate),
		EndDate:                     timestamppb.New(summary.EndDate),
	}, nil
}

// GetPortfolioPerformance retrieves historical portfolio performance.
// Custom method for performance data.
func (h *PortfolioHandler) GetPortfolioPerformance(ctx context.Context, req *pb.GetPortfolioPerformanceRequest) (*pb.GetPortfolioPerformanceResponse, error) {
	// Parse portfolio resource name to get user ID
	userID, err := resourcenames.ParsePortfolioName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	if req.Period == "" {
		return nil, status.Error(codes.InvalidArgument, "period is required")
	}

	// Profiling Setup
	reqID := time.Now().UnixNano()
	var (
		countGetHistoryByPeriod int
	)
	fmt.Printf("[GetPortfolioPerformance-%d] Starting request for user %s\n", reqID, userID)

	// Query historical snapshots from the database
	startHistory := time.Now()
	snapshots, err := h.historyRepo.GetHistoryByPeriod(ctx, userID, req.Period)
	countGetHistoryByPeriod++
	fmt.Printf("[GetPortfolioPerformance-%d] Step: GetHistoryByPeriod | Duration: %dms | Call Count: %d\n",
		reqID, time.Since(startHistory).Milliseconds(), countGetHistoryByPeriod)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get portfolio history: %v", err)
	}

	// Convert snapshots to performance points
	dataPoints := make([]*pb.PortfolioPerformancePoint, len(snapshots))
	for i, snapshot := range snapshots {
		dataPoints[i] = &pb.PortfolioPerformancePoint{
			Timestamp: timestamppb.New(snapshot.Timestamp),
			Value:     snapshot.TotalValue,
		}
	}

	return &pb.GetPortfolioPerformanceResponse{
		DataPoints: dataPoints,
	}, nil
}

// BackfillHistory implements the admin endpoint for backfilling portfolio history.
// Custom method for administrative operations.
func (h *PortfolioHandler) BackfillHistory(
	ctx context.Context,
	req *pb.BackfillHistoryRequest,
) (*pb.BackfillHistoryResponse, error) {
	// 1. Validate admin token
	if !h.validateAdminToken(req.AdminToken) {
		return nil, status.Error(codes.Unauthenticated, "invalid admin token")
	}

	// 2. Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date: %v", err)
	}

	endDate := time.Now()
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_date: %v", err)
		}
	}

	// 3. Determine users to backfill
	var userIDs []string
	if req.Name != "" {
		// Parse portfolio resource name if provided
		userID, err := resourcenames.ParsePortfolioName(req.Name)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
		}
		userIDs = []string{userID}
	} else {
		// Get all users with holdings
		userIDs, err = h.historyRepo.GetAllUserIDs(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
		}
	}

	// 4. Run backfill
	result := h.portfolioUsecase.BackfillPortfolioHistory(ctx, userIDs, startDate, endDate, req.DryRun)

	return &pb.BackfillHistoryResponse{
		SnapshotsCreated: int32(result.Created),
		SnapshotsSkipped: int32(result.Skipped),
		Errors:           int32(result.Errors),
		ErrorMessages:    result.ErrorMessages,
		Status:           result.Status,
	}, nil
}

func (h *PortfolioHandler) validateAdminToken(token string) bool {
	// Simple token validation
	// In production, use proper authentication (JWT, OAuth, etc.)
	adminToken := os.Getenv("ADMIN_TOKEN")
	// If ADMIN_TOKEN is not set, deny all requests for security
	if adminToken == "" {
		return false
	}
	return token != "" && token == adminToken
}
