// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package epd7in5bhd is for the Waveshare 7.5 inch HD (B/C) e-Paper display.
package epd7in5bhd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

const (
	// Device width in pixels.
	DisplayWidth = 880
	// Device width in bytes.
	DisplayWidthBytes = 880 / 8
	// Device height in pixels.
	DisplayHeight = 528
	// Full buffer size in bytes.
	BufSize = DisplayWidthBytes * DisplayHeight
)

// Display is a client for the e-Paper display.
//
// Standard pin locations are as follows:
//  Busy - Busy      - Pin 18 (GPIO 24)
//  CLK  - SPI0 SCLK - Pin 23 (GPIO 11)
//  CS   - SPI0 CE0  - Pin 24 (GPIO 8)
//  DC   - Data/Cmd  - Pin 22 (GPIO 25)
//  DIN  - SPI0 MOSI - Pin 19 (GPIO 10)
//  RST  - Reset     - Pin 11 (GPIO 17)
type Display struct {
	hw *hardware
}

type Pins struct {
	// Busy pin name, typically "P1_18"
	Busy string
	// CS pin name, typically "P1_24"
	CS string
	// DC pin name, typically "P1_22"
	DC string
	// RST pin name, typicaly "P1_11"
	RST string
}

var DefaultPins = Pins{
	Busy: "P1_18",
	CS:   "P1_24",
	DC:   "P1_22",
	RST:  "P1_11",
}

// DefaultSleep is the default time to wait for a screen refresh. The official documented refresh time is 22 seconds.
var DefaultWait = 25 * time.Second

// New creates a Display configured for use.
//
// dcPin, csPin, rstPin, and busyPin all expect valid gpioreg.ByName() values, such as P1_22.
//
//  d, err := epd7in5bhd.New("P1_22", "P1_24", "P1_11", "P1_18")
//  if err != nil {
//    // Handle error.
//  }
func New(p Pins) (*Display, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("host.Init() = %w", err)
	}

	dc := gpioreg.ByName(p.DC)
	if dc == nil {
		return nil, fmt.Errorf("invalid dc pin %q", p.DC)
	}
	if err := dc.Out(gpio.Low); err != nil {
		return nil, fmt.Errorf("dc.Out(%v) = %w", gpio.Low, err)
	}

	cs := gpioreg.ByName(p.CS)
	if cs == nil {
		return nil, fmt.Errorf("invalid cs pin %q", p.CS)
	}
	if err := cs.Out(gpio.Low); err != nil {
		return nil, fmt.Errorf("cs.Out(%v) = %w", gpio.Low, err)
	}

	rst := gpioreg.ByName(p.RST)
	if rst == nil {
		return nil, fmt.Errorf("invalid rst pin %q", p.RST)
	}
	if err := rst.Out(gpio.Low); err != nil {
		return nil, fmt.Errorf("rst.Out(%v) = %w", gpio.Low, err)
	}

	busy := gpioreg.ByName(p.Busy)
	if busy == nil {
		return nil, fmt.Errorf("invalid busy pin %q", p.Busy)
	}
	if err := busy.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		return nil, fmt.Errorf("busy.In(%v, %v) = %w", gpio.PullDown, gpio.RisingEdge, err)
	}

	port, err := spireg.Open("")
	if err != nil {
		return nil, fmt.Errorf("spireg.Open(%q) = _, %w", "", err)
	}
	// 20Mhz is the max for write operations. 2.5Mhz is the max for read operations.
	// Wire length and health impact the maximum workable speed.
	c, err := port.Connect(20*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		connerr := fmt.Errorf("port.Connect(%v, %v, %v) = %w", 5*physic.MegaHertz, spi.Mode0, 8, err)
		if err := port.Close(); err != nil {
			return nil, fmt.Errorf("port.Close() = %w while handling %q", err, connerr)
		}
		return nil, connerr
	}

	e := &Display{
		hw: &hardware{
			txLimit: 4096,
			c:       c,
			dc:      dc,
			cs:      cs,
			rst:     rst,
			busy:    busy,
		},
	}
	return e, nil
}

// Reset can be also used to awaken the device.
func (d *Display) Reset() {
	d.hw.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
	d.hw.rst.Out(gpio.Low)
	time.Sleep(2 * time.Millisecond)
	d.hw.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
}

func (d *Display) sendCommand(cmd command, data ...byte) {
	now := time.Now()
	defer func(start time.Time) {
		d := time.Since(start)
		if d > time.Millisecond {
			log.Printf("sendCommand: %s", time.Since(start).String())
		}
	}(now)
	n, err := d.hw.CommandWriter().Write(append([]byte{byte(cmd)}, data...))
	if err != nil {
		log.Printf("sendCommand Write() = %d, %v", n, err)
	}
}

func (d *Display) sendData(data []byte) {
	if n, err := d.hw.DataWriter().Write(data); err != nil {
		log.Printf("sendData failed: %d, %v", n, err)
	}
}

func (d *Display) waitUntilIdle() {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("waitUntilIdle: %s", time.Since(start).String())
	}(now)
	for d.hw.busy.Read() == gpio.Low {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(10 * time.Millisecond)
}

func (d *Display) turnOnDisplay() {
	// Load LUT from MCU(0x32)
	d.sendCommand(displayUpdateControl2, 0xC7)
	d.sendCommand(masterActivation)
	time.Sleep(2 * time.Millisecond) //!!!The delay here is necessary, 200uS at least!!!
	d.waitUntilIdle()                //waiting for the electronic paper IC to release the idle signal
}

// Init initializes the display config. It should be used if the device is asleep and needs reinitialization.
func (d *Display) Init() {
	d.Reset()

	d.sendCommand(displayRefresh)
	d.waitUntilIdle()

	d.sendCommand(autoWriteRamRed, 0xF7)
	d.waitUntilIdle()
	d.sendCommand(autoWriteRamBW, 0xF7)
	d.waitUntilIdle()

	d.sendCommand(softStart, 0xAE, 0xC7, 0xC3, 0xC0, 0x40)

	// set MUX as 527
	d.sendCommand(setGateDriver, 0xAF, 0x02, 0x01)

	d.sendCommand(dataEntryMode, 0x01)

	// RAM x address starts at 0
	// RAM x address ends at 36Fh -> 879
	d.sendCommand(setRamXStart, 0x00, 0x00, 0x6F, 0x03)
	// RAM y address starts at 20Fh
	// RAM y address ends at 00h
	d.sendCommand(setRamYStart, 0xAF, 0x02, 0x00, 0x00)

	// VBD, LUT1 for white.
	d.sendCommand(borderWaveformControl, 0x01)

	d.sendCommand(tempSensorControl, 0x80)
	//Load Temperature and waveform setting.
	d.sendCommand(displayUpdateControl2, 0xB1)
	d.sendCommand(masterActivation)
	d.waitUntilIdle()

	d.sendCommand(setRamXAddressCtr, 0x00, 0x00)
	d.sendCommand(setRamYAddressCtr, 0xAF, 0x02)
}

// Clear clears the screen.
func (d *Display) Clear() {
	d.Render(nil, nil)
}

// Render updates the screen from the provided io.ByteReaders.
//
// The epd7in5bhd does not support partial refreshes. If the provided buffer is
// smaller than the image, then the rest will be filled with white.
//
// The epd7in5bhd expects a bit per pixel for each color.
//
// For blackImg, 0b1 is a black pixel, and 0b0 is a white pixel. For redImg,
// 0b1 is a red pixel, and 0b0 is a not-red pixel (no change will occur).
//
// Black will always be drawn on the screen before red.
func (d *Display) Render(blackImg, redImg []byte) {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("Render: %s", time.Since(start).String())
	}(now)
	d.sendCommand(setRamYAddressCtr, 0xAF, 0x02)

	// 1 is white, 0 is black.
	blackPad := bytes.Repeat([]byte{0xFF}, BufSize-len(blackImg))
	d.sendCommand(writeRAMBW, append(blackImg, blackPad...)...)

	// 0 is white or black, 1 is red.
	redPad := bytes.Repeat([]byte{0x00}, BufSize-len(redImg))
	d.sendCommand(writeRAMRed, append(redImg, redPad...)...)
	d.turnOnDisplay()
}

// RenderPaletted renders an image in 3 colors (black, white and red/yellow).
//
// If img is a *image.Paletted with exactly 3 colors, each color will be assigned to its
// nearest by euclidean distance. Otherwise, colors will be assigned by a per-pixel calculation.
func (d *Display) RenderPaletted(img image.Image) {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("RenderPaletted: %s", time.Since(start).String())
	}(now)
	bbuf := bytes.NewBuffer(make([]byte, 0, BufSize))
	rbuf := bytes.NewBuffer(make([]byte, 0, BufSize))
	Encode(bbuf, rbuf, img)

	d.sendCommand(setRamYAddressCtr, 0xAF, 0x02)
	d.sendCommand(writeRAMBW, bbuf.Bytes()...)

	d.sendCommand(writeRAMRed, rbuf.Bytes()...)
	d.turnOnDisplay()
}

// Sleep tells the Display to enter deepSleepMode.
//
// The display can be reawakened with Reset(), and re-initialized with Init().
func (d *Display) Sleep() {
	d.sendCommand(deepSleepMode, 0x01) //deep sleep
}

//// Convert converts the input image into a byte buffer suitable for Display.Render.
//func Convert(img image.Image) []byte {
//	return convert(img, false)
//}
//
func convert(img image.Image, invert bool) []byte {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("Convert: %s", time.Since(start).String())
	}(now)
	buffer := make([]byte, BufSize, BufSize)
	p := color.Palette([]color.Color{color.Black, color.White})
	if invert {
		p = []color.Color{color.White, color.Black}
	}
	w := p.Index(color.White)
	for y := 0; y < DisplayHeight; y++ {
		row := y * DisplayWidthBytes
		for x := 0; x < DisplayWidth; x++ {
			c := w
			if (image.Point{x, y}).In(img.Bounds()) {
				c = p.Index(img.At(x, y))
			}
			px := (x / 8) + row
			bit := byte(0x80 >> (uint32(x) % 8))
			if c == 0 {
				buffer[px] &= ^bit
			} else {
				buffer[px] |= bit
			}
		}
	}

	return buffer
}

func Encode(dstBlack, dstRed io.ByteWriter, img image.Image) {
	if pi, ok := img.(*image.Paletted); ok && len(pi.Palette) == 3 {
		encodeExactColors(dstBlack, dstRed, pi)
		return
	}
	p := color.Palette([]color.Color{color.White, color.Black, color.RGBA{255, 0, 0, 255}})
	white, black, highlight := 0, 1, 2
	var rbyte, bbyte byte
	pt := image.Point{}
	bounds := img.Bounds()
	for y := 0; y < DisplayHeight; y++ {
		pt.Y = y
		for x := 0; x < DisplayWidth; x++ {
			pt.X = x
			var c int
			if (pt).In(bounds) {
				c = p.Index(img.At(x, y))
			}
			bit := byte(0x80 >> (uint32(x) % 8))
			switch c {
			case highlight:
				bbyte |= bit
				rbyte |= bit
			case black:
				bbyte &= ^bit
				rbyte &= ^bit
			case white:
				bbyte |= bit
				rbyte &= ^bit
			}
			if (x % 8) == 0 {
				dstBlack.WriteByte(bbyte)
				dstRed.WriteByte(rbyte)
				rbyte, bbyte = 0x00, 0x00
			}
		}
	}
}

func exactColorIndex(img *image.Paletted) (white, black, highlight int) {
	// This order is significant. We want to try to assign white and black before our third color,
	// as they may be closer to a totally non-red color (blue).
	colors := []color.Color{color.White, color.Black, color.RGBA{255, 0, 0, 255}}
	p := color.Palette{}
	ip := make(color.Palette, len(img.Palette))
	copy(ip, img.Palette)
	// Sort Palette p:
	// img.Palette lightest, img.Palette darkest, img.Palette remaining
	// Iterate over colors, popping as we go to avoid duplicates.
	// We don't want both faint red and white to be white.
	for _, c := range colors {
		ci := ip.Index(c)
		p = append(p, ip[ci])
		ip = append(ip[:ci], ip[ci+1:]...)
	}
	// Now, map our expected order to img.Paletted.Palette's order
	return img.Palette.Index(p[0]), img.Palette.Index(p[1]), img.Palette.Index(p[2])
}

func encodeExactColors(dstBlack, dstRed io.ByteWriter, img *image.Paletted) {
	white, black, highlight := exactColorIndex(img)
	var rbyte, bbyte byte
	pt := image.Point{}
	bounds := img.Bounds()
	for y := 0; y < DisplayHeight; y++ {
		pt.Y = y
		for x := 0; x < DisplayWidth; x++ {
			pt.X = x
			var c int
			if (pt).In(bounds) {
				c = int(img.ColorIndexAt(x, y))
			}
			bit := byte(0x80 >> (uint32(x) % 8))
			switch c {
			case highlight:
				bbyte |= bit
				rbyte |= bit
			case black:
				bbyte &= ^bit
				rbyte &= ^bit
			case white:
				bbyte |= bit
				rbyte &= ^bit
			}
			if (x % 8) == 0 {
				dstBlack.WriteByte(bbyte)
				dstRed.WriteByte(rbyte)
				rbyte, bbyte = 0x00, 0x00
			}
		}
	}
}

// RenderImages renders a black image and a red/yellow image on the display.
func (d *Display) RenderImages(black, redyellow image.Image) {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("RenderImages: %s", time.Since(start).String())
	}(now)
	d.Render(convert(black, false), convert(redyellow, true))
}
