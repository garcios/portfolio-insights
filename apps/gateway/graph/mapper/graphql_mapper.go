package mapper

import (
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/usecase"
)

// parseTimestamp parses a timestamp string in RFC3339 format
func parseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %w", err)
	}
	return t, nil
}

// UserEntityToGraphQL converts a User entity to a GraphQL User model
func UserEntityToGraphQL(user *entity.User) *model.User {
	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// PortfolioEntityToGraphQL converts a Portfolio entity to a GraphQL Portfolio model
func PortfolioEntityToGraphQL(portfolio *entity.Portfolio) *model.Portfolio {
	return &model.Portfolio{
		ID:     portfolio.ID,
		UserID: portfolio.UserID,
		Name:   portfolio.Name,
		// Summary and Holdings are loaded via field resolvers
	}
}

// PortfolioSummaryEntityToGraphQL converts a PortfolioSummary entity to a GraphQL PortfolioSummary model
func PortfolioSummaryEntityToGraphQL(summary *entity.PortfolioSummary) *model.PortfolioSummary {
	return &model.PortfolioSummary{
		TotalValue:              summary.TotalValue,
		TotalGainLoss:           summary.TotalGainLoss,
		TotalGainLossPercentage: summary.TotalGainLossPercentage,
		DayChange:               summary.DayChange,
		DayChangePercentage:     summary.DayChangePercentage,
		Currency:                summary.Currency,
		LastUpdated:             summary.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// HoldingEntityToGraphQL converts a Holding entity to a GraphQL Holding model
func HoldingEntityToGraphQL(holding *entity.Holding) *model.Holding {
	return &model.Holding{
		Symbol:             holding.Symbol,
		Quantity:           holding.Quantity,
		AveragePrice:       holding.AveragePrice,
		CurrentPrice:       holding.CurrentPrice,
		CurrentValue:       holding.CurrentValue,
		GainLoss:           holding.GainLoss,
		GainLossPercentage: holding.GainLossPercentage,
		Currency:           holding.Currency,
		AssetName:          holding.AssetName,
	}
}

// HoldingEntitiesToGraphQL converts a slice of Holding entities to GraphQL Holding models
func HoldingEntitiesToGraphQL(holdings []*entity.Holding) []*model.Holding {
	result := make([]*model.Holding, 0, len(holdings))
	for _, h := range holdings {
		result = append(result, HoldingEntityToGraphQL(h))
	}
	return result
}

// PortfolioPerformancePointEntityToGraphQL converts a PortfolioPerformancePoint entity to a GraphQL model
func PortfolioPerformancePointEntityToGraphQL(point *entity.PortfolioPerformancePoint) *model.PortfolioPerformancePoint {
	return &model.PortfolioPerformancePoint{
		Timestamp: point.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Value:     point.Value,
	}
}

// PortfolioPerformancePointEntitiesToGraphQL converts a slice of PortfolioPerformancePoint entities to GraphQL models
func PortfolioPerformancePointEntitiesToGraphQL(points []*entity.PortfolioPerformancePoint) []*model.PortfolioPerformancePoint {
	result := make([]*model.PortfolioPerformancePoint, 0, len(points))
	for _, p := range points {
		result = append(result, PortfolioPerformancePointEntityToGraphQL(p))
	}
	return result
}

// TransactionEntityToGraphQL converts a Transaction entity to a GraphQL Transaction model
func TransactionEntityToGraphQL(tx *entity.Transaction) *model.Transaction {
	notes := tx.Notes
	brokerage := tx.Brokerage
	priceCurrency := tx.PriceCurrency
	brokerageCurrency := tx.BrokerageCurrency

	return &model.Transaction{
		ID:                tx.ID,
		UserID:            tx.UserID,
		Symbol:            tx.Symbol,
		Type:              model.TransactionType(tx.Type),
		Quantity:          tx.Quantity,
		PricePerShare:     tx.PricePerShare,
		PriceCurrency:     &priceCurrency,
		ExecutedAt:        tx.ExecutedAt.Format("2006-01-02T15:04:05Z07:00"),
		Notes:             &notes,
		Brokerage:         &brokerage,
		BrokerageCurrency: &brokerageCurrency,
		CreatedAt:         tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         tx.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GraphQLNewTransactionToUseCaseInput converts a GraphQL NewTransaction input to a use case CreateTransactionInput
func GraphQLNewTransactionToUseCaseInput(input model.NewTransaction) (usecase.CreateTransactionInput, error) {
	executedAt, err := parseTimestamp(input.ExecutedAt)
	if err != nil {
		return usecase.CreateTransactionInput{}, err
	}

	result := usecase.CreateTransactionInput{
		Symbol:        input.Symbol,
		Type:          entity.TransactionType(input.Type),
		Quantity:      input.Quantity,
		PricePerShare: input.PricePerShare,
		ExecutedAt:    executedAt,
	}

	// Handle optional fields
	if input.Notes != nil {
		result.Notes = *input.Notes
	}

	if input.Brokerage != nil {
		result.Brokerage = *input.Brokerage
	}

	if input.PriceCurrency != nil {
		result.PriceCurrency = *input.PriceCurrency
	}

	if input.BrokerageCurrency != nil {
		result.BrokerageCurrency = *input.BrokerageCurrency
	}

	return result, nil
}
