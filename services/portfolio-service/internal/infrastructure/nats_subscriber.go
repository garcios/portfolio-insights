package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"log/slog"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/nats-io/nats.go"
)

const (
	// TransactionCreatedSubject is the NATS subject for transaction created events.
	TransactionCreatedSubject = "transaction-service.transaction.created"
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

// NATSSubscriber subscribes to NATS events.
type NATSSubscriber struct {
	nc                *nats.Conn
	sub               *nats.Subscription
	repo              domain.HoldingRepository
	marketDataGateway *MarketDataGateway
	assetCache        *AssetCache
	logger            *slog.Logger
}

// NewNATSSubscriber creates a new NATS subscriber.
func NewNATSSubscriber(repo domain.HoldingRepository, marketDataGateway *MarketDataGateway, assetCache *AssetCache, l *slog.Logger, cfg config.Config) (*NATSSubscriber, error) {
	natsURL := cfg.NatsURL
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	subscriber := &NATSSubscriber{
		nc:                nc,
		repo:              repo,
		marketDataGateway: marketDataGateway,
		assetCache:        assetCache,
		logger:            l,
	}

	return subscriber, nil
}

// Start starts subscribing to the NATS subject.
func (s *NATSSubscriber) Start() error {
	var err error
	s.sub, err = s.nc.Subscribe(TransactionCreatedSubject, s.handleTransactionCreated)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", TransactionCreatedSubject, err)
	}

	s.logger.Info("Subscribed to NATS topic", "topic", TransactionCreatedSubject)
	return nil
}

// Stop stops the NATS subscriber.
func (s *NATSSubscriber) Stop() {
	if s.sub != nil {
		if err := s.sub.Unsubscribe(); err != nil {
			s.logger.Error("failed to unsubscribe from NATS", "error", err)
		}
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
		// Fetch currency from cache or marketdata service
		currency := "USD" // Default currency
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try cache first
		var cachedAsset *CachedAsset
		if s.assetCache != nil {
			cachedAsset, err = s.assetCache.Get(ctx, event.AssetSymbol)
			if err != nil {
				s.logger.Warn("failed to get asset from cache", "error", err, "symbol", event.AssetSymbol)
			}
		}

		if cachedAsset != nil {
			// Cache hit
			currency = cachedAsset.Currency
			s.logger.Debug("asset currency from cache", "symbol", event.AssetSymbol, "currency", currency)
		} else if s.marketDataGateway != nil {
			// Cache miss - fetch from marketdata service
			assetResp, err := s.marketDataGateway.client.GetAsset(ctx, &pb.GetAssetRequest{
				Symbol: event.AssetSymbol,
			})
			if err != nil {
				s.logger.Warn("failed to fetch asset from marketdata service, using default USD", "error", err, "symbol", event.AssetSymbol)
			} else if assetResp.Asset != nil {
				currency = assetResp.Asset.Currency
				s.logger.Debug("asset currency from marketdata service", "symbol", event.AssetSymbol, "currency", currency)

				// Cache the asset for future use
				if s.assetCache != nil {
					if err := s.assetCache.Set(ctx, assetResp.Asset); err != nil {
						s.logger.Warn("failed to cache asset", "error", err, "symbol", event.AssetSymbol)
					}
				}
			}
		}

		holding = &domain.Holding{
			UserID:      event.UserID,
			Symbol:      event.AssetSymbol,
			Quantity:    0,
			AverageCost: 0,
			Currency:    currency,
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
	case "SPLIT":
		// Stock split: increase quantity, decrease average cost proportionally
		// Total cost basis remains the same
		// Example: 2-for-1 split of 100 shares @ $100 = 200 shares @ $50
		if holding.Quantity > 0 && event.Quantity > 0 {
			// Calculate split ratio from the additional quantity
			// If we had 100 shares and receive 100 more, it's a 2-for-1 split (ratio = 2.0)
			splitRatio := (holding.Quantity + event.Quantity) / holding.Quantity

			// Adjust average cost by the inverse of the split ratio
			holding.AverageCost = holding.AverageCost / splitRatio

			// Add the new shares
			holding.Quantity += event.Quantity

			s.logger.Info("Processed stock split",
				"symbol", event.AssetSymbol,
				"split_ratio", splitRatio,
				"new_quantity", holding.Quantity,
				"new_average_cost", holding.AverageCost,
			)
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
		"currency", holding.Currency,
	)

	metrics.RecordNatsMessage(TransactionCreatedSubject, "success", time.Since(start).Seconds())
}
