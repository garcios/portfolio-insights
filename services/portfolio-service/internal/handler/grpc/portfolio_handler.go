package grpc

import (
	"context"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PortfolioHandler struct {
	pb.UnimplementedPortfolioServiceServer
	portfolioUsecase usecase.PortfolioUsecase
}

func NewPortfolioHandler(portfolioUsecase usecase.PortfolioUsecase) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioUsecase: portfolioUsecase,
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
			LastUpdated:             timestamppb.New(time.Now()),
		},
	}, nil
}

// GetPortfolioPerformance retrieves historical portfolio performance
// Note: This is a stub implementation - you'll need to implement portfolio_history tracking
func (h *PortfolioHandler) GetPortfolioPerformance(ctx context.Context, req *pb.GetPortfolioPerformanceRequest) (*pb.GetPortfolioPerformanceResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// TODO: Implement portfolio performance tracking
	// This would require:
	// 1. A portfolio_history table (which already exists in your schema)
	// 2. A scheduled job to snapshot portfolio values
	// 3. Query logic to retrieve historical data based on the period

	// For now, return empty data
	return &pb.GetPortfolioPerformanceResponse{
		DataPoints: []*pb.PortfolioPerformancePoint{},
	}, nil
}
