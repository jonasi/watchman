package watchman

import (
	"io/ioutil"
	"testing"
)

func TestFind(t *testing.T) {
	cl := &Client{Sockname: sock}

	t.Run("empty path", func(t *testing.T) {
		_, err := cl.Find("")
		expectErrEqual(t, err, "invalid argument")
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := cl.Find("/stupid dumb i guess this could exist")
		expectErrRegex(t, err, "^lstat .*: no such file or directory$")
	})

	t.Run("not watched", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		_, err = cl.Find(path)
		expectErrRegex(t, err, "^unable to resolve root .*: directory .* is not watched$")
	})

	t.Run("success - no patterns", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		w, err := cl.Watch(path)
		if err != nil {
			t.Fatalf("Error watching path %s: %s", path, err)
		}

		defer cl.WatchDel(w.Watch)

		files, err := cl.Find(path)
		if err != nil {
			t.Errorf("Unexpected error calling find: %s", err)
		}

		if files.Clock == "" {
			t.Errorf("Expected non-empty Clock field")
		}

		if len(files.Files) > 0 {
			t.Errorf("Expected no files, found %d", len(files.Files))
		}
	})

	t.Run("success", func(t *testing.T) {
		path, err := ioutil.TempDir("", "watchmantest")
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		err = ioutil.WriteFile(path+"/hey.txt", []byte("hey"), 0700)
		if err != nil {
			t.Fatalf("Error creating temp dir %s", err)
		}

		w, err := cl.Watch(path)
		if err != nil {
			t.Fatalf("Error watching path %s: %s", path, err)
		}

		defer cl.WatchDel(w.Watch)

		files, err := cl.Find(path, "*.txt")
		if err != nil {
			t.Fatalf("Unexpected error calling find: %s", err)
		}

		if len(files.Files) != 1 {
			t.Errorf("Expected to find one file, found %d", len(files.Files))
		}
	})
}
