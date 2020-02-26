package bser

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

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
