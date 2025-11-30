package grpc

import (
	"context"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
)

type UserHandler struct {
	uc domain.UserUsecase
	pb.UnimplementedUserServiceServer
}

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	u, err := h.uc.GetUser(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{Id: u.ID, Username: u.Username, Email: u.Email}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := h.uc.CreateUser(req.Email, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Id: user.ID}, nil
}
