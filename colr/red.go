package colr

import (
	"image"
	"image/color"
)

// MonoRed64 represents an alpha-premultiplied 64-bit color.
// Red channel is set to average of RGB inputs, with GB set to zero.
type MonoRed64 struct {
	R, G, B, A uint16
}

func (c MonoRed64) RGBA() (r, g, b, a uint32) {
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}

func monored64Model(c color.Color) color.Color {
	if _, ok := c.(MonoRed64); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	g = 0
	b = 0
	return MonoRed64{uint16(r), uint16(g), uint16(b), uint16(a)}
}

var MonoRed64Model color.Model = color.ModelFunc(monored64Model)

// MonoRed64Image is an in memory image of MonoRed64 color values
func NewMonoRed64Image(r image.Rectangle) *MonoRed64Image {
	w, h := r.Dx(), r.Dy()
	pix := make([]uint8, 8*w*h)
	return &MonoRed64Image{pix, 8 * w, r}
}

type MonoRed64Image struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
}

func (p *MonoRed64Image) ColorModel() color.Model {
	return MonoRed64Model
}

func (p *MonoRed64Image) Bounds() image.Rectangle {
	return p.Rect
}

func (p *MonoRed64Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return MonoRed64{}
	}
	i := p.PixOffset(x, y)
	return MonoRed64{uint16(p.Pix[i+0]<<8 | p.Pix[i+1]),
		uint16(p.Pix[i+2]<<8 | p.Pix[i+3]),
		uint16(p.Pix[i+4]<<8 | p.Pix[i+5]),
		uint16(p.Pix[i+6]<<8 | p.Pix[i+7]),
	}
}

func (p *MonoRed64Image) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*8
}

func (p *MonoRed64Image) Set(x, y int, c MonoRed64) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = uint8(c.R >> 8)
	p.Pix[i+1] = uint8(c.R)
	p.Pix[i+2] = uint8(c.G >> 8)
	p.Pix[i+3] = uint8(c.G)
	p.Pix[i+4] = uint8(c.B >> 8)
	p.Pix[i+5] = uint8(c.B)
	p.Pix[i+6] = uint8(c.A >> 8)
	p.Pix[i+7] = uint8(c.A)
}
