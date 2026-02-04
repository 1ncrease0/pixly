package art

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	artv1 "github.com/1ncrease0/pixly/proto/gen/art"
	"github.com/1ncrease0/pixly/services/art/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	SavePixelart(ctx context.Context, userID uuid.UUID, title domain.Title, canvas domain.Canvas) (uuid.UUID, error)
	UpdateCanvas(ctx context.Context, userID, pixelartID uuid.UUID, canvas domain.Canvas) error
	DeletePixelart(ctx context.Context, userID, pixelartID uuid.UUID) error
	UserPixelart(ctx context.Context, pixelartID, userID uuid.UUID) (*domain.Pixelart, error)
	UserPreviews(ctx context.Context, userID uuid.UUID) ([]domain.PixelartPreview, error)
}

type Server struct {
	artv1.UnimplementedArtServer
	service Service
	log     *slog.Logger
}

func Register(gRPCServer *grpc.Server, s Service, log *slog.Logger) {
	artv1.RegisterArtServer(gRPCServer, &Server{service: s, log: log})
}

func (s *Server) SavePixelart(ctx context.Context, req *artv1.SavePixelartRequest) (*artv1.SavePixelartResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "user_id", Description: "invalid user id",
		})
	}

	title, err := domain.NewTitle(req.Title)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "title", Description: "invalid title length",
		})
	}

	palette, palViolations := s.parsePalette(req.Palette)
	violations = append(violations, palViolations...)

	pixels := make([]int, len(req.Pixels))
	for i, v := range req.Pixels {
		pixels[i] = int(v)
	}
	canvas, err := domain.NewCanvas(int(req.Width), int(req.Height), palette, pixels)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "canvas", Description: "invalid canvas (size 8–128, pixels length must match width*height, palette indices in range)",
		})
	}

	if len(violations) > 0 {
		s.log.Info("save pixelart validation failed", slog.String("user_id", req.UserId), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		return nil, st.Err()
	}

	id, err := s.service.SavePixelart(ctx, userID, title, canvas)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPixelartConflict):
			s.log.Info("save pixelart failed: conflict", slog.String("user_id", req.UserId))
			return nil, status.Error(codes.AlreadyExists, "pixelart conflict")
		default:
			s.log.Error("save pixelart failed", slog.String("user_id", req.UserId), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to save pixelart")
		}
	}
	s.log.Debug("save pixelart ok", slog.String("user_id", req.UserId), slog.String("pixelart_id", id.String()))
	return &artv1.SavePixelartResponse{PixelartId: id.String()}, nil
}

func (s *Server) UpdateCanvas(ctx context.Context, req *artv1.UpdateCanvasRequest) (*artv1.UpdateCanvasResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "user_id", Description: "invalid user id",
		})
	}
	pixelartID, err := uuid.Parse(req.PixelartId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "pixelart_id", Description: "invalid pixelart id",
		})
	}

	palette, palViolations := s.parsePalette(req.Palette)
	violations = append(violations, palViolations...)

	pixels := make([]int, len(req.Pixels))
	for i, v := range req.Pixels {
		pixels[i] = int(v)
	}
	canvas, err := domain.NewCanvas(int(req.Width), int(req.Height), palette, pixels)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "canvas", Description: "invalid canvas (size 8–128, pixels length must match width*height, palette indices in range)",
		})
	}

	if len(violations) > 0 {
		s.log.Info("update canvas validation failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		return nil, st.Err()
	}

	err = s.service.UpdateCanvas(ctx, userID, pixelartID, canvas)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPixelartNotFound):
			s.log.Info("update canvas failed: pixelart not found", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
			return nil, status.Error(codes.NotFound, "pixelart not found")
		default:
			s.log.Error("update canvas failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to update canvas")
		}
	}
	s.log.Debug("update canvas ok", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
	return &artv1.UpdateCanvasResponse{}, nil
}

func (s *Server) DeletePixelart(ctx context.Context, req *artv1.DeletePixelartRequest) (*artv1.DeletePixelartResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "user_id", Description: "invalid user id",
		})
	}
	pixelartID, err := uuid.Parse(req.PixelartId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "pixelart_id", Description: "invalid pixelart id",
		})
	}

	if len(violations) > 0 {
		s.log.Info("delete pixelart validation failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		return nil, st.Err()
	}

	err = s.service.DeletePixelart(ctx, userID, pixelartID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPixelartNotFound):
			s.log.Info("delete pixelart failed: pixelart not found", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
			return nil, status.Error(codes.NotFound, "pixelart not found")
		default:
			s.log.Error("delete pixelart failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to delete pixelart")
		}
	}
	s.log.Debug("delete pixelart ok", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
	return &artv1.DeletePixelartResponse{}, nil
}

func (s *Server) GetUserPixelart(ctx context.Context, req *artv1.GetUserPixelartRequest) (*artv1.GetUserPixelartResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "user_id", Description: "invalid user id",
		})
	}
	pixelartID, err := uuid.Parse(req.PixelartId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "pixelart_id", Description: "invalid pixelart id",
		})
	}

	if len(violations) > 0 {
		s.log.Info("get user pixelart validation failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		return nil, st.Err()
	}

	art, err := s.service.UserPixelart(ctx, pixelartID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPixelartNotFound):
			s.log.Info("get user pixelart failed: pixelart not found", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
			return nil, status.Error(codes.NotFound, "pixelart not found")
		default:
			s.log.Error("get user pixelart failed", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId), slog.Any("error", err))
			return nil, status.Error(codes.Internal, "failed to get pixelart")
		}
	}

	canvas := art.Canvas()
	paletteHex := make([]string, 0, len(canvas.Palette()))
	for _, c := range canvas.Palette() {
		rgba := c.RGBA()
		paletteHex = append(paletteHex, fmt.Sprintf("#%02X%02X%02X", rgba.R, rgba.G, rgba.B))
	}
	pixels64 := make([]int64, len(canvas.Pixels()))
	for i, p := range canvas.Pixels() {
		pixels64[i] = int64(p)
	}

	s.log.Debug("get user pixelart ok", slog.String("user_id", req.UserId), slog.String("pixelart_id", req.PixelartId))
	return &artv1.GetUserPixelartResponse{
		Palette:  paletteHex,
		Pixels:   pixels64,
		Width:    int64(canvas.Width()),
		Height:   int64(canvas.Height()),
		Title:    art.Title().String(),
		ImageUrl: art.ImageKey(),
	}, nil
}

func (s *Server) GetUserPreviews(ctx context.Context, req *artv1.GetUserPreviewsRequest) (*artv1.GetUserPreviewsResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field: "user_id", Description: "invalid user id",
		})
	}

	if len(violations) > 0 {
		s.log.Info("get user previews validation failed", slog.String("user_id", req.UserId), slog.Any("violations", violations))
		st := status.New(codes.InvalidArgument, "validation failed")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		return nil, st.Err()
	}

	previews, err := s.service.UserPreviews(ctx, userID)
	if err != nil {
		s.log.Error("get user previews failed", slog.String("user_id", req.UserId), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to get previews")
	}

	out := make([]*artv1.Preview, 0, len(previews))
	for _, p := range previews {
		out = append(out, &artv1.Preview{
			PixelartId: p.ID().String(),
			Title:      p.Title().String(),
			ImageUrl:   p.ImageURL(),
		})
	}

	s.log.Debug("get user previews ok", slog.String("user_id", req.UserId), slog.Int("count", len(previews)))
	return &artv1.GetUserPreviewsResponse{Previews: out}, nil
}

func (s *Server) parsePalette(hexes []string) ([]domain.Color, []*errdetails.BadRequest_FieldViolation) {
	if len(hexes) == 0 {
		return nil, []*errdetails.BadRequest_FieldViolation{
			{Field: "palette", Description: "at least one color required"},
		}
	}
	palette := make([]domain.Color, 0, len(hexes))
	for i, h := range hexes {
		c, err := domain.NewColor(h)
		if err != nil {
			return nil, []*errdetails.BadRequest_FieldViolation{
				{Field: "palette", Description: "invalid color at index " + strconv.Itoa(i) + ": must be #RRGGBB"},
			}
		}
		palette = append(palette, *c)
	}
	return palette, nil
}
