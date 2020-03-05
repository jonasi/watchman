package watchman

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

var logLevelTests = map[string]struct {
	expectedLevel string
	expectErr     bool
}{
	LogLevelDebug: {
		expectedLevel: LogLevelDebug,
		expectErr:     false,
	},
	LogLevelError: {
		expectedLevel: LogLevelError,
		expectErr:     false,
	},
	LogLevelOff: {
		expectedLevel: LogLevelOff,
		expectErr:     false,
	},
	"invalid_log_level": {
		expectErr: true,
	},
}

func TestLogLevel(t *testing.T) {
	for logLevel, testCase := range logLevelTests {
		t.Run(logLevel, func(t *testing.T) {
			cl := &Client{Sockname: sock}
			l, err := cl.LogLevel(logLevel)
			if err != nil {
				if !testCase.expectErr {
					t.Fatalf("unexpected error setting log level: %s", err)
				}
				return
			}

			if testCase.expectErr {
				t.Error("unexpectedly no error")
			}

			if l.LogLevel != testCase.expectedLevel {
				t.Errorf("unexpected log-level (expected = '%s', actual = '%s')", testCase.expectedLevel, l.LogLevel)
			}
		})
	}
}

func TestLogLevelListen(t *testing.T) {
	cl := &Client{Sockname: sock}
	ch := make(chan interface{})
	stop, err := cl.Receive(ch)
	if err != nil {
		t.Fatalf("Error calling Receive: %s", err)
	}
	defer stop()

	if _, err := cl.LogLevel(LogLevelDebug); err != nil {
		t.Fatalf("Error calling LogLevel: %s", err)
	}

	go func() {
		err := exec.Command("watchman", "--sockname="+sock, "log", "debug", "GOOD ONE").Run()
		if err != nil {
			t.Fatalf("Error calling Log: %s", err)
		}
	}()

	found := make(chan bool)

	go func() {
		for m := range ch {
			if l, ok := m.(*LogEvent); ok && l.Level == LogLevelDebug && strings.Contains(l.Log, `GOOD ONE`) {
				found <- true
				return
			}
		}
	}()

	select {
	case <-time.After(time.Second):
		t.Fatal("Expected to see the sent log message but it was never received")
	case <-found:
	}
}
