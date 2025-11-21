package grpc

import (
	"context"

	"github.com/garcios/portfolio-insights/services/user-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
)

type UserHandler struct {
	uc *usecase.UserUsecase
	pb.UnimplementedUserServiceServer
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	u, err := h.uc.GetUser(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{Id: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// TODO: Implement CreateUser in usecase and call it here
	return &pb.CreateUserResponse{Id: "not-implemented"}, nil
}
