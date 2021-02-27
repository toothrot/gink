package epd7in5bhd

import (
	"image"
	"image/color"
	"testing"
)

type pixel struct {
	pt image.Point
	c  color.Color
}
type want struct {
	idx int
	b   byte
	h   byte
}

func TestImageSet(t *testing.T) {
	cases := []struct {
		desc      string
		pixels    []pixel
		wantBlack []want
	}{
		{
			desc: "simple",
			pixels: []pixel{
				{
					pt: image.Point{0, 0},
					c:  color.Black,
				},
				{
					pt: image.Point{7, 0},
					c:  Highlight,
				},
				{
					pt: image.Point{8, 0},
					c:  color.Black,
				},
				{
					pt: image.Point{15, 0},
					c:  Highlight,
				},
				{
					pt: image.Point{0, 1},
					c:  color.Black,
				},
				{
					pt: image.Point{7, 1},
					c:  Highlight,
				},
			},
			wantBlack: []want{
				{
					idx: 0,
					b:   0b0111_1111,
					h:   0b0000_0001,
				},
				{
					idx: 1,
					b:   0b0111_1111,
					h:   0b0000_0001,
				},
				{
					idx: 2,
					b:   0b0111_1111,
					h:   0b0000_0001,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			img := NewImage(image.Rect(0, 0, 16, 2))
			for _, p := range c.pixels {
				img.Set(p.pt.X, p.pt.Y, p.c)
			}
			for _, w := range c.wantBlack {
				if img.Black[w.idx] != w.b {
					t.Errorf("img.Black[%d] = %08b, wanted %08b", w.idx, img.Black[w.idx], w.b)
				}
				if img.Highlight[w.idx] != w.h {
					t.Errorf("img.Highlight[%d] = %08b, wanted %08b", w.idx, img.Highlight[w.idx], w.h)
				}
			}
		})
	}
}
