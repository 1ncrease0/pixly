package domain

import (
	"errors"
	"image"
	"time"

	"github.com/google/uuid"
)

var (
	ErrImageNotFound    = errors.New("image not found")
	ErrPixelartNotFound = errors.New("pixelart not found")
	ErrNilPixelart      = errors.New("nil pixelart")
	ErrPixelartConflict = errors.New("pixelart conflict")
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

func NewPixelartFromStorage(id, userID uuid.UUID, title Title, canvas Canvas, imageKey string, createdAt, updatedAt time.Time) *Pixelart {
	return &Pixelart{
		id:        id,
		userID:    userID,
		title:     title,
		canvas:    canvas,
		imageKey:  imageKey,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (p *Pixelart) SetKey(key string) {
	p.imageKey = key
}

func (p *Pixelart) ChangeCanvas(newCanvas Canvas) {
	p.canvas = newCanvas
}

func (p *Pixelart) ID() uuid.UUID {
	return p.id
}

func (p *Pixelart) UserID() uuid.UUID {
	return p.userID
}

func (p *Pixelart) Title() Title {
	return p.title
}

func (p *Pixelart) Canvas() Canvas {
	return p.canvas
}

func (p *Pixelart) ImageKey() string {
	return p.imageKey
}

func (p *Pixelart) Render(size int) *image.RGBA {
	return p.canvas.Render(size)
}

type PixelartPreview struct {
	pixelartID uuid.UUID
	title      Title
	imageURL   string
}

func NewPreview(id uuid.UUID, title Title, url string) PixelartPreview {
	return PixelartPreview{
		pixelartID: id,
		title:      title,
		imageURL:   url,
	}
}

func (p PixelartPreview) ID() uuid.UUID {
	return p.pixelartID
}

func (p PixelartPreview) Title() Title {
	return p.title
}

func (p PixelartPreview) ImageURL() string {
	return p.imageURL
}
