package watchman

import (
	"github.com/yookoala/realpath"
)

// Clock is the return object of the Clock call
type Clock struct {
	Clock string `bser:"clock"`
}

// Clock returns the current clock value for a watched root.
func (c *Client) Clock(path string) (*Clock, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		base
		Clock
	}

	if err := c.Send(&data, "clock", path); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.Clock, nil
}
