package epd7in5bhd

import (
	"bytes"
	"image"
	"image/color"
	"io"

	"golang.org/x/image/draw"
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
		Black:          bytes.Repeat([]byte{0xff}, bufSize),
		Highlight:      make([]byte, bufSize, bufSize),
		Rect:           r,
		rectWidthBytes: widthByte,
		Palette:        defaultPalette,
	}
}

type Image struct {
	// This display represents black pixels as 0, white as 1, and a highlight in a separate buffer.
	// Images are stored as a bit per pixel.
	Black []byte
	// Highlights are represented as 0 white, 1 highlight.
	// Images are stored as a bit per pixel.
	Highlight      []byte
	Rect           image.Rectangle
	Palette        color.Palette
	rectWidthBytes int
}

func (i *Image) SetColorIndex(x, y int, index uint8) {
	px := (x / 8) + (y * i.rectWidthBytes)
	if px >= len(i.Black) {
		return
	}
	bit := byte(0x80 >> (uint32(x) % 8))
	switch index {
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

func (i *Image) Set(x, y int, c color.Color) {
	px := (x / 8) + (y * i.rectWidthBytes)
	if px >= len(i.Black) {
		return
	}
	var cc Color
	if native, ok := c.(Color); ok {
		cc = native
	} else {
		cc = i.Palette.Convert(c).(Color)
	}
	bit := byte(0x80 >> (uint32(x) % 8))
	switch cc.C {
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

func (i *Image) Reset() {
	i.Black = bytes.Repeat([]byte{0xff}, len(i.Black))
	i.Highlight = make([]byte, len(i.Highlight), len(i.Highlight))
}

// drawExactColors is a fast-path for when we have exactly 3 colors in the src image.
//
// If src is a *image.Paletted with exactly 3 colors, each color will be assigned to its
// nearest by euclidean distance. Otherwise, colors will be assigned by a per-pixel calculation.
func (i *Image) drawExactColors(src *image.Paletted) {
	white, black, highlight := exactColorIndex(src)
	for y := 0; y < DisplayBounds.Dy(); y++ {
		for x := 0; x < DisplayBounds.Dx(); x++ {
			switch int(src.ColorIndexAt(x, y)) {
			case white:
				i.SetColorIndex(x, y, 0)
			case black:
				i.SetColorIndex(x, y, 1)
			case highlight:
				i.SetColorIndex(x, y, 2)
			}
		}
	}
}

func exactColorIndex(src *image.Paletted) (white, black, highlight int) {
	// This order is significant. We want to try to assign white and black before our third color,
	// as they may be closer to a totally non-red color (blue).
	colors := []color.Color{color.White, color.Black, color.RGBA{255, 0, 0, 255}}
	p := color.Palette{}
	ip := make(color.Palette, len(src.Palette))
	copy(ip, src.Palette)
	// Sort Palette p:
	// src.Palette lightest, src.Palette darkest, src.Palette remaining
	// Iterate over colors, popping as we go to avoid duplicates.
	// We don't want both faint red and white to be white.
	for _, c := range colors {
		ci := ip.Index(c)
		p = append(p, ip[ci])
		ip = append(ip[:ci], ip[ci+1:]...)
	}
	// Now, map our expected order to src.Paletted.Palette's order
	return src.Palette.Index(p[0]), src.Palette.Index(p[1]), src.Palette.Index(p[2])
}

// Encode encodes an image to the display's wire format.
func Encode(dstBlack, dstRed io.Writer, img image.Image) {
	bounds := img.Bounds()
	dst := NewImage(bounds)
	if pi, ok := img.(*image.Paletted); ok && len(pi.Palette) == 3 {
		dst.drawExactColors(pi)
	} else {
		draw.Draw(dst, bounds, img, image.Point{0, 0}, draw.Src)
	}
	dstBlack.Write(dst.Black)
	dstRed.Write(dst.Highlight)
}
