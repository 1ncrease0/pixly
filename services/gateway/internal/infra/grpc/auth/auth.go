package auth

import (
	"context"
	"fmt"
	authv1 "github.com/1ncrease0/pixly/proto/gen/auth"
	"github.com/1ncrease0/pixly/services/gateway/internal/domain"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type Client struct {
	conn   *grpc.ClientConn
	client authv1.AuthClient
	log    *slog.Logger
}

func NewClient(addr string, timeout time.Duration, retries int, log *slog.Logger) (*Client, error) {
	retryOpts := []retry.CallOption{
		retry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
		retry.WithMax(uint(retries)),
		retry.WithPerRetryTimeout(timeout),
	}

	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			retry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: authv1.NewAuthClient(conn),
		log:    log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Register(ctx context.Context, email, username, password string) error {
	_, err := c.client.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Username: username,
		Password: password,
	})

	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("register failed: unexpected error type", slog.Any("error", err))
		return err
	}

	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok && len(badReq.FieldViolations) > 0 {
			violation := badReq.FieldViolations[0]

			switch st.Code() {
			case codes.InvalidArgument:
				switch violation.Field {
				case "email":
					c.log.Debug("register failed: invalid email format")
					return domain.ErrInvalidEmail
				case "username":
					c.log.Debug("register failed: invalid username format")
					return domain.ErrInvalidUsername
				case "password":
					c.log.Debug("register failed: invalid password format")
					return domain.ErrInvalidPassword
				}

			case codes.AlreadyExists:
				switch violation.Field {
				case "email":
					c.log.Info("register failed: user already exists", slog.String("email", email))
					return domain.ErrUserAlreadyExists
				case "username":
					c.log.Info("register failed: username taken", slog.String("username", username))
					return domain.ErrUsernameTaken
				}
			}
		}
	}
	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("register failed: invalid argument")
		return domain.ErrInvalidArgument
	case codes.AlreadyExists:
		c.log.Info("register failed: already exists")
		return domain.ErrAlreadyExists
	default:
		c.log.Warn("register failed: unexpected status code", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return fmt.Errorf("register failed: %s", st.Message())
	}
}

func (c *Client) VerifyEmail(ctx context.Context, code string) error {
	_, err := c.client.VerifyEmail(ctx, &authv1.VerifyEmailRequest{
		Code: code,
	})
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("verify email failed: unexpected error type", slog.Any("error", err))
		return err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("verify email failed: invalid code")
		return domain.ErrInvalidVerificationCode
	case codes.NotFound:
		c.log.Debug("verify email failed: user not found")
		return domain.ErrUserNotFound
	case codes.Internal:
		c.log.Error("verify email failed: internal error", slog.String("message", st.Message()))
		return fmt.Errorf("verify email failed: %s", st.Message())
	default:
		c.log.Warn("verify email failed: unexpected status code", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return fmt.Errorf("verify email failed: %s", st.Message())
	}
}

func (c *Client) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	resp, err := c.client.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err == nil {
		return &domain.TokenPair{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
		}, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("login failed: unexpected error type", slog.Any("error", err))
		return nil, err
	}

	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok && len(badReq.FieldViolations) > 0 {
			violation := badReq.FieldViolations[0]

			switch st.Code() {
			case codes.InvalidArgument:
				switch violation.Field {
				case "email":
					c.log.Debug("login failed: invalid email format")
					return nil, domain.ErrInvalidEmail
				case "password":
					c.log.Debug("login failed: invalid password format")
					return nil, domain.ErrInvalidPassword
				}
			}
		}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("login failed: invalid argument")
		return nil, domain.ErrInvalidArgument
	case codes.Unauthenticated:
		c.log.Info("login failed: authentication failed")
		return nil, domain.ErrUnauthenticated
	case codes.FailedPrecondition:
		c.log.Info("login failed: user not verified")
		return nil, domain.ErrUserNotVerified
	case codes.NotFound:
		c.log.Info("login failed: user not found")
		return nil, domain.ErrUserNotFound
	case codes.Internal:
		c.log.Error("login failed: internal error", slog.String("message", st.Message()))
		return nil, fmt.Errorf("login failed: %s", st.Message())
	default:
		c.log.Warn("login failed: unexpected status code", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return nil, fmt.Errorf("login failed: %s", st.Message())
	}
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	resp, err := c.client.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err == nil {
		return &domain.TokenPair{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
		}, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("refresh failed: unexpected error type", slog.Any("error", err))
		return nil, err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("refresh failed: invalid argument")
		return nil, domain.ErrInvalidArgument
	case codes.Unauthenticated:
		c.log.Info("refresh failed: session not found or expired")
		return nil, domain.ErrSessionNotFound
	case codes.NotFound:
		c.log.Info("refresh failed: user not found")
		return nil, domain.ErrUserNotFound
	case codes.FailedPrecondition:
		c.log.Info("refresh failed: user not verified")
		return nil, domain.ErrUserNotVerified
	case codes.Internal:
		c.log.Error("refresh failed: internal error", slog.String("message", st.Message()))
		return nil, fmt.Errorf("refresh failed: %s", st.Message())
	default:
		c.log.Warn("refresh failed: unexpected status code", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return nil, fmt.Errorf("refresh failed: %s", st.Message())
	}
}

func (c *Client) ResendVerification(ctx context.Context, email string) error {
	_, err := c.client.ResendVerification(ctx, &authv1.ResendVerificationRequest{
		Email: email,
	})
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("resend verification failed: unexpected error type", slog.Any("error", err))
		return err
	}

	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok && len(badReq.FieldViolations) > 0 {
			violation := badReq.FieldViolations[0]

			switch st.Code() {
			case codes.InvalidArgument:
				if violation.Field == "email" {
					c.log.Debug("resend verification failed: invalid email format")
					return domain.ErrInvalidEmail
				}
			}
		}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("resend verification failed: invalid argument")
		return domain.ErrInvalidEmail
	case codes.NotFound:
		c.log.Info("resend verification failed: user not found")
		return domain.ErrUserNotFound
	case codes.FailedPrecondition:
		c.log.Info("resend verification failed: user already verified")
		return domain.ErrUserAlreadyVerified
	case codes.Internal:
		c.log.Error("resend verification failed: internal error", slog.String("message", st.Message()))
		return fmt.Errorf("resend verification failed: %s", st.Message())
	default:
		c.log.Warn("resend verification failed: unexpected status code", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return fmt.Errorf("resend verification failed: %s", st.Message())
	}
}

func (c *Client) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	resp, err := c.client.GetUser(ctx, &authv1.GetUserRequest{
		Id: userID,
	})
	if err == nil {
		return &domain.User{
			ID:       resp.Id,
			Email:    resp.Email,
			Username: resp.Username,
		}, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("get user failed: unexpected error type", slog.String("user_id", userID), slog.Any("error", err))
		return nil, err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("get user failed: invalid argument", slog.String("user_id", userID))
		return nil, domain.ErrInvalidArgument
	case codes.NotFound:
		c.log.Info("get user failed: user not found", slog.String("user_id", userID))
		return nil, domain.ErrUserNotFound
	case codes.FailedPrecondition:
		c.log.Info("get user failed: user not verified", slog.String("user_id", userID))
		return nil, domain.ErrUserNotVerified
	case codes.Internal:
		c.log.Error("get user failed: internal error", slog.String("user_id", userID), slog.String("message", st.Message()))
		return nil, fmt.Errorf("get user failed: %s", st.Message())
	default:
		c.log.Warn("get user failed: unexpected status code", slog.String("user_id", userID), slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return nil, fmt.Errorf("get user failed: %s", st.Message())
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}
