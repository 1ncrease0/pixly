package domain

import (
	"github.com/google/uuid"
	"image"
	"time"
)

type Pixelart struct {
	id        uuid.UUID
	userID    uuid.UUID
	title     Title
	canvas    Canvas
	imageKey  string
	createdAt time.Time
	updatedAt time.Time
}

func NewPixelart(id, userID uuid.UUID, title Title, canvas Canvas) *Pixelart {
	return &Pixelart{
		id:     id,
		userID: userID,
		title:  title,
		canvas: canvas,
	}
}

func (p *Pixelart) SetKey(key string) {
	p.imageKey = key
}

func (p *Pixelart) Render(size int) *image.RGBA {
	return p.canvas.Render(size)
}
