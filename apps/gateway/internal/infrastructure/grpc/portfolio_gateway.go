package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/mapper"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
)

// PortfolioGRPCGateway implements the PortfolioGateway interface using gRPC
type PortfolioGRPCGateway struct {
	client portfoliopb.PortfolioServiceClient
}

// NewPortfolioGRPCGateway creates a new PortfolioGRPCGateway
func NewPortfolioGRPCGateway(client portfoliopb.PortfolioServiceClient) gateway.PortfolioGateway {
	return &PortfolioGRPCGateway{
		client: client,
	}
}

// GetPortfolio retrieves a portfolio by user ID
func (g *PortfolioGRPCGateway) GetPortfolio(ctx context.Context, userID string) (*entity.Portfolio, error) {
	// For now, we just return a basic portfolio
	// The summary and holdings are loaded separately
	return entity.NewPortfolio(userID, userID, "My Portfolio"), nil
}

// GetPortfolioSummary retrieves portfolio summary metrics
func (g *PortfolioGRPCGateway) GetPortfolioSummary(ctx context.Context, userID string) (*entity.PortfolioSummary, error) {
	req := &portfoliopb.GetPortfolioSummaryRequest{
		UserId: userID,
	}

	resp, err := g.client.GetPortfolioSummary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio summary: %w", err)
	}

	if resp.Summary == nil {
		return nil, nil
	}

	return mapper.ProtoToPortfolioSummaryEntity(resp.Summary), nil
}

// GetHoldings retrieves all holdings for a user
func (g *PortfolioGRPCGateway) GetHoldings(ctx context.Context, userID string) ([]*entity.Holding, error) {
	req := &portfoliopb.GetHoldingsRequest{
		UserId: userID,
	}

	resp, err := g.client.GetHoldings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	holdings := make([]*entity.Holding, 0, len(resp.Holdings))
	for _, h := range resp.Holdings {
		holdings = append(holdings, mapper.ProtoToHoldingEntity(h))
	}

	return holdings, nil
}

// GetPortfolioPerformance retrieves historical performance data
func (g *PortfolioGRPCGateway) GetPortfolioPerformance(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error) {
	req := &portfoliopb.GetPortfolioPerformanceRequest{
		UserId: userID,
		Period: period,
	}

	resp, err := g.client.GetPortfolioPerformance(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio performance: %w", err)
	}

	dataPoints := make([]*entity.PortfolioPerformancePoint, 0, len(resp.DataPoints))
	for _, dp := range resp.DataPoints {
		dataPoints = append(dataPoints, mapper.ProtoToPortfolioPerformancePointEntity(dp))
	}

	return dataPoints, nil
}
