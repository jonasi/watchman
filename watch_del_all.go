package watchman

import (
	"github.com/yookoala/realpath"
)

// WatchDelAll is the return object of the WatchDelAll call
type WatchDelAll struct {
	Roots []string `bser:"roots"`
}

// WatchDelAll removes all watches and associated triggers
func (c *Client) WatchDelAll(path string) (*WatchDelAll, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		WatchDelAll
		base
	}

	if err := c.Send(&data, "watch-del-all", path); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.WatchDelAll, nil
}
