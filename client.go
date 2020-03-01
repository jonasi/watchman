package watchman

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

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
	enc      *bser.Encoder
	dec      *bser.Decoder
	initOnce sync.Once
	initErr  error
	inited   int32
	reqCh    chan request
	cleanup  func() error
}

type request struct {
	dest  interface{}
	args  []interface{}
	errCh chan error
	uniCh chan bser.RawMessage
}

func (c *Client) init() error {
	c.initOnce.Do(func() {
		if !atomic.CompareAndSwapInt32(&c.inited, 0, 1) {
			c.initErr = errors.New("Cannot call send on a closed client")
			return
		}

		var err error
		if c.Sockname == "" {
			c.Sockname, err = inferSockname()
			if err != nil {
				c.initErr = err
				return
			}
		}

		sconn, err := initSock(c.Sockname)
		if err != nil {
			c.initErr = err
			return
		}

		c.cleanup = sconn.Close
		var conn io.ReadWriter = sconn

		if logPDU {
			tap := bser.NewTap(conn, pduLogger("incoming", os.Stderr), pduLogger("outgoing", os.Stderr))
			c.cleanup = func() error {
				tap.Untap()
				return sconn.Close()
			}

			conn = tap
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

// Close closes the connection to the watchman server
func (c *Client) Close() error {
	if !atomic.CompareAndSwapInt32(&c.inited, 1, 2) {
		// never been opened!
		return nil
	}

	close(c.reqCh)

	return c.cleanup()
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
		case req, ok := <-c.reqCh:
			if !ok {
				return
			}

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
