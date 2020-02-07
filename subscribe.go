package watchman

import (
	"github.com/yookoala/realpath"
)

// Subscribe is the return object of the Subscribe call
type Subscribe struct {
	Clock     string `bser:"clock"`
	Subscribe string `bser:"subscribe"`
}

// SubscribeEvent is the unilateral message that indicates an
// fs event occurred for the specified subscription
type SubscribeEvent struct {
	Clock           string          `bser:"clock"`
	Files           []SubscribeFile `bser:"files"`
	IsFreshInstance bool            `bser:"is_fresh_instance"`
	Root            string          `bser:"root"`
	Since           string          `bser:"since"`
	Subscription    string          `bser:"subscription"`
}

// SubscribeFile is a representation of the file that was somehow changed
type SubscribeFile struct {
	Mode   int    `bser:"mode"`
	New    bool   `bser:"new"`
	Size   int    `bser:"size"`
	Exists bool   `bser:"exists"`
	Name   string `bser:"name"`
}

// Subscribe subscribes to changes against a specified root and requests that they be sent to the client via its connection. The updates will continue to be sent while the connection is open. If the connection is closed, the subscription is implicitly removed
// https://facebook.github.io/watchman/docs/cmd/subscribe.html
// todo(isao) - add expression type?
func (c *Client) Subscribe(path, name string, expr map[string]interface{}, ch chan<- *SubscribeEvent) (*Subscribe, func(), error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, nil, err
	}

	var data struct {
		Subscribe
		base
	}

	all := make(chan interface{})
	stop, err := c.Receive(all)
	if err != nil {
		return nil, nil, err
	}

	if err := c.Send(&data, "subscribe", path, name, expr); err != nil {
		stop()
		return nil, nil, err
	}

	go func() {
		// all will get closed when stop is called
		// so we close ch when the func returns
		defer func() {
			close(ch)
		}()

		for m := range all {
			sub, _ := m.(*SubscribeEvent)
			if sub == nil || sub.Subscription != name {
				continue
			}

			ch <- sub
		}
	}()

	if data.Error != "" {
		stop()
		return nil, nil, data.Error
	}

	return &data.Subscribe, stop, nil
}
