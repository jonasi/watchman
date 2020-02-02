package watchman

import (
	"github.com/yookoala/realpath"
)

// Watch is the return object of the Watch call
type Watch struct {
	Watch   string
	Watcher string
}

// Watch requests that the specified dir is watched for changes
func (c *Client) Watch(path string) (*Watch, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		Watch
		base
	}

	if err := c.Send(&data, "watch", path); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.Watch, nil
}
