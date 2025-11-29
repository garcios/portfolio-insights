package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/mapper"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
)

// TransactionGRPCGateway implements the TransactionGateway interface using gRPC
type TransactionGRPCGateway struct {
	client transactionpb.TransactionServiceClient
}

// NewTransactionGRPCGateway creates a new TransactionGRPCGateway
func NewTransactionGRPCGateway(client transactionpb.TransactionServiceClient) gateway.TransactionGateway {
	return &TransactionGRPCGateway{
		client: client,
	}
}

// CreateTransaction creates a new transaction
func (g *TransactionGRPCGateway) CreateTransaction(ctx context.Context, input gateway.CreateTransactionInput) (*entity.Transaction, error) {
	req := mapper.CreateTransactionInputToProto(input)

	resp, err := g.client.CreateTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return mapper.ProtoToTransactionEntity(resp.Transaction), nil
}
