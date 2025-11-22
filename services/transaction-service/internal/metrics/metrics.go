package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "handler", "method", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"service", "handler", "method", "status_code"},
	)

	HTTPInFlightRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests",
		},
		[]string{"service", "handler"},
	)

	// gRPC Metrics
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)
)
