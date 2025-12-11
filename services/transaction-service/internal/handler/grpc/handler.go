// Package grpc implements gRPC handlers for the transaction service.
package grpc

import (
	"context"
	"database/sql"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
// AIP-133 compliant: accepts Transaction object and parent field.
func (h *TransactionHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.Transaction, error) {
	// Validate request
	if req.Transaction == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction is required")
	}

	// Parse parent resource name to get user ID
	userID, err := resourcenames.ParseTransactionParent(req.Parent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid parent: %v", err)
	}

	// Validate required fields
	if req.Transaction.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "transaction.type is required")
	}
	if req.Transaction.ExecutedAt == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction.executed_at is required")
	}

	txn := &domain.Transaction{
		UserID:            userID,
		Type:              req.Transaction.Type,
		ExecutedAt:        req.Transaction.ExecutedAt.AsTime(),
		Brokerage:         req.Transaction.Brokerage,
		Notes:             req.Transaction.Notes,
		PriceCurrency:     req.Transaction.PriceCurrency,
		BrokerageCurrency: req.Transaction.BrokerageCurrency,
	}

	// Set nullable fields from proto optional fields
	if req.Transaction.Symbol != nil {
		txn.Symbol = req.Transaction.Symbol
	}
	if req.Transaction.Quantity != nil {
		txn.Quantity = req.Transaction.Quantity
	}
	if req.Transaction.PricePerShare != nil {
		txn.PricePerShare = req.Transaction.PricePerShare
	}
	if req.Transaction.Amount != nil {
		txn.Amount = req.Transaction.Amount
	}

	// Use client-specified transaction ID if provided
	if req.TransactionId != "" {
		txn.ID = req.TransactionId
	}

	if err := h.usecase.CreateTransaction(ctx, txn); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return mapDomainToProto(txn), nil
}

// GetTransaction retrieves a transaction by its resource name.
// AIP-131 compliant: uses resource name instead of ID.
func (h *TransactionHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.Transaction, error) {
	// Parse hierarchical resource name
	userID, txnID, err := resourcenames.ParseTransactionName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	txn, err := h.usecase.GetTransaction(ctx, txnID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}
	if txn == nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	// Verify transaction belongs to the user (security check)
	if txn.UserID != userID {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	return mapDomainToProto(txn), nil
}

// ListTransactions lists transactions for a user.
// AIP-132 compliant: uses parent field instead of user_id.
func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// Parse parent resource name to get user ID
	userID, err := resourcenames.ParseTransactionParent(req.Parent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid parent: %v", err)
	}

	// Simple pagination logic
	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 50 // Default page size per AIP-132
	}
	if limit > 1000 {
		limit = 1000 // Max page size per AIP-132
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
	txns, err := h.usecase.ListTransactions(ctx, userID, filter, limit, offset)
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
// AIP-134 compliant: uses Transaction object and FieldMask.
func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.Transaction, error) {
	// Validate request
	if req.Transaction == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction is required")
	}

	// Parse hierarchical resource name
	userID, txnID, err := resourcenames.ParseTransactionName(req.Transaction.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	// Get existing transaction
	existing, err := h.usecase.GetTransaction(ctx, txnID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}
	if existing == nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	// Verify transaction belongs to the user (security check)
	if existing.UserID != userID {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	// Apply field mask
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		// Update all mutable fields
		existing.Type = req.Transaction.Type
		existing.Symbol = req.Transaction.Symbol
		existing.Quantity = req.Transaction.Quantity
		existing.PricePerShare = req.Transaction.PricePerShare
		existing.Amount = req.Transaction.Amount
		if req.Transaction.ExecutedAt != nil {
			existing.ExecutedAt = req.Transaction.ExecutedAt.AsTime()
		}
		existing.Brokerage = req.Transaction.Brokerage
		existing.Notes = req.Transaction.Notes
		existing.PriceCurrency = req.Transaction.PriceCurrency
		existing.BrokerageCurrency = req.Transaction.BrokerageCurrency
	} else {
		// Update only specified fields
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "type":
				existing.Type = req.Transaction.Type
			case "symbol":
				existing.Symbol = req.Transaction.Symbol
			case "quantity":
				existing.Quantity = req.Transaction.Quantity
			case "price_per_share":
				existing.PricePerShare = req.Transaction.PricePerShare
			case "amount":
				existing.Amount = req.Transaction.Amount
			case "executed_at":
				if req.Transaction.ExecutedAt != nil {
					existing.ExecutedAt = req.Transaction.ExecutedAt.AsTime()
				}
			case "brokerage":
				existing.Brokerage = req.Transaction.Brokerage
			case "notes":
				existing.Notes = req.Transaction.Notes
			case "price_currency":
				existing.PriceCurrency = req.Transaction.PriceCurrency
			case "brokerage_currency":
				existing.BrokerageCurrency = req.Transaction.BrokerageCurrency
			default:
				return nil, status.Errorf(codes.InvalidArgument, "invalid field path: %s", path)
			}
		}
	}

	if err := h.usecase.UpdateTransaction(ctx, existing); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update transaction: %v", err)
	}

	return mapDomainToProto(existing), nil
}

// DeleteTransaction deletes a transaction by its resource name.
// AIP-135 compliant: returns Empty instead of custom response.
func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*emptypb.Empty, error) {
	// Parse hierarchical resource name
	userID, txnID, err := resourcenames.ParseTransactionName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	// Get transaction to verify ownership
	txn, err := h.usecase.GetTransaction(ctx, txnID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}
	if txn == nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	// Verify transaction belongs to the user (security check)
	if txn.UserID != userID {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	err = h.usecase.DeleteTransaction(ctx, txnID)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete transaction: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// GetOldestTransactionForUser retrieves the oldest transaction for a user.
// This is a custom method.
func (h *TransactionHandler) GetOldestTransactionForUser(ctx context.Context, req *pb.GetOldestTransactionForUserRequest) (*pb.Transaction, error) {
	// Parse parent resource name to get user ID
	userID, err := resourcenames.ParseTransactionParent(req.Parent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid parent: %v", err)
	}

	txn, err := h.usecase.GetOldestTransaction(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get oldest transaction: %v", err)
	}
	if txn == nil {
		return nil, status.Errorf(codes.NotFound, "no transactions found for user")
	}

	return mapDomainToProto(txn), nil
}

func mapDomainToProto(txn *domain.Transaction) *pb.Transaction {
	pbTxn := &pb.Transaction{
		Name:              resourcenames.TransactionName(txn.UserID, txn.ID),
		UserId:            txn.UserID,
		Type:              txn.Type,
		ExecutedAt:        timestamppb.New(txn.ExecutedAt),
		CreatedAt:         timestamppb.New(txn.CreatedAt),
		UpdatedAt:         timestamppb.New(txn.UpdatedAt),
		Brokerage:         txn.Brokerage,
		Notes:             txn.Notes,
		PriceCurrency:     txn.PriceCurrency,
		BrokerageCurrency: txn.BrokerageCurrency,
		TransactionId:     txn.ID,
	}

	// Set optional fields if present
	if txn.Symbol != nil {
		pbTxn.Symbol = txn.Symbol
	}
	if txn.Quantity != nil {
		pbTxn.Quantity = txn.Quantity
	}
	if txn.PricePerShare != nil {
		pbTxn.PricePerShare = txn.PricePerShare
	}
	if txn.Amount != nil {
		pbTxn.Amount = txn.Amount
	}

	return pbTxn
}
