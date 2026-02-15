package art

import (
	"bytes"
	"context"
	"image/png"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/1ncrease0/pixly/services/art/internal/domain"
)

//TODO refactor

type ImageProvider interface {
	GetImageURL(ctx context.Context, objectKey string) (string, error)
	DeleteImage(ctx context.Context, objectKey string) error
	SaveImage(ctx context.Context, objectKey string, reader io.Reader, size int64) error
}

type PixelartRepo interface {
	CreatePixelart(ctx context.Context, p *domain.Pixelart) error
	PixelartByID(ctx context.Context, id uuid.UUID) (*domain.Pixelart, error)
	PixelartsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Pixelart, error)
	UpdatePixelart(ctx context.Context, p *domain.Pixelart) error
	DeletePixelart(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	imageProvider ImageProvider
	repo          PixelartRepo
	log           *slog.Logger
}

func NewService(r PixelartRepo, p ImageProvider, log *slog.Logger) *Service {
	return &Service{
		imageProvider: p,
		repo:          r,
		log:           log,
	}
}

func (s *Service) SavePixelart(ctx context.Context, userID uuid.UUID, title domain.Title, canvas domain.Canvas) (uuid.UUID, error) {
	id := uuid.New()
	objectKey := "pixelarts/" + id.String() + ".png"

	art := domain.NewPixelart(id, userID, title, canvas)

	img := art.Render(domain.DefaultImageSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return uuid.Nil, err
	}

	err := s.imageProvider.SaveImage(ctx, objectKey, &buf, int64(buf.Len()))
	if err != nil {
		return uuid.Nil, err
	}

	art.SetKey(objectKey)

	err = s.repo.CreatePixelart(ctx, art)
	if err != nil {
		nerr := s.imageProvider.DeleteImage(ctx, objectKey)
		if nerr != nil {
			s.log.Error("delete image", slog.String("key", objectKey), slog.Any("err", err))
		}

		return uuid.Nil, err
	}

	return id, nil
}

func (s *Service) UpdateCanvas(ctx context.Context, userID, id uuid.UUID, canvas domain.Canvas) error {
	art, err := s.repo.PixelartByID(ctx, id)
	if err != nil {
		return err
	}

	if userID != art.UserID() {
		return domain.ErrPixelartNotFound
	}

	art.ChangeCanvas(canvas)

	if err := s.repo.UpdatePixelart(ctx, art); err != nil {
		return err
	}

	img := art.Render(domain.DefaultImageSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	return s.imageProvider.SaveImage(ctx, art.ImageKey(), &buf, int64(buf.Len()))
}

func (s *Service) UserPixelart(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Pixelart, error) {
	art, err := s.repo.PixelartByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if art.UserID() != userID {
		return nil, domain.ErrPixelartNotFound
	}

	url, err := s.imageProvider.GetImageURL(ctx, art.ImageKey())
	if err != nil {
		s.log.Warn("get image url", slog.String("key", art.ImageKey()), slog.Any("err", err))
	}

	art.SetKey(url)
	return art, nil
}

func (s *Service) UserPreviews(ctx context.Context, userID uuid.UUID) ([]domain.PixelartPreview, error) {
	arts, err := s.repo.PixelartsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	previews := make([]domain.PixelartPreview, 0, len(arts))
	for _, art := range arts {
		url, err := s.imageProvider.GetImageURL(ctx, art.ImageKey())
		if err != nil {
			s.log.Warn("get image url", slog.String("key", art.ImageKey()), slog.Any("err", err))
			continue
		}
		previews = append(previews, domain.NewPreview(art.ID(), art.Title(), url))
	}
	return previews, nil
}

func (s *Service) DeletePixelart(ctx context.Context, userID, id uuid.UUID) error {
	art, err := s.repo.PixelartByID(ctx, id)
	if err != nil {
		return err
	}

	if userID != art.UserID() {
		return domain.ErrPixelartNotFound
	}

	objectKey := art.ImageKey()

	if err := s.repo.DeletePixelart(ctx, id); err != nil {
		return err
	}

	if err := s.imageProvider.DeleteImage(ctx, objectKey); err != nil {
		s.log.Warn("delete image", slog.String("key", objectKey), slog.Any("err", err))
	}

	return nil
}
