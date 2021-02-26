package epd7in5bhd

import (
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

type hardware struct {
	txLimit int

	mut sync.Mutex
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

func (h *hardware) DataWriter() io.Writer {
	return &batchedWriter{&dataWriter{h}, h.txLimit}
}

func (h *hardware) CommandWriter() io.Writer {
	return &commandWriter{h}
}

type dataWriter struct {
	*hardware
}

func (w *dataWriter) Write(p []byte) (n int, err error) {
	w.mut.Lock()
	defer w.mut.Unlock()
	if len(p) == 0 {
		return 0, nil
	}
	if err := w.cs.Out(gpio.Low); err != nil {
		return 0, fmt.Errorf("%v.Out(%v) = %w", w.cs.String(), gpio.Low.String(), err)
	}
	if err := w.dc.Out(gpio.High); err != nil {
		return 0, fmt.Errorf("%v.Out(%v) = %w", w.dc.String(), gpio.High.String(), err)
	}
	defer func() {
		if e := w.cs.Out(gpio.High); e != nil {
			err = fmt.Errorf("already had err %q, and got e: %w", err, e)
		}
	}()
	if w.txLimit <= 0 {
		return 0, io.ErrShortWrite
	}
	if len(p) > w.txLimit {
		if err := w.c.Tx(p[:w.txLimit], nil); err != nil {
			return w.txLimit, err
		}
		return w.txLimit, io.ErrShortWrite
	}
	if err := w.c.Tx(p, nil); err != nil {
		return len(p), err
	}
	return len(p), nil
}

type commandWriter struct {
	*hardware
}

func (w *commandWriter) writeCommand(p byte) (err error) {
	w.mut.Lock()
	defer w.mut.Unlock()
	if err := w.dc.Out(gpio.Low); err != nil {
		return fmt.Errorf("%v.Out(%v) = %w", w.dc.String(), gpio.Low.String(), err)
	}
	if err := w.cs.Out(gpio.Low); err != nil {
		return fmt.Errorf("%v.Out(%v) = %w", w.cs.String(), gpio.Low.String(), err)
	}
	defer func() {
		if err2 := w.cs.Out(gpio.High); err2 != nil {
			err = fmt.Errorf("%v.Out(%v) = %w, already had error %v", w.cs.String(), gpio.High, err2, err)
		}
	}()
	if err := w.c.Tx([]byte{p}, nil); err != nil {
		return fmt.Errorf("sending command %s: %w", command(p).String(), err)
	}
	return nil
}

func (w *commandWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	cmd, data := p[0], p[1:]
	if err := w.writeCommand(cmd); err != nil {
		return 1, err
	}
	if len(data) == 0 {
		return 1, nil
	}
	n, err := w.DataWriter().Write(data)
	return 1 + n, err
}

type batchedWriter struct {
	dst       io.Writer
	batchSize int
}

func (b *batchedWriter) Write(p []byte) (int, error) {
	var sent int
	for i := 0; i < len(p); i += b.batchSize {
		j := i + b.batchSize
		if j > len(p) {
			j = len(p)
		}
		n, err := b.dst.Write(p[i:j])
		if err != nil {
			return n + sent, err
		}
		n += sent
	}
	return sent, nil
}
