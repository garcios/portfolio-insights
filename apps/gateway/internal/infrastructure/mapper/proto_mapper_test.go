package mapper

import (
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoToUserEntity(t *testing.T) {
	pb := &userpb.User{
		UserId:   "user-1",
		Username: "testuser",
		Email:    "test@example.com",
	}

	user := ProtoToUserEntity(pb)

	if user.ID != pb.UserId {
		t.Errorf("expected ID %s, got %s", pb.UserId, user.ID)
	}
	if user.Username != pb.Username {
		t.Errorf("expected Username %s, got %s", pb.Username, user.Username)
	}
	if user.Email != pb.Email {
		t.Errorf("expected Email %s, got %s", pb.Email, user.Email)
	}
}

func TestProtoToPortfolioSummaryEntity(t *testing.T) {
	now := time.Now()
	pb := &portfoliopb.PortfolioSummary{
		TotalValue:              1000.0,
		TotalGainLoss:           100.0,
		TotalGainLossPercentage: 10.0,
		DayChange:               50.0,
		DayChangePercentage:     5.0,
		Currency:                "USD",
		LastUpdated:             timestamppb.New(now),
	}

	summary := ProtoToPortfolioSummaryEntity(pb)

	if summary.TotalValue != pb.TotalValue {
		t.Errorf("expected TotalValue %f, got %f", pb.TotalValue, summary.TotalValue)
	}
	if !summary.LastUpdated.Equal(now.UTC()) && !summary.LastUpdated.Equal(now) {
		// Protobuf timestamps can lose some precision or timezone info, so we check approximate equality or just conversion success
		// Here we just check if it's the same time instant
		if summary.LastUpdated.Unix() != now.Unix() {
			t.Errorf("expected LastUpdated %v, got %v", now, summary.LastUpdated)
		}
	}
}

func TestProtoToHoldingEntity(t *testing.T) {
	pb := &portfoliopb.Holding{
		Symbol:   "AAPL",
		Quantity: 10,
	}

	holding := ProtoToHoldingEntity(pb)

	if holding.Symbol != pb.Symbol {
		t.Errorf("expected Symbol %s, got %s", pb.Symbol, holding.Symbol)
	}
	if holding.Quantity != pb.Quantity {
		t.Errorf("expected Quantity %f, got %f", pb.Quantity, holding.Quantity)
	}
}

func TestProtoToPortfolioPerformancePointEntity(t *testing.T) {
	now := time.Now()
	pb := &portfoliopb.PortfolioPerformancePoint{
		Timestamp: timestamppb.New(now),
		Value:     100.0,
	}

	point := ProtoToPortfolioPerformancePointEntity(pb)

	if point.Value != pb.Value {
		t.Errorf("expected Value %f, got %f", pb.Value, point.Value)
	}
}

func TestProtoToTransactionEntity(t *testing.T) {
	now := time.Now()
	symbol := "AAPL"
	pb := &transactionpb.Transaction{
		TransactionId: "tx-1",
		UserId:        "user-1",
		Symbol:        &symbol,
		Type:          "BUY",
		ExecutedAt:    timestamppb.New(now),
		CreatedAt:     timestamppb.New(now),
		UpdatedAt:     timestamppb.New(now),
	}

	tx := ProtoToTransactionEntity(pb)

	if tx.ID != pb.TransactionId {
		t.Errorf("expected ID %s, got %s", pb.TransactionId, tx.ID)
	}
	if tx.Type != entity.TransactionTypeBuy {
		t.Errorf("expected Type BUY, got %s", tx.Type)
	}
}

func TestCreateTransactionInputToProto(t *testing.T) {
	now := time.Now()
	input := gateway.CreateTransactionInput{
		UserID:     "user-1",
		Symbol:     "AAPL",
		Type:       entity.TransactionTypeBuy,
		Quantity:   10,
		ExecutedAt: now,
	}

	pb := CreateTransactionInputToProto(input)

	// Check parent field (resource name)
	expectedParent := "users/user-1"
	if pb.Parent != expectedParent {
		t.Errorf("expected Parent %s, got %s", expectedParent, pb.Parent)
	}

	// Check transaction object
	if pb.Transaction == nil {
		t.Fatal("expected Transaction object, got nil")
	}

	if pb.Transaction.Symbol == nil || *pb.Transaction.Symbol != input.Symbol {
		t.Errorf("expected Symbol %s, got %v", input.Symbol, pb.Transaction.Symbol)
	}
	if pb.Transaction.Type != string(input.Type) {
		t.Errorf("expected Type %s, got %s", input.Type, pb.Transaction.Type)
	}
}
