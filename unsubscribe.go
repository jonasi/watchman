package watchman

import "github.com/yookoala/realpath"

// Unsubscribe is the return object of the Subscribe call
type Unsubscribe struct {
	Deleted     bool   `bser:"deleted"`
	Unsubscribe string `bser:"unsubscribe"`
}

// Unsubscribe cancels a named subscription against the specified root. The server side will no longer generate subscription packets for the specified subscription.
// https://facebook.github.io/watchman/docs/cmd/unsubscribe.html
func (c *Client) Unsubscribe(path, name string) (*Unsubscribe, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		base
		Unsubscribe
	}

	if err := c.Send(&data, "unsubscribe", path, name); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.Unsubscribe, nil
}
