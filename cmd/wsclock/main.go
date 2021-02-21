// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Binary wsclock displays a clock on a waveshare display.
package main

import (
	"bytes"
	"flag"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/toothrot/gowaveshare/devices/epd7in5bhd"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/opentype"
)

var (
	format = flag.String("format", time.RFC822, "time.Time format.")
	rotate = flag.Float64("rotate", 0.0, "Image rotation in degrees.")
	red    = flag.Bool("red", false, "Render in red instead of black.")
)

func main() {
	flag.Parse()
	d, err := epd7in5bhd.New(epd7in5bhd.DefaultPins)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing")
	d.Init()
	defer d.Sleep()
	log.Println("Clearing")
	d.Clear()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select{
		case s := <-c:
			log.Printf("Got signal %q, quitting", s.String())
			d.Clear()
			time.Sleep(epd7in5bhd.DefaultWait)
			return
		case t := <-ticker.C:
			update(d, t.Format(*format))
		}
	}
}

func update(d *epd7in5bhd.Display, text string) {
	d.Reset()
	img := imaging.New(epd7in5bhd.DisplayWidth, epd7in5bhd.DisplayHeight, color.White)
	ctx := gg.NewContextForImage(img)
	ctx.SetFontFace(fontFace())
	ctx.SetRGB(0, 0, 0)

	ctx.DrawStringWrapped(text, epd7in5bhd.DisplayWidth/2, epd7in5bhd.DisplayHeight/2, 0.5, 0.5, epd7in5bhd.DisplayWidth-80, 1.0, gg.AlignCenter)
	rot := imaging.Rotate(ctx.Image(), *rotate, color.White)
	fit := imaging.Fit(rot, epd7in5bhd.DisplayWidth, epd7in5bhd.DisplayHeight, imaging.Lanczos)
	final := imaging.PasteCenter(imaging.New(epd7in5bhd.DisplayWidth, epd7in5bhd.DisplayHeight, color.White), fit)
	if *red {
		d.Render(nil, bytes.NewReader(epd7in5bhd.Convert(final)))
	} else {
		d.Render(bytes.NewReader(epd7in5bhd.Convert(final)), nil)
	}
	d.Sleep()
}

func fontFace() font.Face {
	f, err := opentype.Parse(gomonobold.TTF)
	if err != nil {
		log.Fatal(err)
	}
	ff, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    128,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	return ff
}
