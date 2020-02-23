package watchman

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

var (
	sock string
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	d, err := ioutil.TempDir("", "watchman")
	if err != nil {
		fmt.Printf("Error creating tempdir %s\n", err)
		os.Exit(1)
	}

	var stderr bytes.Buffer
	sock = filepath.Join(d, "sock")

	cmd := exec.Command("watchman", "--foreground", "--no-save-state", "--logfile="+filepath.Join(d, "log"), "--pidfile="+filepath.Join(d, "pid"), "--sockname="+sock)
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting watchman %s\n", err)
		os.Exit(1)
	}

	ch := make(chan error)
	go func() {
		ch <- cmd.Wait()
	}()

	time.Sleep(500 * time.Millisecond)
	select {
	case <-ch:
		fmt.Printf("Error running watchman %s\n", stderr.String())
		os.Exit(1)
	default:
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("Error stopping watchman %s\n", err)
			os.Exit(1)
		}
	}()

	return m.Run()
}

func expectErrEqual(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error")
	}

	if err.Error() != msg {
		t.Fatalf("Expected error message to be \"%s\" but found \"%s\"", msg, err.Error())
	}
}

func expectErrRegex(t *testing.T, err error, pattern string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error")
	}

	if !regexp.MustCompile(pattern).MatchString(err.Error()) {
		t.Fatalf("Expected error to match \"%s\" but found \"%s\"", pattern, err.Error())
	}
}
