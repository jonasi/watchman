package watchman

// the supported log levels
const (
	LogLevelDebug = "debug"
	LogLevelError = "error"
	LogLevelOff   = "off"
)

// LogLevel is the return object of SetLogLevel
type LogLevel struct {
	LogLevel string `bser:"log_level"`
}

// LogLevel changes the log level of the connection
// https://facebook.github.io/watchman/docs/cmd/log-level.html
func (c *Client) LogLevel(level string) (*LogLevel, error) {
	var data struct {
		base
		LogLevel
	}

	if err := c.Send(&data, "log-level", level); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.LogLevel, nil
}
