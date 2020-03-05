package bser

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

// infiniteRW is a io.ReadWriter that never returns
type infiniteRW struct{ readCh chan struct{} }

func (r infiniteRW) Read(b []byte) (int, error) {
	r.readCh <- struct{}{}
	select {}
}

func (infiniteRW) Write(b []byte) (int, error) {
	select {}
}

func TestUntapLongRead(t *testing.T) {
	rw := infiniteRW{readCh: make(chan struct{})}

	tap := NewTap(rw, nil, nil)

	go tap.Read(nil)
	// wait for read to have been called
	<-rw.readCh

	untapped := make(chan struct{})
	go func() {
		tap.Untap()
		untapped <- struct{}{}
	}()

	select {
	case <-untapped:
	case <-time.After(100 * time.Millisecond):
		t.Error("Untap did not return after 100 ms")
	}
}

func TestTap(t *testing.T) {
	var (
		buf  bytes.Buffer
		rDst = make(chan []byte)
		wDst = make(chan []byte)
	)

	tap := NewTap(&buf, func(b []byte) {
		rDst <- b
	}, func(b []byte) {
		wDst <- b
	})

	// write some initial data to the tap
	pdu, err := MarshalPDU(42)
	if err != nil {
		t.Fatalf("unexpected error marshaling value: %s", err)
	}
	if _, err := tap.Write(pdu); err != nil {
		t.Fatalf("unexpected error writing marshaled value: %s", err)
	}

	// int 42
	expectWritten := []byte{3, 42}
	written := <-wDst

	// verify initial data written to dst
	if !reflect.DeepEqual(written, expectWritten) {
		t.Errorf("unexpected result from wfn\nexpected=%v\nactual=%v", expectWritten, written)
	}

	// read data from the tap
	_, err = ioutil.ReadAll(tap)
	if err != nil {
		t.Fatalf("unexpected error reading from tap: %s", err)
	}

	// int 42
	expectRead := []byte{3, 42}
	read := <-rDst

	// verify decoded data written to dst
	if !reflect.DeepEqual(read, expectRead) {
		t.Errorf("unexpected result from wfn\nexpected=%v\nactual=%v", expectRead, read)
	}

	// write more data to the tap
	pdu, err = MarshalPDU("foo")
	if err != nil {
		t.Fatalf("unexpected error marshaling value: %s", err)
	}
	if _, err := tap.Write(pdu); err != nil {
		t.Fatalf("unexpected error writing marshaled value: %s", err)
	}

	// string of length 3 "foo"
	expectWritten = []byte("\x02\x03\x03foo")
	written = <-wDst

	// verify initial data written to dst
	if !reflect.DeepEqual(written, expectWritten) {
		t.Errorf("unexpected result from wfn\nexpected=%v\nactual=%v", expectWritten, written)
	}

	// read data from the tap
	_, err = ioutil.ReadAll(tap)
	if err != nil {
		t.Fatalf("unexpected error reading from tap: %s", err)
	}

	// close the tap
	tap.Untap()

	// string of length 3 "foo"
	expectRead = []byte("\x02\x03\x03foo")
	read = <-rDst

	// verify decoded data written to dst
	if !reflect.DeepEqual(read, expectRead) {
		t.Errorf("unexpected result from wfn\nexpected=%v\nactual=%v", expectRead, read)
	}
}
