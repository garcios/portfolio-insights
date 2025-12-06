package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/config"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	"github.com/nats-io/nats.go"
)

// TransactionCreatedEvent represents the event payload for a created transaction.
type TransactionCreatedEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	AssetSymbol   string    `json:"asset_symbol"`
	PricePerShare float64   `json:"price_per_share"`
	Quantity      float64   `json:"quantity"`
	Type          string    `json:"type"`
	ExecutedAt    time.Time `json:"executed_at"`
}

type natsEventPublisher struct {
	nc    *nats.Conn
	topic string
}

// NewNATSEventPublisher creates a new NATS event publisher.
func NewNATSEventPublisher(cfg config.Config) (domain.EventPublisher, error) {
	natsURL := cfg.NatsURL
	topic := cfg.TransactionTopic

	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	if topic == "" {
		topic = "transaction-service.transaction.created"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &natsEventPublisher{nc: nc, topic: topic}, nil
}

func (p *natsEventPublisher) PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error {
	event := TransactionCreatedEvent{
		TransactionID: transaction.ID,
		UserID:        transaction.UserID,
		AssetSymbol:   transaction.Symbol,
		PricePerShare: transaction.PricePerShare,
		Quantity:      transaction.Quantity,
		Type:          transaction.Type,
		ExecutedAt:    transaction.ExecutedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(p.topic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
