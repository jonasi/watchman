package bser

import (
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
