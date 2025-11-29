package mapper

import (
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/graph/model"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

func TestUserEntityToGraphQL(t *testing.T) {
	entity := &entity.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    "test@example.com",
	}

	gql := UserEntityToGraphQL(entity)

	if gql.ID != entity.ID {
		t.Errorf("expected ID %s, got %s", entity.ID, gql.ID)
	}
	if gql.Username != entity.Username {
		t.Errorf("expected Username %s, got %s", entity.Username, gql.Username)
	}
}

func TestPortfolioEntityToGraphQL(t *testing.T) {
	entity := &entity.Portfolio{
		ID:     "p-1",
		UserID: "user-1",
		Name:   "My Portfolio",
	}

	gql := PortfolioEntityToGraphQL(entity)

	if gql.ID != entity.ID {
		t.Errorf("expected ID %s, got %s", entity.ID, gql.ID)
	}
}

func TestPortfolioSummaryEntityToGraphQL(t *testing.T) {
	now := time.Now()
	entity := &entity.PortfolioSummary{
		TotalValue:  1000.0,
		LastUpdated: now,
	}

	gql := PortfolioSummaryEntityToGraphQL(entity)

	if gql.TotalValue != entity.TotalValue {
		t.Errorf("expected TotalValue %f, got %f", entity.TotalValue, gql.TotalValue)
	}
	expectedTime := now.Format("2006-01-02T15:04:05Z07:00")
	if gql.LastUpdated != expectedTime {
		t.Errorf("expected LastUpdated %s, got %s", expectedTime, gql.LastUpdated)
	}
}

func TestHoldingEntityToGraphQL(t *testing.T) {
	entity := &entity.Holding{
		Symbol:   "AAPL",
		Quantity: 10,
	}

	gql := HoldingEntityToGraphQL(entity)

	if gql.Symbol != entity.Symbol {
		t.Errorf("expected Symbol %s, got %s", entity.Symbol, gql.Symbol)
	}
}

func TestTransactionEntityToGraphQL(t *testing.T) {
	now := time.Now()
	entity := &entity.Transaction{
		ID:         "tx-1",
		Type:       entity.TransactionTypeBuy,
		ExecutedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	gql := TransactionEntityToGraphQL(entity)

	if gql.ID != entity.ID {
		t.Errorf("expected ID %s, got %s", entity.ID, gql.ID)
	}
	if gql.Type != model.TransactionTypeBuy {
		t.Errorf("expected Type BUY, got %s", gql.Type)
	}
}

func TestGraphQLNewTransactionToUseCaseInput(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	input := model.NewTransaction{
		Symbol:        "AAPL",
		Type:          model.TransactionTypeBuy,
		Quantity:      10,
		PricePerShare: 150.0,
		ExecutedAt:    nowStr,
	}

	useCaseInput, err := GraphQLNewTransactionToUseCaseInput(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if useCaseInput.Symbol != input.Symbol {
		t.Errorf("expected Symbol %s, got %s", input.Symbol, useCaseInput.Symbol)
	}
	if useCaseInput.Type != entity.TransactionTypeBuy {
		t.Errorf("expected Type BUY, got %s", useCaseInput.Type)
	}
}
