package watchman

// the supported log levels
const (
	LogLevelDebug = "debug"
	LogLevelError = "error"
	LogLevelOff   = "off"
)

// Level is the return object of SetLogLevel
type Level struct {
	LogLevel string `bser:"log_level"`
}

// SetLogLevel changes the log level of the connection
func (c *Client) SetLogLevel(level string) (*Level, error) {
	var data struct {
		base
		Level
	}

	if err := c.Send(&data, "log-level", level); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.Level, nil
}
