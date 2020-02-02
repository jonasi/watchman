package watchman

// Version is the return object of the Version call
type Version struct {
	Version string
}

// Version will tell you the version and build information for the currently running watchman service
func (c *Client) Version() (*Version, error) {
	var data base

	if err := c.Send(&data, "version"); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &Version{Version: data.Version}, nil
}
