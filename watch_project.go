package watchman

// WatchProject is the return object of the WatchProject call
type WatchProject struct {
	Version string
	Watch   string
	Watcher string
}

// WatchProject requests that the project containing the requested dir is watched for changes
func (c *Client) WatchProject(path string) (*WatchProject, error) {
	var data struct {
		WatchProject
		base
	}

	if err := c.Send(&data, "watch-project", path); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.WatchProject, nil
}
