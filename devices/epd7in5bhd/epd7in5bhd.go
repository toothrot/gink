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
	"time"

	"periph.io/x/periph/conn"
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
	// c is a perhiph conn.Conn.
	c conn.Conn

	// busy pin, when waiting for device to be ready.
	busy gpio.PinIO
	// cs is the Chip Enable pin.
	cs gpio.PinOut
	// dc is the data/command pin.
	dc gpio.PinOut
	// rst is the CE1 pin.
	rst gpio.PinOut
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
var DefaultWait = 27 * time.Second

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
	c, err := port.Connect(5*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		connerr := fmt.Errorf("port.Connect(%v, %v, %v) = %w", 5*physic.MegaHertz, spi.Mode0, 8, err)
		if err := port.Close(); err != nil {
			return nil, fmt.Errorf("port.Close() = %w while handling %q", err, connerr)
		}
		return nil, connerr
	}

	e := &Display{
		c:    c,
		dc:   dc,
		cs:   cs,
		rst:  rst,
		busy: busy,
	}
	return e, nil
}

// Reset can be also used to awaken the device.
func (d *Display) Reset() {
	d.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
	d.rst.Out(gpio.Low)
	time.Sleep(2 * time.Millisecond)
	d.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
}

func (d *Display) sendCommand(cmd byte, data ...byte) {
	d.dc.Out(gpio.Low)
	d.cs.Out(gpio.Low)
	d.c.Tx([]byte{cmd}, nil)
	d.cs.Out(gpio.High)
	if len(data) != 0 {
		for _, b := range data {
			d.sendData(b)
		}
	}
}

func (d *Display) sendData(data byte) {
	d.cs.Out(gpio.Low)
	d.dc.Out(gpio.High)
	d.c.Tx([]byte{data}, nil)
	d.cs.Out(gpio.High)
}

func (d *Display) waitUntilIdle() {
	for d.busy.Read() == gpio.Low {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
}

func (d *Display) turnOnDisplay() {
	// Load LUT from MCU(0x32)
	d.sendCommand(displayUpdateControl2, 0xC7)
	d.sendCommand(masterActivation)
	time.Sleep(200 * time.Millisecond) //!!!The delay here is necessary, 200uS at least!!!
	d.waitUntilIdle()                  //waiting for the electronic paper IC to release the idle signal
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

// FullBuffer returns a buffer of BufSize populated with repeated val.
func FullBuffer(val byte) []byte {
	b := make([]byte, BufSize, BufSize)
	for i := range b {
		b[i] = val
	}
	return b
}

// Display updates the screen from the provided io.ByteReaders..
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
func (d *Display) Render(blackImg, redImg io.ByteReader) {
	if blackImg == nil {
		blackImg = bytes.NewReader([]byte{})
	}
	if redImg == nil {
		redImg = bytes.NewReader([]byte{})
	}
	d.sendCommand(setRamYAddressCtr, 0xAF, 0x02)

	d.sendCommand(writeRAMBW)
	for i := 0; i < BufSize; i++ {
		b, err := blackImg.ReadByte()
		if err != nil {
			// io.ByteReader never returns a valid byte on any error, including io.EOF.
			b = 0xFF
		}
		d.sendData(b)
	}

	d.sendCommand(writeRAMRed)
	for i := 0; i < BufSize; i++ {
		b, err := redImg.ReadByte()
		if err != nil {
			// io.ByteReader never returns a valid byte on any error, including io.EOF.
			b = 0xFF
		}
		d.sendData(^b)
	}
	d.turnOnDisplay()
}

// Sleep tells the Display to enter deepSleepMode.
//
// The display can be reawakened with Reset(), and re-initialized with Init().
func (d *Display) Sleep() {
	d.sendCommand(deepSleepMode, 0x01) //deep sleep
}

// Convert converts the input image into a byte buffer suitable for Display.Render.
func Convert(img image.Image) []byte {
	buffer := make([]byte, BufSize, BufSize)
	for j := 0; j < DisplayHeight; j++ {
		for i := 0; i < DisplayWidth; i++ {
			pixel := 1
			if i < img.Bounds().Dx() && j < img.Bounds().Dy() {
				pixel = color.Palette([]color.Color{color.Black, color.White}).Index(img.At(i, j))
			}
			if pixel == 1 {
				buffer[(i/8)+(j*DisplayWidthBytes)] |= 0x80 >> (uint32(i) % 8)
			}
		}
	}

	return buffer
}
