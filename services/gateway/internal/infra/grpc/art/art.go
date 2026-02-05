package art

import (
	"context"
	"fmt"

	artv1 "github.com/1ncrease0/pixly/proto/gen/art"
	"github.com/1ncrease0/pixly/services/gateway/internal/domain"
	artdomain "github.com/1ncrease0/pixly/services/gateway/internal/domain/art"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type Client struct {
	conn   *grpc.ClientConn
	client artv1.ArtClient
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
		client: artv1.NewArtClient(conn),
		log:    log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SavePixelart(ctx context.Context, in artdomain.SavePixelartInput) (string, error) {
	resp, err := c.client.SavePixelart(ctx, &artv1.SavePixelartRequest{
		UserId:  in.UserID,
		Title:   in.Title,
		Palette: in.Palette,
		Pixels:  in.Pixels,
		Width:   in.Width,
		Height:  in.Height,
	})
	if err == nil {
		return resp.PixelartId, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("save pixelart failed: unexpected error type", slog.Any("error", err))
		return "", err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("save pixelart failed: invalid argument")
		return "", domain.ErrInvalidArgument
	case codes.AlreadyExists:
		c.log.Debug("save pixelart failed: pixelart conflict")
		return "", domain.ErrPixelartConflict
	case codes.Internal:
		c.log.Error("save pixelart failed: internal error", slog.String("message", st.Message()))
		return "", fmt.Errorf("art service: %s", st.Message())
	case codes.Unavailable:
		c.log.Warn("save pixelart failed: art service unavailable")
		return "", fmt.Errorf("art service unavailable")
	default:
		c.log.Warn("save pixelart failed: unexpected status", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return "", fmt.Errorf("art service: %s", st.Message())
	}
}

func (c *Client) UpdateCanvas(ctx context.Context, in artdomain.UpdateCanvasInput) error {
	_, err := c.client.UpdateCanvas(ctx, &artv1.UpdateCanvasRequest{
		UserId:     in.UserID,
		PixelartId: in.PixelartID,
		Palette:    in.Palette,
		Pixels:     in.Pixels,
		Width:      in.Width,
		Height:     in.Height,
	})
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("update canvas failed: unexpected error type", slog.Any("error", err))
		return err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("update canvas failed: invalid argument")
		return domain.ErrInvalidArgument
	case codes.NotFound:
		c.log.Debug("update canvas failed: pixelart not found")
		return domain.ErrPixelartNotFound
	case codes.Internal:
		c.log.Error("update canvas failed: internal error", slog.String("message", st.Message()))
		return fmt.Errorf("art service: %s", st.Message())
	case codes.Unavailable:
		c.log.Warn("update canvas failed: art service unavailable")
		return fmt.Errorf("art service unavailable")
	default:
		c.log.Warn("update canvas failed: unexpected status", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return fmt.Errorf("art service: %s", st.Message())
	}
}

func (c *Client) DeletePixelart(ctx context.Context, in artdomain.DeletePixelartInput) error {
	_, err := c.client.DeletePixelart(ctx, &artv1.DeletePixelartRequest{
		UserId:     in.UserID,
		PixelartId: in.PixelartID,
	})
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("delete pixelart failed: unexpected error type", slog.Any("error", err))
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		c.log.Debug("delete pixelart failed: pixelart not found")
		return domain.ErrPixelartNotFound
	case codes.Internal:
		c.log.Error("delete pixelart failed: internal error", slog.String("message", st.Message()))
		return fmt.Errorf("art service: %s", st.Message())
	case codes.Unavailable:
		c.log.Warn("delete pixelart failed: art service unavailable")
		return fmt.Errorf("art service unavailable")
	default:
		c.log.Warn("delete pixelart failed: unexpected status", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return fmt.Errorf("art service: %s", st.Message())
	}
}

func (c *Client) GetUserPixelart(ctx context.Context, in artdomain.GetUserPixelartInput) (*artdomain.Pixelart, error) {
	resp, err := c.client.GetUserPixelart(ctx, &artv1.GetUserPixelartRequest{
		UserId:     in.UserID,
		PixelartId: in.PixelartID,
	})
	if err == nil {
		return &artdomain.Pixelart{
			Title:    resp.Title,
			Palette:  resp.Palette,
			Pixels:   resp.Pixels,
			Width:    resp.Width,
			Height:   resp.Height,
			ImageURL: resp.ImageUrl,
		}, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("get user pixelart failed: unexpected error type", slog.Any("error", err))
		return nil, err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.log.Debug("get user pixelart failed: invalid argument")
		return nil, domain.ErrInvalidArgument
	case codes.NotFound:
		c.log.Debug("get user pixelart failed: pixelart not found")
		return nil, domain.ErrPixelartNotFound
	case codes.Internal:
		c.log.Error("get user pixelart failed: internal error", slog.String("message", st.Message()))
		return nil, fmt.Errorf("art service: %s", st.Message())
	case codes.Unavailable:
		c.log.Warn("get user pixelart failed: art service unavailable")
		return nil, fmt.Errorf("art service unavailable")
	default:
		c.log.Warn("get user pixelart failed: unexpected status", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return nil, fmt.Errorf("art service: %s", st.Message())
	}
}

func (c *Client) GetUserPreviews(ctx context.Context, in artdomain.GetUserPreviewsInput) ([]artdomain.Preview, error) {
	resp, err := c.client.GetUserPreviews(ctx, &artv1.GetUserPreviewsRequest{
		UserId: in.UserID,
	})
	if err == nil {
		out := make([]artdomain.Preview, len(resp.Previews))
		for i, p := range resp.Previews {
			out[i] = artdomain.Preview{
				PixelartID: p.PixelartId,
				Title:      p.Title,
				ImageURL:   p.ImageUrl,
			}
		}
		return out, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		c.log.Error("get user previews failed: unexpected error type", slog.Any("error", err))
		return nil, err
	}

	switch st.Code() {
	case codes.Internal:
		c.log.Error("get user previews failed: internal error", slog.String("message", st.Message()))
		return nil, fmt.Errorf("art service: %s", st.Message())
	case codes.Unavailable:
		c.log.Warn("get user previews failed: art service unavailable")
		return nil, fmt.Errorf("art service unavailable")
	default:
		c.log.Warn("get user previews failed: unexpected status", slog.String("code", st.Code().String()), slog.String("message", st.Message()))
		return nil, fmt.Errorf("art service: %s", st.Message())
	}
}
