package domain

import (
	"errors"
	"image/color"
	"strconv"
)

const (
	maxTitleLength = 32
)

var (
	ErrTitleLength  = errors.New("invalid title length")
	ErrInvalidColor = errors.New("invalid color: must be #RRGGBB")
)

type Color struct {
	rgba color.RGBA
}

func NewColor(hex string) (*Color, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return nil, ErrInvalidColor
	}
	values, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return nil, ErrInvalidColor
	}
	c := color.RGBA{
		R: uint8(values >> 16),
		G: uint8((values >> 8) & 0xFF),
		B: uint8(values & 0xFF),
		A: 255,
	}
	return &Color{c}, nil
}

func (c *Color) RGBA() color.RGBA {
	return c.rgba
}

type Title struct {
	title string
}

func NewTitle(title string) (Title, error) {
	if len(title) > maxTitleLength {
		return Title{}, ErrTitleLength
	}

	return Title{title: title}, nil
}

func (t Title) String() string {
	return t.title
}
