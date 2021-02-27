package epd7in5bhd

import (
	"bytes"
	"image"
	"image/color"
)

var (
	White     = Color{0}
	Black     = Color{1}
	Highlight = Color{2}

	Model = color.ModelFunc(model)

	defaultPalette = color.Palette{White, Black, Highlight}
)

type Color struct {
	// 0 white, 1 black, 2 highlight
	C uint8
}

func (c Color) RGBA() (r, g, b, a uint32) {
	switch c.C {
	case 0:
		return 0xffff, 0xffff, 0xffff, 0xffff
	case 1:
		return 0, 0, 0, 0xffff
	case 2:
		return 0xffff, 0, 0, 0xffff
	}
	return 0, 0, 0, 0
}

func model(c color.Color) color.Color {
	return defaultPalette.Convert(c)
}

func NewImage(r image.Rectangle) *Image {
	widthByte := r.Dx() / 8
	if r.Dx()%8 != 0 {
		widthByte += 1
	}
	bufSize := r.Dy() * widthByte
	return &Image{
		Black:     bytes.Repeat([]byte{0xff}, bufSize),
		Highlight: make([]byte, bufSize, bufSize),
		Rect:      r,
	}
}

type Image struct {
	// This display represents black pixels as 0, white as 1, and a highlight in a separate buffer.
	// Images are stored as a bit per pixel.
	Black []byte
	// Highlights are represented as 0 white, 1 highlight.
	// Images are stored as a bit per pixel.
	Highlight []byte
	Rect      image.Rectangle
}

func (i *Image) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}
	var pi int
	if cc, ok := c.(Color); ok {
		pi = int(cc.C)
	} else {
		pi = defaultPalette.Index(c)
	}
	px := (x / 8) + y*DisplayWidthBytes
	bit := byte(0x80 >> (uint32(x) % 8))
	switch pi {
	case 0:
		i.Black[px] |= bit
		i.Highlight[px] &= ^bit
	case 1:
		i.Black[px] &= ^bit
		i.Highlight[px] &= ^bit
	case 2:
		i.Black[px] |= bit
		i.Highlight[px] |= bit
	}
	return
}

func (i *Image) ColorModel() color.Model {
	return Model
}

func (i *Image) Bounds() image.Rectangle {
	return i.Rect
}

func (i *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}).In(i.Rect) {
		return White
	}
	px := (x / 8) + y*DisplayWidthBytes
	bit := byte(0x80 >> (uint32(x) % 8))
	bbit := i.Black[px] & bit
	hbit := i.Highlight[px] & bit
	if hbit >= 1 {
		return Highlight
	}
	if bbit >= 1 {
		return White
	}
	return Black
}
