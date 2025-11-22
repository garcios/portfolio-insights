package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Transaction Business Metrics
	TransactionsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_created_total",
			Help: "Total number of transactions created",
		},
		[]string{"type"}, // BUY or SELL
	)

	TransactionValueTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_value_total",
			Help: "Total transaction value in USD",
		},
		[]string{"type"},
	)

	TransactionProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "transaction_processing_duration_seconds",
			Help:    "Time spent processing transaction business logic",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	// External Service Call Metrics
	UserValidationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_validation_duration_seconds",
			Help:    "Time spent validating user via gRPC",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)

	AssetValidationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "asset_validation_duration_seconds",
			Help:    "Time spent validating asset via gRPC",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)

	// NATS Publishing Metrics
	NATSPublishTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nats_publish_total",
			Help: "Total number of NATS messages published",
		},
		[]string{"subject", "status"}, // status: success, failed
	)

	NATSPublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nats_publish_duration_seconds",
			Help:    "Time spent publishing NATS messages",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"subject"},
	)
)
