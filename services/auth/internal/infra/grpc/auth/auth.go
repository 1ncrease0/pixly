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
	"log/slog"
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
	log     *slog.Logger
}

func Register(gRPCServer *grpc.Server, s Service, log *slog.Logger) {
	authv1.RegisterAuthServer(gRPCServer, &Server{service: s, log: log})
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
		s.log.Info("login validation failed", slog.String("email", req.Email), slog.Any("violations", violations))
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
			s.log.Info("login failed: user not found", slog.String("email", req.Email))
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")

		case errors.Is(err, domain.ErrUserNotVerified):
			s.log.Info("login failed: user not verified", slog.String("email", req.Email))
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		case errors.Is(err, domain.ErrInvalidPassword):
			s.log.Info("login failed: invalid password", slog.String("email", req.Email))
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")

		default:
			s.log.Error("login failed: internal error", slog.String("email", req.Email), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to login")
		}
	}

	s.log.Info("login successful", slog.String("email", req.Email))
	return &authv1.LoginResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	}, nil
}
func (s *Server) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	if req.Id == "" {
		s.log.Info("get user failed: empty user id")
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	userID, err := uuid.Parse(req.Id)
	if err != nil {
		s.log.Info("get user failed: invalid user id format", slog.String("id", req.Id), slog.Any("error", err))
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.service.User(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			s.log.Info("get user failed: user not found", slog.String("user_id", req.Id))
			return nil, status.Error(codes.NotFound, "user not found")

		case errors.Is(err, domain.ErrUserNotVerified):
			s.log.Info("get user failed: user not verified", slog.String("user_id", req.Id))
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		default:
			s.log.Error("get user failed: internal error", slog.String("user_id", req.Id), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to get user")
		}
	}

	s.log.Debug("get user successful", slog.String("user_id", req.Id))
	return &authv1.GetUserResponse{
		Id:       user.ID().String(),
		Email:    user.Email().String(),
		Username: user.Name().String(),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.RefreshToken == "" {
		s.log.Info("refresh failed: empty refresh token")
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokens, err := s.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			s.log.Info("refresh failed: session not found")
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")

		case errors.Is(err, domain.ErrSessionExpired):
			s.log.Info("refresh failed: session expired")
			return nil, status.Error(codes.Unauthenticated, "refresh token expired")

		case errors.Is(err, domain.ErrUserNotFound):
			s.log.Info("refresh failed: user not found")
			return nil, status.Error(codes.NotFound, "user not found")

		case errors.Is(err, domain.ErrUserNotVerified):
			s.log.Info("refresh failed: user not verified")
			return nil, status.Error(codes.FailedPrecondition, "user not verified")

		default:
			s.log.Error("refresh failed: internal error", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to refresh token")
		}
	}

	s.log.Info("refresh successful")
	return &authv1.RefreshResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	}, nil
}

func (s *Server) ResendVerification(ctx context.Context, req *authv1.ResendVerificationRequest) (*authv1.ResendVerificationResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		s.log.Info("resend verification failed: invalid email format", slog.String("email", req.Email))
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
			s.log.Info("resend verification failed: user not found", slog.String("email", req.Email))
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, domain.ErrUserAlreadyVerified):
			s.log.Info("resend verification failed: user already verified", slog.String("email", req.Email))
			return nil, status.Error(codes.FailedPrecondition, "user is already verified")
		default:
			s.log.Error("resend verification failed: internal error", slog.String("email", req.Email), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to resend verification")
		}
	}

	s.log.Info("resend verification successful", slog.String("email", req.Email))
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
		s.log.Info("register validation failed", slog.String("email", req.Email), slog.String("username", req.Username), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		br := &errdetails.BadRequest{
			FieldViolations: violations,
		}
		st, _ = st.WithDetails(br)
		return nil, st.Err()
	}

	if err := s.service.Register(ctx, email, username, password); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			s.log.Info("register failed: user already exists", slog.String("email", req.Email))
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
			s.log.Info("register failed: username taken", slog.String("username", req.Username))
			st := status.New(codes.AlreadyExists, "username taken")
			br := &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{Field: "username", Description: "username is already taken"},
				},
			}
			st, _ = st.WithDetails(br)
			return nil, st.Err()
		}
		s.log.Error("register failed: internal error", slog.String("email", req.Email), slog.String("username", req.Username), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to register user")
	}
	s.log.Info("register successful", slog.String("email", req.Email), slog.String("username", req.Username))
	return &authv1.RegisterResponse{
		Success: true,
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	code := domain.NewVerificationCodeFromString(req.Code)
	if err := s.service.VerifyEmail(ctx, code); err != nil {
		if errors.Is(err, domain.ErrVerificationCodeNotFound) {
			s.log.Info("verify email failed: invalid verification code")
			return nil, status.Error(codes.InvalidArgument, "invalid verification code")
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			s.log.Info("verify email failed: user not found")
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("verify email failed: internal error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to verify email")
	}
	s.log.Info("verify email successful")
	return &authv1.VerifyEmailResponse{
		Success: true,
	}, nil
}
