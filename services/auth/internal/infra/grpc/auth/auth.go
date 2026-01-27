package auth

import (
	"context"
	"errors"
	authv1 "github.com/1ncrease0/pixly/proto/gen/auth"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	VerifyEmail(ctx context.Context, code domain.VerificationCode) error
	Register(ctx context.Context, email domain.Email, name domain.Username, password domain.Password) error
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	ResendVerification(ctx context.Context, email domain.Email) error
	Login(ctx context.Context, email domain.Email, password domain.Password) (*domain.TokenPair, error)
	User(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type Server struct {
	authv1.UnimplementedAuthServer
	service Service
}

func Register(gRPCServer *grpc.Server, s Service) {
	authv1.RegisterAuthServer(gRPCServer, &Server{service: s})
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
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
			Description: "password must be at least 8 characters",
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

	tokens, err := s.service.Login(ctx, email, password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")

		case errors.Is(err, domain.ErrUserNotVerified):
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		case errors.Is(err, domain.ErrInvalidPassword):
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")

		default:
			return nil, status.Error(codes.Internal, "failed to login")
		}
	}

	return &authv1.LoginResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	}, nil
}
func (s *Server) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.service.User(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")

		case errors.Is(err, domain.ErrUserNotVerified):
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		default:
			return nil, status.Error(codes.Internal, "failed to get user")
		}
	}

	return &authv1.GetUserResponse{
		Id:       user.ID().String(),
		Email:    user.Email().String(),
		Username: user.Name().String(),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokens, err := s.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")

		case errors.Is(err, domain.ErrSessionExpired):
			return nil, status.Error(codes.Unauthenticated, "refresh token expired")

		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")

		case errors.Is(err, domain.ErrUserNotVerified):
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		default:
			return nil, status.Error(codes.Internal, "failed to refresh token")
		}
	}

	return &authv1.RefreshResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	}, nil
}

func (s *Server) ResendVerification(ctx context.Context, req *authv1.ResendVerificationRequest) (*authv1.ResendVerificationResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		st := status.New(codes.InvalidArgument, "validation failed")
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "email", Description: "invalid email format"},
			},
		}
		st, _ = st.WithDetails(br)
		return nil, st.Err()
	}

	if err := s.service.ResendVerification(ctx, email); err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, domain.ErrUserAlreadyVerified):
			return nil, status.Error(codes.FailedPrecondition, "user is already verified")
		default:
			return nil, status.Error(codes.Internal, "failed to resend verification")
		}
	}

	return &authv1.ResendVerificationResponse{
		Success: true,
	}, nil
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
	}, nil
}
