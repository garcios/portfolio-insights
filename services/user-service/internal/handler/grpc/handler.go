// Package grpc implements the gRPC handlers for the user service.
package grpc

import (
	"context"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler implements the gRPC user service.
type UserHandler struct {
	uc domain.UserUsecase
	pb.UnimplementedUserServiceServer
}

// NewUserHandler creates a new user handler.
func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// GetUser handles the GetUser gRPC request.
// AIP-131 compliant: uses resource name instead of ID.
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// Parse resource name to extract user ID
	userID, err := resourcenames.ParseUserName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	// Get user from usecase
	u, err := h.uc.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Return user with resource name
	return &pb.User{
		Name:     resourcenames.UserName(u.ID),
		Email:    u.Email,
		Username: u.Username,
		UserId:   u.ID,
	}, nil
}

// CreateUser handles the CreateUser gRPC request.
// AIP-133 compliant: accepts User object instead of individual fields.
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// Validate request
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	// Use client-specified user ID if provided, otherwise let usecase generate one
	var userID string
	if req.UserId != "" {
		userID = req.UserId
	}

	// Create user via usecase
	user, err := h.uc.CreateUser(req.User.Email, req.User.Username, req.User.Password)
	if err != nil {
		return nil, err
	}

	// If client specified an ID, we would need to handle that in the usecase
	// For now, we use the generated ID
	_ = userID // TODO: Support client-specified IDs in usecase

	// Return created user with resource name
	return &pb.User{
		Name:     resourcenames.UserName(user.ID),
		Email:    user.Email,
		Username: user.Username,
		UserId:   user.ID,
	}, nil
}

// VerifyUser handles the VerifyUser gRPC request.
// This is a custom method for authentication.
func (h *UserHandler) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.VerifyUserResponse, error) {
	// Verify user credentials
	user, err := h.uc.VerifyUser(req.Email, req.Password)
	if err != nil {
		// Return invalid response without error for failed authentication
		return &pb.VerifyUserResponse{Valid: false}, nil
	}

	// Return valid response with user information
	return &pb.VerifyUserResponse{
		Valid: true,
		User: &pb.User{
			Name:     resourcenames.UserName(user.ID),
			Email:    user.Email,
			Username: user.Username,
			UserId:   user.ID,
		},
	}, nil
}
