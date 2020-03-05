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
	reqCh    chan interface{}
	cleanup  func() error
}

// request to send an outgoing message
type sendReq struct {
	dest  interface{}
	args  []interface{}
	errCh chan error
}

// request to listen for uniteral messages
type recReq struct {
	rec  chan<- interface{}
	stop chan struct{}
}

// request to stop listening for unilateral messages
type stopReq struct {
	idx int
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

		c.reqCh = make(chan interface{})
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

type watch struct {
	sync.RWMutex
	ch     chan<- interface{}
	closed bool
}

func (c *Client) handleReqs(ch chan interface{}) {
	var (
		activeReq  *sendReq
		queuedReqs = []*sendReq{}
		watches    = []*watch{}
	)

	processNext := func() {
		if activeReq != nil || len(queuedReqs) == 0 {
			return
		}

		activeReq = queuedReqs[0]
		queuedReqs = queuedReqs[1:]

		if err := c.enc.Encode(activeReq.args); err != nil {
			activeReq.errCh <- err
			activeReq = nil
		}
	}

	for {
		select {
		case v, ok := <-c.reqCh:
			if !ok {
				return
			}

			switch req := v.(type) {

			// send request - add it to the queue of messages
			// and immediately process
			case sendReq:
				queuedReqs = append(queuedReqs, &req)
				processNext()

			// rec request - add it to our list of watches
			// and start listen for the stop watching signal
			case recReq:
				watches = append(watches, &watch{ch: req.rec})
				go func(idx int, stop chan struct{}) {
					<-stop
					c.reqCh <- stopReq{idx}
				}(len(watches)-1, req.stop)

			// stop request - remove the watch channel
			// from the list of channels to watch and close the
			// channel
			case stopReq:
				w := watches[req.idx]

				w.Lock()
				close(w.ch)
				w.closed = true
				w.Unlock()

				watches[req.idx] = nil
			}
		case v := <-ch:
			if activeReq == nil {
				c.handleUnilateral(watches, v)
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

func (c *Client) handleUnilateral(watches []*watch, v interface{}) {
	// should only ever be this...
	msg, ok := v.(bser.RawMessage)
	if !ok {
		return
	}

	// todo(isao) - add SubscribeEvent here when we add that functionality
	var data struct {
		base
		*LogEvent
	}

	if err := bser.UnmarshalValue(msg, &data); err != nil {
		// todo(isao) - log?
		return
	}

	var d interface{}
	if data.LogEvent != nil {
		d = data.LogEvent
	}

	if d == nil {
		// unhandled
		// todo(isao) - log?
		return
	}

	// dispatch msg too all watchers asynchronously
	for _, w := range watches {
		if w == nil {
			continue
		}

		go func(w *watch, d interface{}) {
			w.RLock()
			defer w.RUnlock()
			if w.closed {
				return
			}

			w.ch <- d
		}(w, d)
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

	r := sendReq{args: args, dest: dest, errCh: make(chan error)}
	c.reqCh <- r

	return <-r.errCh
}

// Receive listens for unilateral messages from the server on ch.
func (c *Client) Receive(ch chan<- interface{}) (func(), error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	stop := make(chan struct{})
	c.reqCh <- recReq{rec: ch, stop: stop}
	return func() {
		stop <- struct{}{}
	}, nil
}

type base struct {
	Version    string `bser:"version"`
	Error      Error  `bser:"error"`
	Warning    string `bser:"warning"`
	Unilateral bool   `bser:"unilateral"`
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
