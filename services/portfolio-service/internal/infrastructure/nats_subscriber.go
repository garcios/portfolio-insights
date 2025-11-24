package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/nats-io/nats.go"
)

const (
	TransactionCreatedSubject = "transaction-service.transaction.created"
)

type TransactionCreatedEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	AssetSymbol   string    `json:"asset_symbol"`
	PricePerShare float64   `json:"price_per_share"`
	Quantity      float64   `json:"quantity"`
	Type          string    `json:"type"`
	ExecutedAt    time.Time `json:"executed_at"`
}

type NATSSubscriber struct {
	nc     *nats.Conn
	sub    *nats.Subscription
	repo   domain.HoldingRepository
	logger *slog.Logger
}

func NewNATSSubscriber(repo domain.HoldingRepository, l *slog.Logger) (*NATSSubscriber, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	subscriber := &NATSSubscriber{
		nc:     nc,
		repo:   repo,
		logger: l,
	}

	return subscriber, nil
}

func (s *NATSSubscriber) Start() error {
	var err error
	s.sub, err = s.nc.Subscribe(TransactionCreatedSubject, s.handleTransactionCreated)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", TransactionCreatedSubject, err)
	}

	s.logger.Info("Subscribed to NATS topic", "topic", TransactionCreatedSubject)
	return nil
}

func (s *NATSSubscriber) Stop() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
	if s.nc != nil {
		s.nc.Close()
	}
}

func (s *NATSSubscriber) handleTransactionCreated(msg *nats.Msg) {
	start := time.Now()

	var event TransactionCreatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("failed to unmarshal transaction created event", "error", err)
		metrics.RecordNatsMessage(TransactionCreatedSubject, "unmarshal_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Received transaction created event",
		"transaction_id", event.TransactionID,
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"type", event.Type,
		"quantity", event.Quantity,
	)

	// Get existing holding
	holding, err := s.repo.GetByUserAndSymbol(event.UserID, event.AssetSymbol)
	if err != nil {
		// If holding doesn't exist, create a new one
		holding = &domain.Holding{
			UserID:      event.UserID,
			Symbol:      event.AssetSymbol,
			Quantity:    0,
			AverageCost: 0,
		}
	}

	// Update holding based on transaction type
	switch event.Type {
	case "BUY":
		// Calculate new average cost
		totalCost := (holding.Quantity * holding.AverageCost) + (event.Quantity * event.PricePerShare)
		newQuantity := holding.Quantity + event.Quantity
		holding.AverageCost = totalCost / newQuantity
		holding.Quantity = newQuantity
	case "SELL":
		// Reduce quantity, keep average cost the same
		holding.Quantity -= event.Quantity
		// If quantity goes to zero or negative, we might want to delete the holding
		if holding.Quantity <= 0 {
			holding.Quantity = 0
		}
	default:
		s.logger.Warn("unknown transaction type", "type", event.Type, "transaction_id", event.TransactionID)
		metrics.RecordNatsMessage(TransactionCreatedSubject, "unknown_type", time.Since(start).Seconds())
		return
	}

	holding.LastUpdated = time.Now()

	// Save updated holding
	if err := s.repo.Upsert(holding); err != nil {
		s.logger.Error("failed to update holding", "error", err, "user_id", event.UserID, "symbol", event.AssetSymbol)
		metrics.RecordNatsMessage(TransactionCreatedSubject, "db_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Updated portfolio holding",
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"new_quantity", holding.Quantity,
		"average_cost", holding.AverageCost,
	)

	metrics.RecordNatsMessage(TransactionCreatedSubject, "success", time.Since(start).Seconds())
}
