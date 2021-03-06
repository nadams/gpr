package colr

import (
	"image"
	"image/color"
)

// MonoGreen64 represents an alpha-premultiplied 64-bit color.
// Green channel is set to average of RGB inputs, with GB set to zero.
type MonoGreen64 struct {
	R, G, B, A uint16
}

func (c MonoGreen64) RGBA() (r, g, b, a uint32) {
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}

func monogreen64Model(c color.Color) color.Color {
	if _, ok := c.(MonoGreen64); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	r = 0
	b = 0
	return MonoGreen64{uint16(r), uint16(g), uint16(b), uint16(a)}
}

var MonoGreen64Model color.Model = color.ModelFunc(monogreen64Model)

// MonoGreen64Image is an in memory image of MonoGreen64 color values
func NewMonoGreen64Image(r image.Rectangle) *MonoGreen64Image {
	w, h := r.Dx(), r.Dy()
	pix := make([]uint8, 8*w*h)
	return &MonoGreen64Image{pix, 8 * w, r}
}

type MonoGreen64Image struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
}

func (p *MonoGreen64Image) ColorModel() color.Model {
	return MonoGreen64Model
}

func (p *MonoGreen64Image) Bounds() image.Rectangle {
	return p.Rect
}

func (p *MonoGreen64Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return MonoGreen64{}
	}
	i := p.PixOffset(x, y)
	return MonoGreen64{uint16(p.Pix[i+0]<<8 | p.Pix[i+1]),
		uint16(p.Pix[i+2]<<8 | p.Pix[i+3]),
		uint16(p.Pix[i+4]<<8 | p.Pix[i+5]),
		uint16(p.Pix[i+6]<<8 | p.Pix[i+7]),
	}
}

func (p *MonoGreen64Image) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*8
}

func (p *MonoGreen64Image) Set(x, y int, c MonoGreen64) {
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
