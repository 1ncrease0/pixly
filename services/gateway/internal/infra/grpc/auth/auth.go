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
		return err
	}

	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok && len(badReq.FieldViolations) > 0 {
			violation := badReq.FieldViolations[0]

			switch st.Code() {
			case codes.InvalidArgument:
				switch violation.Field {
				case "email":
					return domain.ErrInvalidEmail
				case "username":
					return domain.ErrInvalidUsername
				case "password":
					return domain.ErrInvalidPassword
				}

			case codes.AlreadyExists:
				switch violation.Field {
				case "email":
					return domain.ErrUserAlreadyExists
				case "username":
					return domain.ErrUsernameTaken
				}
			}
		}
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return domain.ErrInvalidArgument
	case codes.AlreadyExists:
		return domain.ErrAlreadyExists
	default:
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
		return err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return domain.ErrInvalidVerificationCode
	case codes.NotFound:
		return domain.ErrUserNotFound
	default:
		return fmt.Errorf("verify email failed: %s", st.Message())
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}
