// Package grpc implements the gRPC handlers for the user service.
package grpc

import (
	"context"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/user-service/internal/validation"
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
	// Validate request
	if err := validation.ValidateGetUserRequest(req); err != nil {
		return nil, err
	}

	// Parse resource name to extract user ID
	userID, _ := resourcenames.ParseUserName(req.Name)

	// Get user from usecase
	u, err := h.uc.GetUser(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", req.Name)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Return user with resource name
	return &pb.User{
		Name:      resourcenames.UserName(u.ID),
		Email:     u.Email,
		Username:  u.Username,
		UserId:    u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
	}, nil
}

// CreateUser handles the CreateUser gRPC request.
// AIP-133 compliant: accepts User object instead of individual fields.
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// Validate request
	if err := validation.ValidateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Use client-specified user ID if provided, otherwise let usecase generate one
	var userID string
	if req.UserId != "" {
		userID = req.UserId
	}

	// Create user via usecase
	user := &domain.User{
		Email:       req.User.Email,
		Username:    req.User.Username,
		FirstName:   req.User.FirstName,
		LastName:    req.User.LastName,
		Role:        req.User.Role,
		Preferences: req.User.Preferences.AsMap(),
	}

	if req.User.LastLoginAt != nil {
		user.LastLoginAt = req.User.LastLoginAt.AsTime()
	}

	createdUser, err := h.uc.CreateUser(user, req.User.Password)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// If client specified an ID, we would need to handle that in the usecase
	// For now, we use the generated ID
	_ = userID // TODO: Support client-specified IDs in usecase

	// Return created user with resource name
	return &pb.User{
		Name:      resourcenames.UserName(createdUser.ID),
		Email:     createdUser.Email,
		Username:  createdUser.Username,
		UserId:    createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Role:      createdUser.Role,
		// TODO: Convert Map to Struct for Preferences
	}, nil
}

// VerifyUser handles the VerifyUser gRPC request.
// This is a custom method for authentication.
func (h *UserHandler) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.VerifyUserResponse, error) {
	// Validate request
	if err := validation.ValidateVerifyUserRequest(req); err != nil {
		return nil, err
	}

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
			Name:      resourcenames.UserName(user.ID),
			Email:     user.Email,
			Username:  user.Username,
			UserId:    user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			// TODO: Convert Map to Struct for Preferences
		},
	}, nil
}
