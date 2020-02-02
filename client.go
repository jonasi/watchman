package watchman

import (
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/jonasi/watchman/bser"
)

type enc interface {
	Encode(interface{}) error
}

type dec interface {
	Decode(interface{}) error
}

// Error is a Watchman API error
type Error string

func (e Error) Error() string {
	return string(e)
}

// Client is a watchman client
type Client struct {
	Sockname string
	JSON     bool
	conn     net.Conn
	enc      enc
	dec      dec
	encMu    sync.Mutex
	initOnce sync.Once
	initErr  error
}

func (c *Client) init() error {
	c.initOnce.Do(func() {
		var err error
		if c.Sockname == "" {
			c.Sockname, err = inferSockname()
			if err != nil {
				c.initErr = err
				return
			}
		}

		c.conn, err = initSock(c.Sockname)
		if err != nil {
			c.initErr = err
			return
		}

		if c.JSON {
			c.enc = json.NewEncoder(c.conn)
			c.dec = json.NewDecoder(c.conn)
		} else {
			c.enc = bser.NewEncoder(c.conn)
			c.dec = bser.NewDecoder(c.conn)
		}
	})

	return c.initErr
}

func initSock(sock string) (net.Conn, error) {
	addr, err := net.ResolveUnixAddr("unix", sock)
	if err != nil {
		return nil, err
	}

	return net.DialUnix("unix", nil, addr)
}

func inferSockname() (string, error) {
	if v := os.Getenv("WATCHMAN_SOCK"); v != "" {
		return v, nil
	}

	b, err := exec.Command("watchman", "get-sockname").Output()
	if err != nil {
		return "", err
	}

	var d map[string]string
	if err := json.Unmarshal(b, &d); err != nil {
		return "", err
	}

	return d["sockname"], nil
}

// Send makes a client call
func (c *Client) Send(dest interface{}, args ...interface{}) error {
	if err := c.init(); err != nil {
		return err
	}

	c.encMu.Lock()
	defer c.encMu.Unlock()

	if err := c.enc.Encode(args); err != nil {
		return err
	}

	return c.dec.Decode(dest)
}

type base struct {
	Version string `bser:"version"`
	Error   Error  `bser:"error"`
	Warning string `bser:"warning"`
}
