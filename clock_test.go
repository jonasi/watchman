package watchman

import (
	"io/ioutil"
	"testing"
)

func TestClock(t *testing.T) {
	cl := &Client{Sockname: sock}

	t.Run("empty path", func(t *testing.T) {
		_, err := cl.Clock("")
		if err == nil {
			// todo(isao) - specific error
			t.Errorf("Expected error")
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := cl.Clock("/stupid dumb i guess this could exist")
		if err == nil {
			// todo(isao) - specific error
			t.Errorf("Expected error")
		}
	})

	t.Run("no watch", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		_, err = cl.Clock(path)
		if err == nil {
			// todo(isao) - specific error
			t.Errorf("Expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		w, err := cl.Watch(path)
		if err != nil {
			t.Fatalf("Error watching path %s: %s", path, err)
		}

		defer cl.WatchDel(w.Watch)

		c, err := cl.Clock(w.Watch)
		if err != nil {
			t.Errorf("Unexpected error calling clock: %s", err)
		}

		if c == nil || c.Clock == "" {
			t.Errorf("Clock is empty: %#v", c)
		}
	})
}
