package watchman

// WatchList is the return object of the WatchList call
type WatchList struct {
	Roots []string
}

// WatchList returns a list of watched dirs
// https://facebook.github.io/watchman/docs/cmd/watch-list.html
func (c *Client) WatchList() (*WatchList, error) {
	var data struct {
		WatchList
		base
	}

	if err := c.Send(&data, "watch-list"); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.WatchList, nil
}
