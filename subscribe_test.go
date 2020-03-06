package watchman

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

func TestSubscribe(t *testing.T) {
	cl := &Client{Sockname: sock}

	path, err := ioutil.TempDir("", "watchmantest")
	if err != nil {
		t.Fatalf("Error creating temp dir %s", err)
	}

	ch := make(chan *SubscribeEvent)
	s, stop, err := cl.Subscribe(path, "testone!", map[string]interface{}{}, ch)
	if err != nil {
		t.Fatalf("error subscribing %s", err)
	}
	defer stop()

	if s.Clock == "" || s.Subscribe != "testone!" {
		t.Fatalf("Invalid subscribe object: %#v", s)
	}

	ioutil.WriteFile(filepath.Join(path, "test1"), []byte("OK"), 0755)
	select {
	case ev := <-ch:
		if !(len(ev.Files) == 1 && ev.Files[0].New && ev.Files[0].Exists && ev.Files[0].Name == "test1") {
			t.Fatalf("Expected one new file, found %#v", ev.Files)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected event after writing file, but none came")
	}

	_, err = cl.Unsubscribe(path, "testone!")
	if err != nil {
		t.Fatalf("Unexpected error unsubscribing: %s", err)
	}

	ioutil.WriteFile(filepath.Join(path, "test2"), []byte("OK"), 0755)
	select {
	case ev := <-ch:
		t.Fatalf("Unexpected event after writing file: %#v", ev)
	case <-time.After(200 * time.Millisecond):
	}
}
