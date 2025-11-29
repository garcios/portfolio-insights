package service

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model" // Assuming your Go module path
	"github.com/garcios/portfolio-insights/apps/gateway/internal/util"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
)

// TransactionService wraps the gRPC client
type TransactionService struct {
	client transactionpb.TransactionServiceClient
}

func NewTransactionService(client transactionpb.TransactionServiceClient) *TransactionService {
	return &TransactionService{client: client}
}

// Create maps the GraphQL input to gRPC request, calls gRPC, and maps gRPC response to GraphQL model.
func (s *TransactionService) Create(ctx context.Context, userID string, input model.NewTransaction) (*model.Transaction, error) {
	executedAt, err := util.ParseTimestamp(input.ExecutedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid executedAt timestamp: %w", err)
	}

	// Create transaction request
	req := &transactionpb.CreateTransactionRequest{
		UserId:            userID,
		Symbol:            input.Symbol,
		Type:              string(input.Type),
		Quantity:          input.Quantity,
		PricePerShare:     input.PricePerShare,
		ExecutedAt:        executedAt,
		Notes:             util.DerefString(input.Notes),
		PriceCurrency:     util.DerefString(input.PriceCurrency),
		BrokerageCurrency: util.DerefString(input.BrokerageCurrency),
	}

	if input.Brokerage != nil {
		req.Brokerage = *input.Brokerage
	}

	// Call transaction service
	resp, err := s.client.CreateTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Map response to GraphQL model
	return &model.Transaction{
		ID:                resp.Transaction.Id,
		UserID:            resp.Transaction.UserId,
		Symbol:            resp.Transaction.Symbol,
		Type:              model.TransactionType(resp.Transaction.Type),
		Quantity:          resp.Transaction.Quantity,
		PricePerShare:     resp.Transaction.PricePerShare,
		PriceCurrency:     &resp.Transaction.PriceCurrency,
		ExecutedAt:        util.FormatTime(resp.Transaction.ExecutedAt),
		Notes:             &resp.Transaction.Notes,
		Brokerage:         &resp.Transaction.Brokerage,
		BrokerageCurrency: &resp.Transaction.BrokerageCurrency,
		CreatedAt:         util.FormatTime(resp.Transaction.CreatedAt),
		UpdatedAt:         util.FormatTime(resp.Transaction.UpdatedAt),
	}, nil
}
