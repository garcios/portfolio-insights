package middleware

import (
	"context"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor for metrics
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := status.Code(err).String()

		metrics.GRPCRequestsTotal.WithLabelValues(
			serviceName,
			info.FullMethod,
			statusCode,
		).Inc()

		metrics.GRPCRequestDuration.WithLabelValues(
			serviceName,
			info.FullMethod,
			statusCode,
		).Observe(duration)

		return resp, err
	}
}
