// Package grpc implements gRPC clients.
package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/mapper"
	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
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

// GetPortfolioSummary retrieves portfolio summary metrics using AIP singleton resource
func (g *PortfolioGRPCGateway) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *string) (*entity.PortfolioSummary, error) {
	req := &portfoliopb.GetPortfolioSummaryRequest{
		Name:      resourcenames.PortfolioName(userID),
		StartDate: startDate,
		EndDate:   endDate,
	}

	resp, err := g.client.GetPortfolioSummary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio summary: %w", err)
	}

	if resp == nil {
		return nil, nil
	}

	return mapper.ProtoToPortfolioSummaryEntity(resp), nil
}

// GetHoldings retrieves all holdings for a user using AIP parent field
func (g *PortfolioGRPCGateway) GetHoldings(ctx context.Context, userID string) ([]*entity.Holding, error) {
	req := &portfoliopb.GetHoldingsRequest{
		Parent: resourcenames.UserName(userID),
	}

	resp, err := g.client.GetHoldings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	holdings := make([]*entity.Holding, 0, len(resp.Holdings))
	for _, h := range resp.Holdings {
		holding := mapper.ProtoToHoldingEntity(h)
		holding.UserID = userID
		holdings = append(holdings, holding)
	}

	return holdings, nil
}

// GetPortfolioPerformance retrieves historical performance data using AIP singleton resource
func (g *PortfolioGRPCGateway) GetPortfolioPerformance(ctx context.Context, userID, period string) ([]*entity.PortfolioPerformancePoint, error) {
	req := &portfoliopb.GetPortfolioPerformanceRequest{
		Name:   resourcenames.PortfolioName(userID),
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
