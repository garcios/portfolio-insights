// Package middleware provides metrics middleware for the portfolio service.
package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsUnaryInterceptor is a gRPC unary interceptor that records metrics
func MetricsUnaryInterceptor(recordFunc func(method, status string, duration float64)) grpc.UnaryServerInterceptor {
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

		if recordFunc != nil {
			recordFunc(info.FullMethod, statusCode, duration)
		}

		return resp, err
	}
}
