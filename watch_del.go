package watchman

import (
	"github.com/yookoala/realpath"
)

// WatchDel is the return object of the WatchDel call
type WatchDel struct {
	Root     string `bser:"root"`
	WatchDel bool   `bser:"watch-del"`
}

// WatchDel removes a watch and any associated triggers
func (c *Client) WatchDel(path string) (*WatchDel, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		WatchDel
		base
	}

	if err := c.Send(&data, "watch-del", path); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.WatchDel, nil
}
