package watchman

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/jonasi/watchman/bser"
)

var logPDU = func() bool {
	return os.Getenv("WATCHMAN_LOG_PDU") != ""
}()

// Error is a Watchman API error
type Error string

func (e Error) Error() string {
	return string(e)
}

// Client is a watchman client
type Client struct {
	Sockname string
	conn     net.Conn
	enc      *bser.Encoder
	dec      *bser.Decoder
	initOnce sync.Once
	initErr  error
	reqCh    chan request
}

type request struct {
	dest  interface{}
	args  []interface{}
	errCh chan error
	uniCh chan bser.RawMessage
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

		var conn io.ReadWriter = c.conn
		if logPDU {
			conn = bser.Tap(c.conn, pduLogger("incoming", os.Stderr), pduLogger("outgoing", os.Stderr))
		}

		c.enc = bser.NewEncoder(conn)
		c.dec = bser.NewDecoder(conn)

		c.reqCh = make(chan request)
		decCh := make(chan interface{})

		go c.readPDUs(decCh)
		go c.handleReqs(decCh)
	})

	return c.initErr
}

func (c *Client) readPDUs(ch chan interface{}) {
	for {
		var m bser.RawMessage
		if err := c.dec.Decode(&m); err != nil {
			ch <- err
		} else {
			ch <- m
		}
	}
}

func (c *Client) handleReqs(ch chan interface{}) {
	var (
		activeReq  *request
		queuedReqs = []*request{}
		watches    = []chan bser.RawMessage{}
	)

	processNext := func() {
		if activeReq != nil || len(queuedReqs) == 0 {
			return
		}

		activeReq = queuedReqs[0]
		queuedReqs = queuedReqs[1:]

		if activeReq.uniCh != nil {
			watches = append(watches, activeReq.uniCh)
		}

		if err := c.enc.Encode(activeReq.args); err != nil {
			activeReq.errCh <- err
			activeReq = nil
		}
	}

	for {
		select {
		case req := <-c.reqCh:
			queuedReqs = append(queuedReqs, &req)
			processNext()
		case v := <-ch:
			if activeReq == nil {
				if msg, ok := v.(bser.RawMessage); ok {
					for _, w := range watches {
						w <- msg
					}
				}
				continue
			}

			if err, ok := v.(error); ok {
				activeReq.errCh <- err
			} else {
				activeReq.errCh <- bser.UnmarshalValue(v.(bser.RawMessage), activeReq.dest)
			}

			activeReq = nil
			processNext()
		}
	}
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

	r := request{args: args, dest: dest, errCh: make(chan error)}
	c.reqCh <- r
	return <-r.errCh
}

// SendAndWatch makes a client call and also listens for all unilateral messages
func (c *Client) SendAndWatch(ch chan bser.RawMessage, dest interface{}, args ...interface{}) error {
	if err := c.init(); err != nil {
		return err
	}

	r := request{args: args, dest: dest, errCh: make(chan error), uniCh: ch}
	c.reqCh <- r

	return <-r.errCh
}

type base struct {
	Version string `bser:"version"`
	Error   Error  `bser:"error"`
	Warning string `bser:"warning"`
}

func pduLogger(dir string, w io.Writer) func([]byte) {
	return func(b []byte) {
		var d interface{}
		if err := bser.UnmarshalValue(b, &d); err != nil {
			fmt.Fprintf(w, "[pdu logger - %s - bser unmarshal error]: %s\n", dir, err)
			return
		}
		b, err := json.Marshal(d)
		if err != nil {
			fmt.Fprintf(w, "[pdu logger - %s - json marshal error]: %s\n", dir, err)
			return
		}

		fmt.Fprintf(w, "[pdu logger - %s]: %s\n", dir, string(b))
	}
}
