// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Binary wsimage displays an image on a waveshare display.
package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"time"

	"github.com/disintegration/imaging"
	"github.com/toothrot/gowaveshare/devices/epd7in5bhd"
	"github.com/toothrot/gowaveshare/static"
)

var (
	rotate = flag.Float64("rotate", 0.0, "Image rotation in degrees.")
)

func main() {
	flag.Parse()
	d, err := epd7in5bhd.New(epd7in5bhd.DefaultPins)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing")
	d.Init()
	log.Println("Clearing")
	d.Clear()
	log.Printf("Waiting %vs", epd7in5bhd.DefaultWait.Seconds())
	time.Sleep(epd7in5bhd.DefaultWait)

	bimg, err := staticImage("images/7in5B_HD_b.png")
	if err != nil {
		log.Fatal(err)
	}
	rimg, err := staticImage("images/7in5B_HD_r.png")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Displaying image")
	d.Render(bytes.NewReader(epd7in5bhd.Convert(bimg)), bytes.NewReader(epd7in5bhd.Convert(rimg)))
	log.Printf("Waiting %vs", epd7in5bhd.DefaultWait.Seconds())
	time.Sleep(epd7in5bhd.DefaultWait)
	log.Println("Powering off")
	d.Sleep()
}

func staticImage(path string) (image.Image, error) {
	imgf, err := static.Images.Open(path)
	if err != nil {
	}
	img, _, err := image.Decode(imgf)
	if err != nil {
		return nil, err
	}
	rot := imaging.Rotate(img, *rotate, color.White)
	fit := imaging.Fit(rot, epd7in5bhd.DisplayWidth, epd7in5bhd.DisplayHeight, imaging.Lanczos)
	return imaging.PasteCenter(imaging.New(epd7in5bhd.DisplayWidth, epd7in5bhd.DisplayHeight, color.White), fit), err
}
