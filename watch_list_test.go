package watchman

import "testing"

func TestWatchList(t *testing.T) {
	cl := &Client{Sockname: sock}
	_, err := cl.WatchList()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
}
