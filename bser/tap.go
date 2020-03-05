package bser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
)

// NewTap sends all PDUs sent in either direction along rw
// through rfn for reads and wfn for writes
func NewTap(rw io.ReadWriter, rfn func([]byte), wfn func([]byte)) *Tap {
	t := &Tap{rw: rw, rfn: rfn, wfn: wfn}
	t.Tap()

	return t
}

// Tap is a io.ReadWriter that will pass data through some additional functions
type Tap struct {
	mu       sync.RWMutex
	rw       io.ReadWriter
	rfn      func([]byte)
	wfn      func([]byte)
	r        io.Reader
	w        io.Writer
	tapped   bool
	cleanups []func()
}

// Tap enables the functionality
func (t *Tap) Tap() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.tapped {
		return
	}

	t.tapped = true

	t.r = t.rw
	t.w = t.rw
	t.cleanups = []func(){}

	if t.rfn != nil {
		lw, cl := t.logWriter(t.rfn)
		t.r = io.TeeReader(t.rw, lw)
		t.cleanups = append(t.cleanups, cl)
	}
	if t.wfn != nil {
		lw, cl := t.logWriter(t.wfn)
		t.w = io.MultiWriter(t.rw, lw)
		t.cleanups = append(t.cleanups, cl)
	}
}

// Untap disables the functionality
func (t *Tap) Untap() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.tapped {
		return
	}

	// reset
	t.tapped = false
	t.r = t.rw
	t.w = t.rw

	for _, fn := range t.cleanups {
		fn()
	}

	t.cleanups = nil
}

func (t *Tap) Read(b []byte) (int, error) {
	t.mu.RLock()
	r := t.r
	t.mu.RUnlock()

	return r.Read(b)
}

func (t *Tap) Write(b []byte) (int, error) {
	t.mu.RLock()
	w := t.w
	t.mu.RUnlock()

	return w.Write(b)
}

func (t *Tap) logWriter(fn func([]byte)) (io.Writer, func()) {
	pr, pw := io.Pipe()
	go func() {
		for {
			buf, err := readPDU(pr)
			if err != nil {
				return
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Fprintf(os.Stderr, "Tap Panic Recovery: %s\n%s\n", r, debug.Stack())
					}
				}()

				fn(buf)
			}()
		}
	}()

	return pw, func() {
		pr.Close()
	}
}

func readPDU(r io.Reader) ([]byte, error) {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	if !bytes.Equal(buf, protocolPrefix) {
		return nil, fmt.Errorf("Expected %x, found %x", protocolPrefix, buf)
	}

	var size int
	if err := decodeValue(r, reflect.ValueOf(&size), nil); err != nil {
		return nil, err
	}

	buf = make([]byte, size)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}
