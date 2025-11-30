package mapper

import (
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToUserEntity converts a protobuf GetUserResponse to a User entity
func ProtoToUserEntity(resp *userpb.GetUserResponse) *entity.User {
	return entity.NewUser(resp.Id, resp.Username, resp.Email)
}

// ProtoToPortfolioSummaryEntity converts a protobuf PortfolioSummary to a PortfolioSummary entity
func ProtoToPortfolioSummaryEntity(pb *portfoliopb.PortfolioSummary) *entity.PortfolioSummary {
	return &entity.PortfolioSummary{
		TotalValue:              pb.TotalValue,
		TotalGainLoss:           pb.TotalGainLoss,
		TotalGainLossPercentage: pb.TotalGainLossPercentage,
		DayChange:               pb.DayChange,
		DayChangePercentage:     pb.DayChangePercentage,
		Currency:                pb.Currency,
		LastUpdated:             pb.LastUpdated.AsTime(),
	}
}

// ProtoToHoldingEntity converts a protobuf Holding to a Holding entity
func ProtoToHoldingEntity(pb *portfoliopb.Holding) *entity.Holding {
	return &entity.Holding{
		Symbol:             pb.Symbol,
		Quantity:           pb.Quantity,
		AveragePrice:       pb.AveragePrice,
		CurrentPrice:       pb.CurrentPrice,
		CurrentValue:       pb.CurrentValue,
		GainLoss:           pb.GainLoss,
		GainLossPercentage: pb.GainLossPercentage,
		Currency:           pb.Currency,
		AssetName:          pb.AssetName,
	}
}

// ProtoToPortfolioPerformancePointEntity converts a protobuf performance data point to a PortfolioPerformancePoint entity
func ProtoToPortfolioPerformancePointEntity(pb *portfoliopb.PortfolioPerformancePoint) *entity.PortfolioPerformancePoint {
	return &entity.PortfolioPerformancePoint{
		Timestamp: pb.Timestamp.AsTime(),
		Value:     pb.Value,
	}
}

// ProtoToTransactionEntity converts a protobuf Transaction to a Transaction entity
func ProtoToTransactionEntity(pb *transactionpb.Transaction) *entity.Transaction {
	return &entity.Transaction{
		ID:                pb.Id,
		UserID:            pb.UserId,
		Symbol:            pb.Symbol,
		Type:              entity.TransactionType(pb.Type),
		Quantity:          pb.Quantity,
		PricePerShare:     pb.PricePerShare,
		PriceCurrency:     pb.PriceCurrency,
		ExecutedAt:        pb.ExecutedAt.AsTime(),
		Notes:             pb.Notes,
		Brokerage:         pb.Brokerage,
		BrokerageCurrency: pb.BrokerageCurrency,
		CreatedAt:         pb.CreatedAt.AsTime(),
		UpdatedAt:         pb.UpdatedAt.AsTime(),
	}
}

// CreateTransactionInputToProto converts a CreateTransactionInput to a protobuf CreateTransactionRequest
func CreateTransactionInputToProto(input gateway.CreateTransactionInput) *transactionpb.CreateTransactionRequest {
	return &transactionpb.CreateTransactionRequest{
		UserId:            input.UserID,
		Symbol:            input.Symbol,
		Type:              string(input.Type),
		Quantity:          input.Quantity,
		PricePerShare:     input.PricePerShare,
		PriceCurrency:     input.PriceCurrency,
		ExecutedAt:        timestamppb.New(input.ExecutedAt),
		Notes:             input.Notes,
		Brokerage:         input.Brokerage,
		BrokerageCurrency: input.BrokerageCurrency,
	}
}
