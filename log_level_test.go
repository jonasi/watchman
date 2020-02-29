package watchman

import "testing"

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
