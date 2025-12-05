// Package metrics defines prometheus metrics for the application.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal is the total number of HTTP requests.
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_consent_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HttpRequestDuration is the duration of HTTP requests.
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "login_consent_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// GrpcClientRequestsTotal is the total number of gRPC client requests.
	GrpcClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_consent_grpc_client_requests_total",
			Help: "Total number of gRPC client requests",
		},
		[]string{"method", "status"},
	)

	// GrpcClientRequestDuration is the duration of gRPC client requests.
	GrpcClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "login_consent_grpc_client_request_duration_seconds",
			Help:    "Duration of gRPC client requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)
