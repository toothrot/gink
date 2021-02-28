// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package epd7in5bhd is for the Waveshare 7.5 inch HD (B/C) e-Paper display.
package epd7in5bhd

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"time"

	"golang.org/x/image/draw"
	"periph.io/x/periph/conn/gpio"
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

var (
	DisplayBounds = image.Rect(0, 0, DisplayWidth, DisplayHeight)
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
	buffer *Image
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
	hw, err := newHardware(p)
	if err != nil {
		return nil, err
	}
	return &Display{
		hw: hw,
		buffer: NewImage(DisplayBounds),
	}, nil
}

// Reset clears all variables set on the Display.
//
// Reset can be also used to awaken the device after a call to Sleep.
func (d *Display) Reset() {
	d.hw.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
	d.hw.rst.Out(gpio.Low)
	time.Sleep(2 * time.Millisecond)
	d.hw.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
}

func (d *Display) sendCommand(cmd command, data ...byte) {
	n, err := d.hw.CommandWriter().Write(append([]byte{byte(cmd)}, data...))
	if err != nil {
		log.Printf("sendCommand Write() = %d, %v", n, err)
	}
}

// waitUntilIdle waits for the busy pin to be low voltage. It's required after some commands, and should not be
// called unless necessary.
func (d *Display)  waitUntilIdle() {
	for d.hw.busy.Read() == gpio.Low {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(10 * time.Millisecond)
}

// As far as I can tell this actually triggers a draw.
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
	d.buffer.Reset()
	d.Refresh()
}

// Upload updates the screen from the provided io.ByteReaders.
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
func (d *Display) Upload(blackImg, redImg []byte) {
	d.sendCommand(setRamYAddressCtr, 0xAF, 0x02)

	// 1 is white, 0 is black.
	blackPad := bytes.Repeat([]byte{0xFF}, BufSize-len(blackImg))
	d.sendCommand(writeRAMBW, append(blackImg, blackPad...)...)

	// 0 is white or black, 1 is red.
	redPad := bytes.Repeat([]byte{0x00}, BufSize-len(redImg))
	d.sendCommand(writeRAMRed, append(redImg, redPad...)...)
	d.turnOnDisplay()
}

// Refresh uploads the buffer to the display.
func (d *Display) Refresh() {
	d.Upload(d.buffer.Black, d.buffer.Highlight)
}

// DrawAndRefresh is a convenience method for Draw and Refresh.
func (d *Display) DrawAndRefresh(img image.Image) {
	d.Draw(img)
	d.Refresh()
}

// DrawAndRefresh draws an image to the display buffer in 3 colors (black, white and red/yellow).
//
// If img is a *image.Paletted with exactly 3 colors, each color will be assigned to its
// nearest by euclidean distance. Otherwise, colors will be assigned by a per-pixel calculation.
func (d *Display) Draw(img image.Image) {
	if pi, ok := img.(*image.Paletted); ok && len(pi.Palette) == 3 {
		d.buffer.drawExactColors(pi)
		return
	}
	draw.Draw(d.buffer, d.buffer.Bounds(), img, image.Point{0, 0}, draw.Src)
}

// Sleep tells the Display to enter deepSleepMode.
//
// The display can be reawakened with Reset(), and re-initialized with Init().
func (d *Display) Sleep() {
	d.sendCommand(deepSleepMode, 0x01) //deep sleep
}

// Convert converts the input image into a byte buffer suitable for Display.Upload.
func convert(img image.Image, p color.Palette) *Image {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("Convert: %s", time.Since(start).String())
	}(now)
	dst := NewImage(DisplayBounds)
	dst.Palette = p
	draw.Draw(dst, dst.Bounds(), img, image.Point{0, 0}, draw.Src)
	return dst
}

// DrawAndRefreshImages renders a black image and a red/yellow image on the display.
func (d *Display) DrawAndRefreshImages(black, redyellow image.Image) {
	now := time.Now()
	defer func(start time.Time) {
		log.Printf("DrawAndRefreshImages: %s", time.Since(start).String())
	}(now)
	bi, hi := convert(black, color.Palette{White, Black}), convert(redyellow, color.Palette{White, Highlight})
	d.buffer.Black = bi.Black
	d.buffer.Highlight = hi.Highlight
	d.Refresh()
}
