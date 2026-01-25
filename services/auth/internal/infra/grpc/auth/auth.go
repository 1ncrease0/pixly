package auth

import (
	"context"
	"errors"
	authv1 "github.com/1ncrease0/pixly/proto/gen/auth"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
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
	var violations []*errdetails.BadRequest_FieldViolation

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "email",
			Description: "invalid email format",
		})
	}
	password, err := domain.NewPassword(req.Password)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "password",
			Description: "invalid password",
		})
	}
	username, err := domain.NewUsername(req.Username)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "username",
			Description: "invalid username",
		})
	}

	if len(violations) > 0 {
		st := status.New(codes.InvalidArgument, "validation failed")
		br := &errdetails.BadRequest{
			FieldViolations: violations,
		}
		st, _ = st.WithDetails(br)
		return nil, st.Err()
	}

	if err := s.service.Register(ctx, email, username, password); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			st := status.New(codes.AlreadyExists, "user already exists")
			br := &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{Field: "email", Description: "user with this email already exists"},
				},
			}
			st, _ = st.WithDetails(br)
			return nil, st.Err()
		}
		if errors.Is(err, domain.ErrUsernameTaken) {
			st := status.New(codes.AlreadyExists, "username taken")
			br := &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{Field: "username", Description: "username is already taken"},
				},
			}
			st, _ = st.WithDetails(br)
			return nil, st.Err()
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
