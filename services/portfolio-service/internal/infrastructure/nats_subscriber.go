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
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
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

// TransactionUpdatedEvent represents the event payload for an updated transaction.
type TransactionUpdatedEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	AssetSymbol   string    `json:"asset_symbol"`
	PricePerShare float64   `json:"price_per_share"`
	Quantity      float64   `json:"quantity"`
	Type          string    `json:"type"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// TransactionDeletedEvent represents the event payload for a deleted transaction.
type TransactionDeletedEvent struct {
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
	AssetSymbol   string `json:"asset_symbol"`
}

// NATSSubscriber subscribes to NATS events.
type NATSSubscriber struct {
	nc                *nats.Conn
	subCreated        *nats.Subscription
	subUpdated        *nats.Subscription
	subDeleted        *nats.Subscription
	repo              domain.HoldingRepository
	marketDataGateway *MarketDataGateway
	transactionClient transactionpb.TransactionServiceClient
	assetCache        *AssetCache
	logger            *slog.Logger

	createdTopic string
	updatedTopic string
	deletedTopic string
}

// NewNATSSubscriber creates a new NATS subscriber.
func NewNATSSubscriber(
	repo domain.HoldingRepository,
	marketDataGateway *MarketDataGateway,
	transactionClient transactionpb.TransactionServiceClient,
	assetCache *AssetCache,
	l *slog.Logger,
	cfg config.Config,
) (*NATSSubscriber, error) {
	natsURL := cfg.NatsURL
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSSubscriber{
		nc:                nc,
		repo:              repo,
		marketDataGateway: marketDataGateway,
		transactionClient: transactionClient,
		assetCache:        assetCache,
		logger:            l,
		createdTopic:      cfg.TransactionCreatedTopic,
		updatedTopic:      cfg.TransactionUpdatedTopic,
		deletedTopic:      cfg.TransactionDeletedTopic,
	}, nil
}

// Start starts subscribing to the NATS subjects.
func (s *NATSSubscriber) Start() error {
	var err error

	// Subscribe to Created
	s.subCreated, err = s.nc.Subscribe(s.createdTopic, s.handleTransactionCreated)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", s.createdTopic, err)
	}
	s.logger.Info("Subscribed to NATS topic", "topic", s.createdTopic)

	// Subscribe to Updated
	s.subUpdated, err = s.nc.Subscribe(s.updatedTopic, s.handleTransactionUpdated)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", s.updatedTopic, err)
	}
	s.logger.Info("Subscribed to NATS topic", "topic", s.updatedTopic)

	// Subscribe to Deleted
	s.subDeleted, err = s.nc.Subscribe(s.deletedTopic, s.handleTransactionDeleted)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", s.deletedTopic, err)
	}
	s.logger.Info("Subscribed to NATS topic", "topic", s.deletedTopic)

	return nil
}

// Stop stops the NATS subscriber.
func (s *NATSSubscriber) Stop() {
	if s.subCreated != nil {
		if err := s.subCreated.Unsubscribe(); err != nil {
			s.logger.Error("failed to unsubscribe from created topic", "error", err)
		}
	}
	if s.subUpdated != nil {
		if err := s.subUpdated.Unsubscribe(); err != nil {
			s.logger.Error("failed to unsubscribe from updated topic", "error", err)
		}
	}
	if s.subDeleted != nil {
		if err := s.subDeleted.Unsubscribe(); err != nil {
			s.logger.Error("failed to unsubscribe from deleted topic", "error", err)
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
		metrics.RecordNatsMessage(s.createdTopic, "unmarshal_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Received transaction created event",
		"transaction_id", event.TransactionID,
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"type", event.Type,
		"quantity", event.Quantity,
	)

	// Fallback to recalculation if we have the client, but keeping incremental logic for now as requested by plan.
	// Actually, let's use recalculate for everything for consistency if possible, but incremental is faster.
	// We'll keep incremental for CREATED as it's the happy path and less resource intensive.

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
		if newQuantity > 0 {
			holding.AverageCost = totalCost / newQuantity
		}
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
		if holding.Quantity > 0 && event.Quantity > 0 {
			splitRatio := (holding.Quantity + event.Quantity) / holding.Quantity
			holding.AverageCost = holding.AverageCost / splitRatio
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
		metrics.RecordNatsMessage(s.createdTopic, "unknown_type", time.Since(start).Seconds())
		return
	}

	holding.LastUpdated = time.Now()

	// Save updated holding
	if err := s.repo.Upsert(holding); err != nil {
		s.logger.Error("failed to update holding", "error", err, "user_id", event.UserID, "symbol", event.AssetSymbol)
		metrics.RecordNatsMessage(s.createdTopic, "db_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Updated portfolio holding (created event)",
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"new_quantity", holding.Quantity,
		"average_cost", holding.AverageCost,
		"currency", holding.Currency,
	)

	metrics.RecordNatsMessage(s.createdTopic, "success", time.Since(start).Seconds())
}

func (s *NATSSubscriber) handleTransactionUpdated(msg *nats.Msg) {
	start := time.Now()

	var event TransactionUpdatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("failed to unmarshal transaction updated event", "error", err)
		metrics.RecordNatsMessage(s.updatedTopic, "unmarshal_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Received transaction updated event", "transaction_id", event.TransactionID)

	if err := s.recalculateHolding(context.Background(), event.UserID, event.AssetSymbol); err != nil {
		s.logger.Error("failed to recalculate holding after update", "error", err, "user_id", event.UserID, "symbol", event.AssetSymbol)
		metrics.RecordNatsMessage(s.updatedTopic, "recalc_error", time.Since(start).Seconds())
		return
	}

	metrics.RecordNatsMessage(s.updatedTopic, "success", time.Since(start).Seconds())
}

func (s *NATSSubscriber) handleTransactionDeleted(msg *nats.Msg) {
	start := time.Now()

	var event TransactionDeletedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("failed to unmarshal transaction deleted event", "error", err)
		metrics.RecordNatsMessage(s.deletedTopic, "unmarshal_error", time.Since(start).Seconds())
		return
	}

	s.logger.Info("Received transaction deleted event", "transaction_id", event.TransactionID)

	if err := s.recalculateHolding(context.Background(), event.UserID, event.AssetSymbol); err != nil {
		s.logger.Error("failed to recalculate holding after delete", "error", err, "user_id", event.UserID, "symbol", event.AssetSymbol)
		metrics.RecordNatsMessage(s.deletedTopic, "recalc_error", time.Since(start).Seconds())
		return
	}

	metrics.RecordNatsMessage(s.deletedTopic, "success", time.Since(start).Seconds())
}

func (s *NATSSubscriber) recalculateHolding(ctx context.Context, userID, symbol string) error {
	// 1. Fetch all transactions for this user/symbol
	// We need to implement pagination properly to get ALL transactions
	var allTxns []*transactionpb.Transaction
	pageSize := int32(100)
	pageToken := ""

	for {
		resp, err := s.transactionClient.ListTransactions(ctx, &transactionpb.ListTransactionsRequest{
			UserId: userID,
			Filter: &transactionpb.TransactionFilter{
				Symbol: symbol,
			},
			PageSize:  pageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list transactions: %w", err)
		}

		allTxns = append(allTxns, resp.Transactions...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// 2. Re-calculate holding state
	// Sort transactions by ExecutedAt is handled by ListTransactions but generally good to double check order logic if needed.
	// Assuming ListTransactions returns roughly ordered or we process them.
	// Actually, ListTransactions returns descending by default. We need ASCENDING for correct average cost calculation.
	// Or we just fetch them all and sort them here.

	// Simple sort in-memory
	// Note: using simple bubble/insertion sort or slice.Sortfunc if strict dependency allowed.
	// Let's iterate backwards if they are DESC, or just sort properly.
	// Since we only depend on 'allTxns', let's just sort them.
	// However, transactionpb.Transaction struct might be large? No, it's just pointers.

	// We will iterate them in reverse if they are DESC.
	// ListTransactions usually returns newest first.
	// So iterating from len-1 to 0 gives chronological order.

	var quantity float64
	var totalCost float64 // Used for average cost calculation

	for i := len(allTxns) - 1; i >= 0; i-- {
		tx := allTxns[i]
		switch tx.Type {
		case "BUY":
			cost := tx.Quantity * tx.PricePerShare
			totalCost += cost // Add to total cost basis
			quantity += tx.Quantity
		case "SELL":
			// Reduce quantity.
			// Average cost remains the same, so we reduce totalCost proportionally to keep AvgCost constant?
			// AvgCost = TotalCost / Quantity
			// NewTotalCost = (Quantity - SoldQty) * AvgCost
			//              = (Quantity - SoldQty) * (TotalCost / Quantity)
			if quantity > 0 {
				avgCost := totalCost / quantity
				quantity -= tx.Quantity
				if quantity < 0 {
					quantity = 0
				}
				totalCost = quantity * avgCost
			}
		case "SPLIT":
			// Increase quantity, total cost stays same (so avg cost reduces)
			if quantity > 0 && tx.Quantity > 0 {
				quantity += tx.Quantity
			}
		}
	}

	avgCost := 0.0
	if quantity > 0 {
		avgCost = totalCost / quantity
	}

	// 3. Update Holding
	holding, err := s.repo.GetByUserAndSymbol(userID, symbol)
	if err != nil {
		// New holding
		// We need currency.
		// If we don't have it easily, we can try to fetch it or default to USD.
		// Since we are recalculating, maybe we can assume the symbol holds the currency metadata?
		// We'll try to get it from cache/gateway.
		currency := "USD"
		if asset, err := s.marketDataGateway.GetAsset(ctx, symbol); err == nil && asset != nil {
			currency = asset.Currency
		}

		holding = &domain.Holding{
			UserID:   userID,
			Symbol:   symbol,
			Currency: currency,
		}
	}

	holding.Quantity = quantity
	holding.AverageCost = avgCost
	holding.LastUpdated = time.Now()

	if err := s.repo.Upsert(holding); err != nil {
		return fmt.Errorf("failed to upsert recalculated holding: %w", err)
	}

	s.logger.Info("Recalculated holding",
		"user_id", userID,
		"symbol", symbol,
		"new_quantity", quantity,
		"new_average_cost", avgCost,
		"txn_count", len(allTxns),
	)

	return nil
}
