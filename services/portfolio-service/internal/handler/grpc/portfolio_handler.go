// Package grpc implements gRPC handlers for the portfolio service.
package grpc

import (
	"context"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
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

// GetHoldings retrieves all holdings for a user
func (h *PortfolioHandler) GetHoldings(ctx context.Context, req *pb.GetHoldingsRequest) (*pb.GetHoldingsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	holdings, err := h.portfolioUsecase.GetHoldings(ctx, req.UserId)
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
			Symbol:             holding.Symbol,
			Quantity:           holding.Quantity,
			AveragePrice:       holding.AverageCost,
			CurrentPrice:       holding.CurrentPrice,
			CurrentValue:       currentValue,
			GainLoss:           gainLoss,
			GainLossPercentage: gainLossPct,
			Currency:           holding.Currency,
			AssetName:          holding.AssetName,
		}
	}

	return &pb.GetHoldingsResponse{
		Holdings: pbHoldings,
	}, nil
}

// GetPortfolioSummary retrieves the portfolio summary for a user
func (h *PortfolioHandler) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.GetPortfolioSummaryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	summary, err := h.portfolioUsecase.GetPortfolioSummary(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get portfolio summary: %v", err)
	}

	return &pb.GetPortfolioSummaryResponse{
		Summary: &pb.PortfolioSummary{
			UserId:                  summary.UserID,
			TotalValue:              summary.TotalValue,
			TotalGainLoss:           summary.GainLoss,
			TotalGainLossPercentage: summary.GainLossPct,
			DayChange:               summary.DayChange,
			DayChangePercentage:     summary.DayChangePct,
			Currency:                summary.Currency,
			LastUpdated:             timestamppb.New(time.Now()),
		},
	}, nil
}

// GetPortfolioPerformance retrieves historical portfolio performance
func (h *PortfolioHandler) GetPortfolioPerformance(ctx context.Context, req *pb.GetPortfolioPerformanceRequest) (*pb.GetPortfolioPerformanceResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.Period == "" {
		return nil, status.Error(codes.InvalidArgument, "period is required")
	}

	// Query historical snapshots from the database
	snapshots, err := h.historyRepo.GetHistoryByPeriod(ctx, req.UserId, req.Period)
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

// BackfillHistory implements the admin endpoint for backfilling portfolio history
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
	if req.UserId != "" {
		userIDs = []string{req.UserId}
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
