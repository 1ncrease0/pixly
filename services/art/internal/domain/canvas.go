package domain

import (
	"errors"
	"image"
	"image/draw"
)

const (
	minSize          = 8
	maxSize          = 128
	DefaultImageSize = 512
)

var ErrInvalidCanvas = errors.New("invalid canvas")

type Canvas struct {
	width   int
	height  int
	palette []Color
	pixels  []int
}

func (c Canvas) Width() int       { return c.width }
func (c Canvas) Height() int      { return c.height }
func (c Canvas) Palette() []Color { return c.palette }
func (c Canvas) Pixels() []int    { return c.pixels }

func NewCanvas(width, height int, palette []Color, pixels []int) (Canvas, error) {
	if width < minSize || width > maxSize || height < minSize || height > maxSize {
		return Canvas{}, ErrInvalidCanvas
	}

	if len(pixels) != width*height {
		return Canvas{}, ErrInvalidCanvas
	}

	if len(palette) < 1 {
		return Canvas{}, ErrInvalidCanvas
	}

	for _, idx := range pixels {
		if idx < 0 || idx >= len(palette) {
			return Canvas{}, ErrInvalidCanvas
		}
	}

	return Canvas{
		width:   width,
		height:  height,
		palette: palette,
		pixels:  pixels,
	}, nil
}

func (c Canvas) Render(outMaxSide int) *image.RGBA {
	w, h := c.width, c.height

	long := w
	if h > long {
		long = h
	}

	scale := outMaxSide / long
	if scale < 1 {
		scale = 1
	}

	outW := w * scale
	outH := h * scale
	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))
	palette := c.palette
	pixels := c.pixels

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := pixels[y*w+x]
			clr := palette[idx].RGBA()
			r := image.Rect(x*scale, y*scale, (x+1)*scale, (y+1)*scale)
			draw.Draw(dst, r, &image.Uniform{C: clr}, image.Point{}, draw.Src)
		}
	}

	return dst
}
