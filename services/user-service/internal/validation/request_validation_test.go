package validation

import (
	"testing"

	pb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateCreateUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.CreateUserRequest
		wantErr bool
		errCode codes.Code
		errMsg  string
	}{
		{
			name: "valid request",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "test@example.com",
					Username: "testuser",
					Password: "password123",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil user",
			req:     &pb.CreateUserRequest{User: nil},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "user is required",
		},
		{
			name: "empty email",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "",
					Username: "testuser",
					Password: "password123",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "not-an-email",
					Username: "testuser",
					Password: "password123",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "invalid email format",
		},
		{
			name: "empty username",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "test@example.com",
					Username: "",
					Password: "password123",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "username is required",
		},
		{
			name: "empty password",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "test@example.com",
					Username: "testuser",
					Password: "",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateUserRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateUserRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error, got: %v", err)
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
				}
				if st.Message() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, st.Message())
				}
			}
		})
	}
}

func TestValidateGetUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.GetUserRequest
		wantErr bool
		errCode codes.Code
		errMsg  string
	}{
		{
			name: "valid request",
			req: &pb.GetUserRequest{
				Name: "users/123e4567-e89b-12d3-a456-426614174000",
			},
			wantErr: false,
		},
		{
			name:    "empty resource name",
			req:     &pb.GetUserRequest{Name: ""},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "resource name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetUserRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGetUserRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error, got: %v", err)
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
				}
				if st.Message() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, st.Message())
				}
			}
		})
	}
}

func TestValidateVerifyUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.VerifyUserRequest
		wantErr bool
		errCode codes.Code
		errMsg  string
	}{
		{
			name: "valid request",
			req: &pb.VerifyUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			req: &pb.VerifyUserRequest{
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "email is required",
		},
		{
			name: "empty password",
			req: &pb.VerifyUserRequest{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
			errMsg:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVerifyUserRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVerifyUserRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error, got: %v", err)
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
				}
				if st.Message() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, st.Message())
				}
			}
		})
	}
}
