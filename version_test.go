package watchman

import "testing"

func TestVersion(t *testing.T) {
	cl := &Client{Sockname: sock}
	v, err := cl.Version()
	if err != nil {
		t.Errorf("Error %s", err)
	}

	if v == nil || v.Version == "" {
		t.Error("Version is empty")
	}
}
