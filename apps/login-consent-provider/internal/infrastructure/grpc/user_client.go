package grpc

import (
	"context"
	"fmt"
	"log"

	"time"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/domain"
	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/metrics"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// UserServiceClient wraps the gRPC client for user-service
type UserServiceClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserServiceClient creates a new user service client
func NewUserServiceClient(userServiceAddr string) (*UserServiceClient, error) {
	conn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := pb.NewUserServiceClient(conn)
	log.Printf("Connected to user-service at %s", userServiceAddr)

	return &UserServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// VerifyUser verifies user credentials via gRPC call
func (c *UserServiceClient) VerifyUser(ctx context.Context, email, password string) (*domain.User, error) {
	start := time.Now()
	var err error
	defer func() {
		recordMetrics("VerifyUser", start, err)
	}()

	resp, err := c.client.VerifyUser(ctx, &pb.VerifyUserRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if !resp.Valid {
		err = fmt.Errorf("invalid credentials")
		return nil, err
	}

	return &domain.User{
		ID:       resp.Id,
		Email:    resp.Email,
		Username: resp.Username,
	}, nil
}

// GetUser retrieves user by ID via gRPC call
func (c *UserServiceClient) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	start := time.Now()
	var err error
	defer func() {
		recordMetrics("GetUser", start, err)
	}()

	resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &domain.User{
		ID:       resp.Id,
		Email:    resp.Email,
		Username: resp.Username,
	}, nil
}

func recordMetrics(method string, start time.Time, err error) {
	duration := time.Since(start).Seconds()
	statusCode := codes.OK.String()
	if err != nil {
		if s, ok := status.FromError(err); ok {
			statusCode = s.Code().String()
		} else {
			statusCode = "ERROR"
		}
	}
	metrics.GrpcClientRequestsTotal.WithLabelValues(method, statusCode).Inc()
	metrics.GrpcClientRequestDuration.WithLabelValues(method).Observe(duration)
}
