package watchman

import (
	"io/ioutil"
	"testing"
)

func TestClock(t *testing.T) {
	cl := &Client{Sockname: sock}

	t.Run("empty path", func(t *testing.T) {
		_, err := cl.Clock("")
		expectErrEqual(t, err, "invalid argument")
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := cl.Clock("/stupid dumb i guess this could exist")
		expectErrRegex(t, err, "^lstat .*: no such file or directory$")
	})

	t.Run("no watch", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		_, err = cl.Clock(path)
		expectErrRegex(t, err, "^unable to resolve root .*: directory .* is not watched$")
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
