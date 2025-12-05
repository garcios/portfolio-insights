// Package grpc implements the gRPC handlers for the user service.
package grpc

import (
	"context"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
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
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	u, err := h.uc.GetUser(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{Id: u.ID, Username: u.Username, Email: u.Email}, nil
}

// CreateUser handles the CreateUser gRPC request.
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := h.uc.CreateUser(req.Email, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Id: user.ID}, nil
}

// VerifyUser handles the VerifyUser gRPC request.
func (h *UserHandler) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.VerifyUserResponse, error) {
	user, err := h.uc.VerifyUser(req.Email, req.Password)
	if err != nil {
		return &pb.VerifyUserResponse{Valid: false}, nil
	}
	return &pb.VerifyUserResponse{
		Valid:    true,
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
