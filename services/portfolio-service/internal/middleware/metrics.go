package middleware

import (
	"context"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsUnaryInterceptor is a gRPC unary interceptor that records metrics
func MetricsUnaryInterceptor() grpc.UnaryServerInterceptor {
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
		statusCode := "OK"
		if err != nil {
			statusCode = status.Code(err).String()
		}

		metrics.RecordGrpcRequest(info.FullMethod, statusCode, duration)

		return resp, err
	}
}
