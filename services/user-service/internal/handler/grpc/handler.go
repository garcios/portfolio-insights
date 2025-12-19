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
	"google.golang.org/protobuf/types/known/structpb"
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
	// Convert preferences map to structpb
	preferences, err := structpb.NewStruct(u.Preferences)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert preferences: %v", err)
	}

	// Return user with resource name
	return &pb.User{
		Name:        resourcenames.UserName(u.ID),
		Email:       u.Email,
		Username:    u.Username,
		UserId:      u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Role:        u.Role,
		Preferences: preferences,
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
	// Convert preferences map to structpb
	preferences, err := structpb.NewStruct(createdUser.Preferences)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert preferences: %v", err)
	}

	// Return created user with resource name
	return &pb.User{
		Name:        resourcenames.UserName(createdUser.ID),
		Email:       createdUser.Email,
		Username:    createdUser.Username,
		UserId:      createdUser.ID,
		FirstName:   createdUser.FirstName,
		LastName:    createdUser.LastName,
		Role:        createdUser.Role,
		Preferences: preferences,
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
	// Convert preferences map to structpb
	preferences, err := structpb.NewStruct(user.Preferences)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert preferences: %v", err)
	}

	// Return valid response with user information
	return &pb.VerifyUserResponse{
		Valid: true,
		User: &pb.User{
			Name:        resourcenames.UserName(user.ID),
			Email:       user.Email,
			Username:    user.Username,
			UserId:      user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Role:        user.Role,
			Preferences: preferences,
		},
	}, nil
}

// UpdateUser handles the UpdateUser gRPC request.
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// Validate request
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	// Parse resource name to extract user ID
	userID, err := resourcenames.ParseUserName(req.User.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	// Fetch existing user
	existingUser, err := h.uc.GetUser(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", req.User.Name)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Apply updates based on FieldMask
	if req.UpdateMask != nil && len(req.UpdateMask.Paths) > 0 {
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "email":
				existingUser.Email = req.User.Email
			case "username":
				existingUser.Username = req.User.Username
			case "first_name":
				existingUser.FirstName = req.User.FirstName
			case "last_name":
				existingUser.LastName = req.User.LastName
			case "preferences":
				existingUser.Preferences = req.User.Preferences.AsMap()
			}
		}
	} else {
		// No mask, partial update strategy: update non-empty fields from request
		if req.User.Email != "" {
			existingUser.Email = req.User.Email
		}
		if req.User.Username != "" {
			existingUser.Username = req.User.Username
		}
		if req.User.FirstName != "" {
			existingUser.FirstName = req.User.FirstName
		}
		if req.User.LastName != "" {
			existingUser.LastName = req.User.LastName
		}
		if req.User.Preferences != nil {
			existingUser.Preferences = req.User.Preferences.AsMap()
		}
	}

	updatedUser, err := h.uc.UpdateUser(existingUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Convert preferences map to structpb
	preferences, err := structpb.NewStruct(updatedUser.Preferences)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert preferences: %v", err)
	}

	return &pb.User{
		Name:        resourcenames.UserName(updatedUser.ID),
		Email:       updatedUser.Email,
		Username:    updatedUser.Username,
		UserId:      updatedUser.ID,
		FirstName:   updatedUser.FirstName,
		LastName:    updatedUser.LastName,
		Role:        updatedUser.Role,
		Preferences: preferences,
	}, nil
}
