// Package grpc implements gRPC handlers for the transaction service.
package grpc

import (
	"context"
	"database/sql"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TransactionHandler implements the gRPC transaction service.
type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	usecase domain.TransactionUsecase
}

// NewTransactionHandler creates a new transaction handler.
func NewTransactionHandler(usecase domain.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: usecase}
}

// CreateTransaction handles the creation of a new transaction.
func (h *TransactionHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	// Validate required fields
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}
	if req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}
	if req.PricePerShare < 0 {
		return nil, status.Error(codes.InvalidArgument, "price_per_share must be non-negative")
	}

	txn := &domain.Transaction{
		UserID:            req.UserId,
		Symbol:            req.Symbol,
		Type:              req.Type,
		Quantity:          req.Quantity,
		PricePerShare:     req.PricePerShare,
		ExecutedAt:        req.ExecutedAt.AsTime(),
		Brokerage:         req.Brokerage,
		Notes:             req.Notes,
		PriceCurrency:     req.PriceCurrency,
		BrokerageCurrency: req.BrokerageCurrency,
	}

	if err := h.usecase.CreateTransaction(ctx, txn); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return &pb.CreateTransactionResponse{
		Transaction: mapDomainToProto(txn),
	}, nil
}

// GetTransaction retrieves a transaction by ID.
func (h *TransactionHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	txn, err := h.usecase.GetTransaction(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}
	if txn == nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	return &pb.GetTransactionResponse{
		Transaction: mapDomainToProto(txn),
	}, nil
}

// ListTransactions lists transactions for a user.
func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// Simple pagination logic for now
	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 100
	}

	// Decode page_token → offset
	offset, err := decodeOffset(req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_token")
	}

	filter := domain.TransactionFilter{}
	if req.Filter != nil {
		filter.Symbol = req.Filter.Symbol
		filter.Type = req.Filter.Type
		if req.Filter.FromExecutedAt != nil {
			filter.FromExecutedAt = req.Filter.FromExecutedAt.AsTime()
		}
		if req.Filter.ToExecutedAt != nil {
			filter.ToExecutedAt = req.Filter.ToExecutedAt.AsTime()
		}
	}

	// Query limit+offset
	txns, err := h.usecase.ListTransactions(ctx, req.UserId, filter, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	var protoTxns []*pb.Transaction
	for _, txn := range txns {
		protoTxns = append(protoTxns, mapDomainToProto(txn))
	}

	// Compute NextPageToken
	nextToken := ""
	if len(txns) == limit {
		nextToken = encodeOffset(offset + limit)
	}

	return &pb.ListTransactionsResponse{
		Transactions:  protoTxns,
		NextPageToken: nextToken,
	}, nil
}

// UpdateTransaction updates an existing transaction.
func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	txn := &domain.Transaction{
		ID:                req.Id,
		Symbol:            req.Symbol,
		Type:              req.Type,
		Quantity:          req.Quantity,
		PricePerShare:     req.PricePerShare,
		ExecutedAt:        req.ExecutedAt.AsTime(),
		Brokerage:         req.Brokerage,
		Notes:             req.Notes,
		PriceCurrency:     req.PriceCurrency,
		BrokerageCurrency: req.BrokerageCurrency,
	}

	if err := h.usecase.UpdateTransaction(ctx, txn); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update transaction: %v", err)
	}

	return &pb.UpdateTransactionResponse{
		Transaction: mapDomainToProto(txn),
	}, nil
}

// DeleteTransaction deletes a transaction by ID.
func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
	err := h.usecase.DeleteTransaction(ctx, req.Id)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete transaction: %v", err)
	}

	return &pb.DeleteTransactionResponse{Success: true}, nil
}

// GetOldestTransactionForUser retrieves the oldest transaction for a user.
func (h *TransactionHandler) GetOldestTransactionForUser(ctx context.Context, req *pb.GetOldestTransactionForUserRequest) (*pb.GetOldestTransactionForUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	txn, err := h.usecase.GetOldestTransaction(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get oldest transaction: %v", err)
	}
	if txn == nil {
		return nil, status.Errorf(codes.NotFound, "no transactions found for user")
	}

	return &pb.GetOldestTransactionForUserResponse{
		Transaction: mapDomainToProto(txn),
	}, nil
}

func mapDomainToProto(txn *domain.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:                txn.ID,
		UserId:            txn.UserID,
		Symbol:            txn.Symbol,
		Type:              txn.Type,
		Quantity:          txn.Quantity,
		PricePerShare:     txn.PricePerShare,
		ExecutedAt:        timestamppb.New(txn.ExecutedAt),
		CreatedAt:         timestamppb.New(txn.CreatedAt),
		UpdatedAt:         timestamppb.New(txn.UpdatedAt),
		Brokerage:         txn.Brokerage,
		Notes:             txn.Notes,
		PriceCurrency:     txn.PriceCurrency,
		BrokerageCurrency: txn.BrokerageCurrency,
	}
}
