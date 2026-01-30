package art

import (
	"bytes"
	"github.com/1ncrease0/pixly/services/art/internal/domain"
	"github.com/google/uuid"
	"image/png"
)

type ImageProvider interface {
}

type PixelartRepo interface {
	CreatePixelart(p *domain.Pixelart) error
}

type Service struct {
	imgProvider ImageProvider
	repo        PixelartRepo
}

func NewService(r PixelartRepo, p ImageProvider) *Service {
	return &Service{
		imgProvider: p,
		repo:        r,
	}
}

func (s *Service) SavePixelart(userID uuid.UUID, title domain.Title, canvas domain.Canvas) (uuid.UUID, error) {
	id := uuid.New()

	art := domain.NewPixelart(id, userID, title, canvas)

	err := s.repo.CreatePixelart(art)
	if err != nil {
		return uuid.Nil, err
	}

	img := art.Render(domain.DefaultImageSize)

}
