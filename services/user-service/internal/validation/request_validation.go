// Package validation provides request validation functions for the user service.
package validation

import (
	"regexp"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	pb "github.com/garcios/portfolio-insights/services/user-service/user"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// emailRegex is a simple regex for basic email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidateCreateUserRequest validates the CreateUser request.
func ValidateCreateUserRequest(req *pb.CreateUserRequest) error {
	if req.User == nil {
		return status.Error(codes.InvalidArgument, "user is required")
	}

	if req.User.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if !emailRegex.MatchString(req.User.Email) {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	if req.User.Username == "" {
		return status.Error(codes.InvalidArgument, "username is required")
	}

	if req.User.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

// ValidateGetUserRequest validates the GetUser request.
func ValidateGetUserRequest(req *pb.GetUserRequest) error {
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "resource name is required")
	}

	// Parse and validate the resource name format
	userID, err := resourcenames.ParseUserName(req.Name)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	if userID == "" {
		return status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	return nil
}

// ValidateVerifyUserRequest validates the VerifyUser request.
func ValidateVerifyUserRequest(req *pb.VerifyUserRequest) error {
	if req.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}
