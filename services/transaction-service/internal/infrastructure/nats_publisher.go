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
	Type          string    `json:"type"`
	AssetSymbol   *string   `json:"asset_symbol,omitempty"`
	PricePerShare *float64  `json:"price_per_share,omitempty"`
	Quantity      *float64  `json:"quantity,omitempty"`
	Amount        *float64  `json:"amount,omitempty"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// TransactionUpdatedEvent represents the event payload for an updated transaction.
type TransactionUpdatedEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	Type          string    `json:"type"`
	AssetSymbol   *string   `json:"asset_symbol,omitempty"`
	PricePerShare *float64  `json:"price_per_share,omitempty"`
	Quantity      *float64  `json:"quantity,omitempty"`
	Amount        *float64  `json:"amount,omitempty"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// TransactionDeletedEvent represents the event payload for a deleted transaction.
type TransactionDeletedEvent struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	Type          string  `json:"type"`
	AssetSymbol   *string `json:"asset_symbol,omitempty"`
}

type natsEventPublisher struct {
	nc           *nats.Conn
	createdTopic string
	updatedTopic string
	deletedTopic string
}

// NewNATSEventPublisher creates a new NATS event publisher.
func NewNATSEventPublisher(cfg config.Config) (domain.EventPublisher, error) {
	natsURL := cfg.NatsURL
	createdTopic := cfg.TransactionTopic
	updatedTopic := cfg.TransactionUpdatedTopic
	deletedTopic := cfg.TransactionDeletedTopic

	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	if createdTopic == "" {
		createdTopic = "transaction-service.transaction.created"
	}
	if updatedTopic == "" {
		updatedTopic = "transaction-service.transaction.updated"
	}
	if deletedTopic == "" {
		deletedTopic = "transaction-service.transaction.deleted"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &natsEventPublisher{
		nc:           nc,
		createdTopic: createdTopic,
		updatedTopic: updatedTopic,
		deletedTopic: deletedTopic,
	}, nil
}

func (p *natsEventPublisher) PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error {
	event := TransactionCreatedEvent{
		TransactionID: transaction.ID,
		UserID:        transaction.UserID,
		Type:          transaction.Type,
		AssetSymbol:   transaction.Symbol,
		PricePerShare: transaction.PricePerShare,
		Quantity:      transaction.Quantity,
		Amount:        transaction.Amount,
		ExecutedAt:    transaction.ExecutedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(p.createdTopic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (p *natsEventPublisher) PublishTransactionUpdated(ctx context.Context, transaction *domain.Transaction) error {
	event := TransactionUpdatedEvent{
		TransactionID: transaction.ID,
		UserID:        transaction.UserID,
		Type:          transaction.Type,
		AssetSymbol:   transaction.Symbol,
		PricePerShare: transaction.PricePerShare,
		Quantity:      transaction.Quantity,
		Amount:        transaction.Amount,
		ExecutedAt:    transaction.ExecutedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(p.updatedTopic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (p *natsEventPublisher) PublishTransactionDeleted(ctx context.Context, transaction *domain.Transaction) error {
	event := TransactionDeletedEvent{
		TransactionID: transaction.ID,
		UserID:        transaction.UserID,
		Type:          transaction.Type,
		AssetSymbol:   transaction.Symbol,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(p.deletedTopic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
