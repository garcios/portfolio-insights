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

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	usecase domain.TransactionUsecase
}

func NewTransactionHandler(usecase domain.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: usecase}
}

func (h *TransactionHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	if req.UserId == "" || req.Symbol == "" || req.Type == "" || req.Quantity <= 0 || req.PricePerShare < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid arguments")
	}

	tx, err := h.usecase.CreateTransaction(
		ctx,
		req.UserId,
		req.Symbol,
		req.Type,
		req.Quantity,
		req.PricePerShare,
		req.ExecutedAt.AsTime(),
	)
	if err != nil {
		// Check if it's a user not found error
		// In a real app, we'd have custom error types. For now, string check or generic error.
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return &pb.CreateTransactionResponse{
		Transaction: convertToProto(tx),
	}, nil
}

func (h *TransactionHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	tx, err := h.usecase.GetTransaction(ctx, req.Id)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}

	return &pb.GetTransactionResponse{
		Transaction: convertToProto(tx),
	}, nil
}

func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	txs, nextPageToken, err := h.usecase.ListTransactions(ctx, req.UserId, pageSize, req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	var pbTxs []*pb.Transaction
	for _, tx := range txs {
		pbTxs = append(pbTxs, convertToProto(tx))
	}

	return &pb.ListTransactionsResponse{
		Transactions:  pbTxs,
		NextPageToken: nextPageToken,
	}, nil
}

func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	tx, err := h.usecase.UpdateTransaction(
		ctx,
		req.Id,
		req.Symbol,
		req.Type,
		req.Quantity,
		req.PricePerShare,
		req.ExecutedAt.AsTime(),
	)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update transaction: %v", err)
	}

	return &pb.UpdateTransactionResponse{
		Transaction: convertToProto(tx),
	}, nil
}

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

func convertToProto(tx *domain.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:            tx.ID,
		UserId:        tx.UserID,
		Symbol:        tx.Symbol,
		Type:          tx.Type,
		Quantity:      tx.Quantity,
		PricePerShare: tx.PricePerShare,
		ExecutedAt:    timestamppb.New(tx.ExecutedAt),
		CreatedAt:     timestamppb.New(tx.CreatedAt),
		UpdatedAt:     timestamppb.New(tx.UpdatedAt),
	}
}
