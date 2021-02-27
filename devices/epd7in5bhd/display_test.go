package epd7in5bhd

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/draw"
)

func BenchmarkEncode(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, DisplayWidth, DisplayHeight))
	b.ResetTimer()
	var rbuf, bbuf bytes.Buffer
	for i := 0; i < b.N; i++ {
		Encode(&bbuf, &rbuf, img)
		rbuf.Reset()
		bbuf.Reset()
	}
}

func BenchmarkEncodeExactPalette(b *testing.B) {
	img := image.NewPaletted(image.Rect(0, 0, DisplayWidth, DisplayHeight), color.Palette{color.White, color.Black, color.RGBA{255, 0, 0, 255}})
	b.ResetTimer()
	var rbuf, bbuf bytes.Buffer
	for i := 0; i < b.N; i++ {
		Encode(&bbuf, &rbuf, img)
		rbuf.Reset()
		bbuf.Reset()
	}
}

func BenchmarkEncodeExactPaletteDifferentHighlight(b *testing.B) {
	img := image.NewPaletted(image.Rect(0, 0, DisplayWidth, DisplayHeight), color.Palette{color.White, color.Black, color.RGBA{0, 0, 255, 255}})
	b.ResetTimer()
	var rbuf, bbuf bytes.Buffer
	for i := 0; i < b.N; i++ {
		Encode(&bbuf, &rbuf, img)
		rbuf.Reset()
		bbuf.Reset()
	}
}

func BenchmarkEncodeTwoColor(b *testing.B) {
	img := image.NewPaletted(image.Rect(0, 0, DisplayWidth, DisplayHeight), color.Palette{color.White, color.Black})
	b.ResetTimer()
	var rbuf, bbuf bytes.Buffer
	for i := 0; i < b.N; i++ {
		Encode(&bbuf, &rbuf, img)
		rbuf.Reset()
		bbuf.Reset()
	}
}

func BenchmarkDraw(b *testing.B) {
	u := image.NewUniform(color.Black)
	r := image.Rect(0, 0, DisplayWidth, DisplayHeight)
	img := NewImage(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(img, r, u, image.Point{0, 0}, draw.Over)
	}
}

func BenchmarkDrawPaletted(b *testing.B) {
	r := image.Rect(0, 0, DisplayWidth, DisplayHeight)
	p := image.NewPaletted(r, color.Palette{color.White, color.Black, color.RGBA{255, 0, 0, 255}})
	img := NewImage(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(img, r, p, image.Point{0, 0}, draw.Over)
	}
}

func BenchmarkDrawPalettedNativeColor(b *testing.B) {
	r := image.Rect(0, 0, DisplayWidth, DisplayHeight)
	p := image.NewPaletted(r, color.Palette{White, Black, Highlight})
	img := NewImage(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(img, r, p, image.Point{0, 0}, draw.Src)
	}
}
