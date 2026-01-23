package auth

import (
	"context"
	"errors"
	authv1 "github.com/1ncrease0/pixly/proto/gen/auth"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	VerifyEmail(ctx context.Context, code domain.VerificationCode) error
	Register(ctx context.Context, email domain.Email, name domain.Username, password domain.Password) error
}

type Server struct {
	authv1.UnimplementedAuthServer
	service Service
}

func Register(gRPCServer *grpc.Server, s Service) {
	authv1.RegisterAuthServer(gRPCServer, &Server{service: s})
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	username, err := domain.NewUsername(req.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username")
	}

	if err := s.service.Register(ctx, email, username, password); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		if errors.Is(err, domain.ErrUsernameTaken) {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}
	return &authv1.RegisterResponse{
		Success: true,
		Message: "user registered successfully",
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	code := domain.NewVerificationCodeFromString(req.Code)
	if err := s.service.VerifyEmail(ctx, code); err != nil {
		if errors.Is(err, domain.ErrVerificationCodeNotFound) {
			return nil, status.Error(codes.InvalidArgument, "invalid verification code")
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to verify email")
	}
	return &authv1.VerifyEmailResponse{
		Success: true,
		Message: "email verified successfully",
	}, nil
}
