package bser

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

// Tap sends all PDUs sent in either direction along rw
// through rfn for reads and wfn for writes
func Tap(rw io.ReadWriter, rfn func([]byte), wfn func([]byte)) io.ReadWriter {
	var r io.Reader = rw
	var w io.Writer = rw

	if rfn != nil {
		r = io.TeeReader(rw, logWriter(rfn))
	}
	if wfn != nil {
		w = io.MultiWriter(rw, logWriter(wfn))
	}

	return tap{r, w}
}

type tap struct {
	io.Reader
	io.Writer
}

func logWriter(fn func([]byte)) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		for {
			buf, err := readPDU(pr)
			if err != nil {
				return
			}

			fn(buf)
		}
	}()

	return pw
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
