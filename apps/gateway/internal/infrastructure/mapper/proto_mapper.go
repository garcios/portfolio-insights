// Package mapper provides functions to map between domain entities and Proto models.
package mapper

import (
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToUserEntity converts a protobuf User to a User entity
func ProtoToUserEntity(user *userpb.User) *entity.User {
	u := entity.NewUser(user.UserId, user.Username, user.Email)
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.Role = user.Role

	if user.LastLoginAt != nil {
		t := user.LastLoginAt.AsTime()
		u.LastLoginAt = &t
	}

	if user.Preferences != nil {
		u.Preferences = user.Preferences.AsMap()
	}

	return u
}

// ProtoToPortfolioSummaryEntity converts a protobuf PortfolioSummary to a PortfolioSummary entity
func ProtoToPortfolioSummaryEntity(pb *portfoliopb.PortfolioSummary) *entity.PortfolioSummary {
	summary := &entity.PortfolioSummary{
		TotalValue:              pb.TotalValue,
		TotalGainLoss:           pb.TotalGainLoss,
		TotalGainLossPercentage: pb.TotalGainLossPercentage,
		DayChange:               pb.DayChange,
		DayChangePercentage:     pb.DayChangePercentage,
		Currency:                pb.Currency,
		LastUpdated:             pb.LastUpdated.AsTime(),
		CapitalGain:             pb.CapitalGainLoss,
		CapitalGainPercentage:   pb.CapitalGainLossPercentage,
		CurrencyGain:            pb.CurrencyGainLoss,
		CurrencyGainPercentage:  pb.CurrencyGainLossPercentage,
		Dividends:               pb.DividendsReceived,
		DividendsPercentage:     pb.DividendsReceivedPercentage,
	}

	if pb.StartDate != nil {
		t := pb.StartDate.AsTime()
		summary.StartDate = &t
	}
	if pb.EndDate != nil {
		t := pb.EndDate.AsTime()
		summary.EndDate = &t
	}

	return summary
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
	txn := &entity.Transaction{
		ID:                pb.TransactionId,
		UserID:            pb.UserId,
		Type:              entity.TransactionType(pb.Type),
		PriceCurrency:     pb.PriceCurrency,
		ExecutedAt:        pb.ExecutedAt.AsTime(),
		Notes:             pb.Notes,
		Brokerage:         pb.Brokerage,
		BrokerageCurrency: pb.BrokerageCurrency,
		CreatedAt:         pb.CreatedAt.AsTime(),
		UpdatedAt:         pb.UpdatedAt.AsTime(),
	}

	// Handle nullable fields
	if pb.Symbol != nil {
		txn.Symbol = *pb.Symbol
	}
	if pb.Quantity != nil {
		txn.Quantity = *pb.Quantity
	}
	if pb.PricePerShare != nil {
		txn.PricePerShare = *pb.PricePerShare
	}

	return txn
}

// CreateTransactionInputToProto converts a CreateTransactionInput to a protobuf CreateTransactionRequest
func CreateTransactionInputToProto(input gateway.CreateTransactionInput) *transactionpb.CreateTransactionRequest {
	transaction := &transactionpb.Transaction{
		Type:              string(input.Type),
		PriceCurrency:     input.PriceCurrency,
		ExecutedAt:        timestamppb.New(input.ExecutedAt),
		Notes:             input.Notes,
		Brokerage:         input.Brokerage,
		BrokerageCurrency: input.BrokerageCurrency,
	}

	// Set nullable fields if non-zero
	if input.Symbol != "" {
		transaction.Symbol = &input.Symbol
	}
	if input.Quantity != 0 {
		transaction.Quantity = &input.Quantity
	}
	if input.PricePerShare != 0 {
		transaction.PricePerShare = &input.PricePerShare
	}

	return &transactionpb.CreateTransactionRequest{
		Parent:      resourcenames.UserName(input.UserID),
		Transaction: transaction,
	}
}

// ListTransactionsInputToProto converts a ListTransactionsInput to a protobuf ListTransactionsRequest
func ListTransactionsInputToProto(input gateway.ListTransactionsInput) *transactionpb.ListTransactionsRequest {
	req := &transactionpb.ListTransactionsRequest{
		Parent:    resourcenames.UserName(input.UserID),
		PageSize:  input.PageSize,
		PageToken: input.PageToken,
	}

	// Add filter if provided
	if input.Filter != nil {
		filter := &transactionpb.TransactionFilter{}

		if input.Filter.Symbol != nil {
			filter.Symbol = *input.Filter.Symbol
		}

		if input.Filter.Type != nil {
			filter.Type = string(*input.Filter.Type)
		}

		if input.Filter.FromExecutedAt != nil {
			filter.FromExecutedAt = timestamppb.New(*input.Filter.FromExecutedAt)
		}

		if input.Filter.ToExecutedAt != nil {
			filter.ToExecutedAt = timestamppb.New(*input.Filter.ToExecutedAt)
		}

		req.Filter = filter
	}

	return req
}

// ProtoToListTransactionsResult converts a protobuf ListTransactionsResponse to a ListTransactionsResult
func ProtoToListTransactionsResult(resp *transactionpb.ListTransactionsResponse) *gateway.ListTransactionsResult {
	transactions := make([]*entity.Transaction, len(resp.Transactions))
	for i, pb := range resp.Transactions {
		transactions[i] = ProtoToTransactionEntity(pb)
	}

	return &gateway.ListTransactionsResult{
		Transactions:  transactions,
		NextPageToken: resp.NextPageToken,
	}
}
